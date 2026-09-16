import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { FilterDryRunPanel } from './FilterDryRunPanel';
import { api } from '../services/api';
import { FilterDryRunSummaryResponse } from '../types';

vi.mock('../services/api', () => ({
  api: {
    dryRunFilter: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    code: string;
    constructor(code: string, message: string) {
      super(message);
      this.code = code;
    }
  },
}));

const baseSummary: FilterDryRunSummaryResponse = {
  no_archive: false,
  total_lines: 10,
  matched_count: 7,
  excluded_count: 3,
  top_matched: [{ value: 'Movies HD', count: 5 }],
  top_excluded: [{ value: 'Adult XXX', count: 3 }],
};

function renderPanel(props: Partial<React.ComponentProps<typeof FilterDryRunPanel>> = {}) {
  return render(
    <I18nextProvider i18n={i18n}>
      <FilterDryRunPanel
        status="idle"
        summary={null}
        errorMessage={null}
        sourceName="main"
        attribute="group_title"
        includePatterns=""
        excludePatterns=""
        {...props}
      />
    </I18nextProvider>
  );
}

describe('FilterDryRunPanel', () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('renders nothing and hides the search field in the idle state', () => {
    const { container } = renderPanel({ status: 'idle' });
    expect(container).toBeEmptyDOMElement();
    expect(screen.queryByPlaceholderText('E.g.: a channel or group name')).not.toBeInTheDocument();
  });

  it('renders the summary counts and top matched/excluded values side by side in a grid', () => {
    const { container } = renderPanel({ status: 'summary', summary: baseSummary });

    expect(screen.getByText('10 lines scanned')).toBeInTheDocument();
    expect(screen.getByText('7 would match')).toBeInTheDocument();
    expect(screen.getByText('3 would be excluded')).toBeInTheDocument();
    expect(screen.getByText('Movies HD (5)')).toBeInTheDocument();
    expect(screen.getByText('Adult XXX (3)')).toBeInTheDocument();

    const columns = container.querySelector('.dry-run-columns');
    expect(columns).not.toBeNull();
    expect(columns).toContainElement(screen.getByText('Movies HD (5)'));
    expect(columns).toContainElement(screen.getByText('Adult XXX (3)'));
    // Both lists are direct children of the same grid container, not nested
    // one inside the other in a single stacked column.
    expect(columns?.children).toHaveLength(2);
  });

  it('shows the no-archive message', () => {
    renderPanel({ status: 'no_archive' });
    expect(screen.getByText(/No archive has been downloaded yet/)).toBeInTheDocument();
  });

  it('shows the inline error for an invalid pattern', () => {
    renderPanel({ status: 'error', errorMessage: 'Invalid pattern: ^(Movies' });
    expect(screen.getByText('Invalid pattern: ^(Movies')).toBeInTheDocument();
  });

  it('searches and renders would-match/would-be-excluded labels once a summary is shown', async () => {
    vi.mocked(api.dryRunFilter).mockResolvedValue({
      no_archive: false,
      truncated: false,
      results: [
        { group_title: 'Movies HD', tvg_name: 'A', matched: true },
        { group_title: 'Adult XXX', tvg_name: 'B', matched: false },
      ],
    });

    renderPanel({ status: 'summary', summary: baseSummary });

    const searchInput = screen.getByPlaceholderText('E.g.: a channel or group name');
    fireEvent.change(searchInput, { target: { value: 'Movies' } });

    await waitFor(() => expect(api.dryRunFilter).toHaveBeenCalledWith({
      source_name: 'main',
      attribute: 'group_title',
      include_patterns: '',
      exclude_patterns: '',
      search: 'Movies',
    }));

    expect(await screen.findByText('Would match')).toBeInTheDocument();
    expect(await screen.findByText('Would be excluded')).toBeInTheDocument();
    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  it('renders the truncated message when the search result is truncated', async () => {
    vi.mocked(api.dryRunFilter).mockResolvedValue({
      no_archive: false,
      truncated: true,
      results: [
        { group_title: 'Movies HD', tvg_name: 'A', matched: true },
      ],
    });

    renderPanel({ status: 'summary', summary: baseSummary });

    const searchInput = screen.getByPlaceholderText('E.g.: a channel or group name');
    fireEvent.change(searchInput, { target: { value: 'Movies' } });

    expect(await screen.findByRole('table')).toBeInTheDocument();
    expect(screen.getByText(/Showing the first 100 matches/)).toBeInTheDocument();
  });
});
