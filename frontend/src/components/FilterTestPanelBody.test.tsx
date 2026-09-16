import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { FilterTestPanelBody } from './FilterTestPanelBody';
import { api } from '../services/api';
import { FilterTestTarget, M3uSource } from '../types';

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

const sources: M3uSource[] = [{ name: 'main', is_runtime: false } as M3uSource];

function selectSource(name = 'main') {
  fireEvent.change(screen.getAllByRole('combobox')[0], { target: { value: name } });
}

function renderBody(target: FilterTestTarget) {
  return render(
    <I18nextProvider i18n={i18n}>
      <FilterTestPanelBody target={target} sources={sources} />
    </I18nextProvider>
  );
}

const singleTarget: FilterTestTarget = {
  mode: 'single',
  attribute: 'group_title',
  includePatterns: 'FRENCH',
  excludePatterns: '',
  label: 'Test: Group Title — in progress',
};

const combinedTarget: FilterTestTarget = {
  mode: 'combined',
  groupTitleInclude: '',
  groupTitleExclude: 'XXX',
  tvgNameInclude: '',
  tvgNameExclude: 'Bad',
  label: 'Test everything',
};

describe('FilterTestPanelBody', () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('lists the M3U sources in the source select', () => {
    renderBody(singleTarget);
    expect(screen.getByText('main')).toBeInTheDocument();
  });

  it('runs a single-attribute test and renders the summary plus top values grid', async () => {
    vi.mocked(api.dryRunFilter).mockResolvedValue({
      no_archive: false,
      total_lines: 10,
      matched_count: 7,
      excluded_count: 3,
      top_matched: [{ value: 'Movies HD', count: 5 }],
      top_excluded: [{ value: 'Adult XXX', count: 3 }],
    });
    renderBody(singleTarget);

    selectSource();
    fireEvent.click(screen.getByText('Test'));

    await waitFor(() => expect(api.dryRunFilter).toHaveBeenCalledWith({
      source_name: 'main',
      attributes: ['group_title'],
      group_title_include_patterns: 'FRENCH',
      group_title_exclude_patterns: '',
    }));

    expect(await screen.findByText('10 lines scanned')).toBeInTheDocument();
    expect(screen.getByText('7 would match')).toBeInTheDocument();
    expect(screen.getByText('3 would be excluded')).toBeInTheDocument();
    expect(screen.getByText('Movies HD (5)')).toBeInTheDocument();
    expect(screen.getByText('Adult XXX (3)')).toBeInTheDocument();
  });

  it('searches and renders would-match/would-be-excluded labels once a summary is shown', async () => {
    vi.mocked(api.dryRunFilter).mockResolvedValueOnce({
      no_archive: false, total_lines: 5, matched_count: 3, excluded_count: 2, top_matched: [], top_excluded: [],
    }).mockResolvedValueOnce({
      no_archive: false,
      truncated: false,
      results: [
        { group_title: 'Movies HD', tvg_name: 'A', matched: true },
        { group_title: 'Adult XXX', tvg_name: 'B', matched: false },
      ],
    });
    renderBody(singleTarget);

    selectSource();
    fireEvent.click(screen.getByText('Test'));
    await screen.findByText('5 lines scanned');

    fireEvent.change(screen.getByPlaceholderText('E.g.: a channel or group name'), { target: { value: 'Movies' } });

    await waitFor(() => expect(api.dryRunFilter).toHaveBeenLastCalledWith({
      source_name: 'main',
      attributes: ['group_title'],
      group_title_include_patterns: 'FRENCH',
      group_title_exclude_patterns: '',
      search: 'Movies',
      search_attribute: 'group_title',
    }));

    expect(await screen.findByText('Would match')).toBeInTheDocument();
    expect(await screen.findByText('Would be excluded')).toBeInTheDocument();
  });

  it('shows the no-archive message', async () => {
    vi.mocked(api.dryRunFilter).mockResolvedValue({
      no_archive: true, total_lines: 0, matched_count: 0, excluded_count: 0, top_matched: [], top_excluded: [],
    });
    renderBody(singleTarget);

    selectSource();
    fireEvent.click(screen.getByText('Test'));

    expect(await screen.findByText(/No archive has been downloaded yet/)).toBeInTheDocument();
  });

  it('runs a combined test and renders the four cause counts plus each attribute block', async () => {
    vi.mocked(api.dryRunFilter).mockResolvedValue({
      no_archive: false,
      total_lines: 4,
      kept_count: 1,
      excluded_by_group_title_only: 1,
      excluded_by_tvg_name_only: 1,
      excluded_by_both: 1,
      group_title_top_matched: [{ value: 'Movies HD', count: 2 }],
      group_title_top_excluded: [],
      tvg_name_top_matched: [],
      tvg_name_top_excluded: [{ value: 'Bad Channel', count: 2 }],
    });
    renderBody(combinedTarget);

    selectSource();
    fireEvent.click(screen.getByText('Test'));

    await waitFor(() => expect(api.dryRunFilter).toHaveBeenCalledWith({
      source_name: 'main',
      attributes: ['group_title', 'tvg_name'],
      group_title_include_patterns: '',
      group_title_exclude_patterns: 'XXX',
      tvg_name_include_patterns: '',
      tvg_name_exclude_patterns: 'Bad',
    }));

    expect(await screen.findByText('1 kept')).toBeInTheDocument();
    expect(screen.getByText('1 excluded by Group Title only')).toBeInTheDocument();
    expect(screen.getByText('1 excluded by TVG Name only')).toBeInTheDocument();
    expect(screen.getByText('1 excluded by both')).toBeInTheDocument();
    expect(screen.getByText('Movies HD (2)')).toBeInTheDocument();
    expect(screen.getByText('Bad Channel (2)')).toBeInTheDocument();
  });

  it('lets the user pick the search attribute in combined mode and shows verdict labels', async () => {
    vi.mocked(api.dryRunFilter).mockResolvedValueOnce({
      no_archive: false, total_lines: 2, kept_count: 1,
      excluded_by_group_title_only: 0, excluded_by_tvg_name_only: 1, excluded_by_both: 0,
      group_title_top_matched: [], group_title_top_excluded: [],
      tvg_name_top_matched: [], tvg_name_top_excluded: [],
    }).mockResolvedValueOnce({
      no_archive: false,
      truncated: false,
      results: [
        { group_title: 'Movies HD', tvg_name: 'Good Channel', matched: false, verdict: 'kept' },
        { group_title: 'Movies HD', tvg_name: 'Bad Channel', matched: false, verdict: 'excluded_by_tvg_name' },
      ],
    });
    renderBody(combinedTarget);

    selectSource();
    fireEvent.click(screen.getByText('Test'));
    await screen.findByText('1 kept');

    fireEvent.change(screen.getByPlaceholderText('E.g.: a channel or group name'), { target: { value: 'Channel' } });

    await waitFor(() => expect(api.dryRunFilter).toHaveBeenLastCalledWith({
      source_name: 'main',
      attributes: ['group_title', 'tvg_name'],
      group_title_include_patterns: '',
      group_title_exclude_patterns: 'XXX',
      tvg_name_include_patterns: '',
      tvg_name_exclude_patterns: 'Bad',
      search: 'Channel',
      search_attribute: 'group_title',
    }));

    expect(await screen.findByText('Kept')).toBeInTheDocument();
    expect(screen.getByText('Excluded by TVG Name')).toBeInTheDocument();
  });

  it('resets the previous test result when the target changes', async () => {
    vi.mocked(api.dryRunFilter).mockResolvedValue({
      no_archive: false, total_lines: 5, matched_count: 3, excluded_count: 2, top_matched: [], top_excluded: [],
    });
    const { rerender } = renderBody(singleTarget);

    selectSource();
    fireEvent.click(screen.getByText('Test'));
    await screen.findByText('5 lines scanned');

    rerender(
      <I18nextProvider i18n={i18n}>
        <FilterTestPanelBody target={{ ...singleTarget, includePatterns: 'OTHER' }} sources={sources} />
      </I18nextProvider>
    );

    expect(screen.queryByText('5 lines scanned')).not.toBeInTheDocument();
  });
});
