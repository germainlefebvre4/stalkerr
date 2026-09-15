import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { CreateFilterDialog } from './CreateFilterDialog';
import { api } from '../services/api';
import { FilterConfig, FilterOriginEntry } from '../types';

vi.mock('../services/api', () => ({
  api: {
    createFilter: vi.fn(),
    getM3uSources: vi.fn().mockResolvedValue({ sources: [{ name: 'main', is_runtime: false }, { name: 'backup', is_runtime: false }] }),
    getM3uSourcesOrigin: vi.fn().mockResolvedValue({ sources: [] }),
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

const filterOrigin: FilterOriginEntry[] = [
  { attribute: 'group_title', include_patterns: ['FRENCH', 'VFF'], exclude_patterns: ['VOSTFR'] },
  { attribute: 'tvg_name', include_patterns: [], exclude_patterns: [] },
];

function renderDialog(filters: FilterConfig[]) {
  const onSuccess = vi.fn();
  render(
    <I18nextProvider i18n={i18n}>
      <CreateFilterDialog isOpen onOpenChange={vi.fn()} onSuccess={onSuccess} filters={filters} filterOrigin={filterOrigin} />
    </I18nextProvider>
  );
  return { onSuccess };
}

describe('CreateFilterDialog', () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('loads the origin configuration into the form when there is no active override', () => {
    renderDialog([]);

    fireEvent.click(screen.getByText('Load current configuration'));

    expect(screen.getByPlaceholderText('E.g.: FRENCH, TRUEFRENCH, VFF')).toHaveValue('FRENCH, VFF');
    expect(screen.getByPlaceholderText('E.g.: VOSTFR, SUBBED, HDLight')).toHaveValue('VOSTFR');
  });

  it('loads the active override into the form when one exists for the selected attribute', () => {
    const filters: FilterConfig[] = [
      { id: 1, name: 'My Override', attribute: 'group_title', include_patterns: 'OVERRIDE_INC', exclude_patterns: 'OVERRIDE_EXC', is_runtime: true },
    ];
    renderDialog(filters);

    fireEvent.click(screen.getByText('Load current configuration'));

    expect(screen.getByPlaceholderText('E.g.: FRENCH, TRUEFRENCH, VFF')).toHaveValue('OVERRIDE_INC');
    expect(screen.getByPlaceholderText('E.g.: VOSTFR, SUBBED, HDLight')).toHaveValue('OVERRIDE_EXC');
  });

  it('warns inline (not via a native confirm) when the selected attribute already has an active override', () => {
    const filters: FilterConfig[] = [
      { id: 1, name: 'My Override', attribute: 'group_title', is_runtime: true },
    ];
    renderDialog(filters);

    expect(screen.getByRole('alert')).toHaveTextContent('My Override');
  });

  it('shows no warning when the selected attribute has no active override', () => {
    renderDialog([]);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('submits the create request and reports success', async () => {
    vi.mocked(api.createFilter).mockResolvedValue({} as never);
    const { onSuccess } = renderDialog([]);

    fireEvent.change(screen.getByPlaceholderText('E.g.: French Movies Filter'), { target: { value: 'New Filter' } });
    fireEvent.click(screen.getByText('✓ Save Filter'));

    await waitFor(() => expect(api.createFilter).toHaveBeenCalled());
    expect(onSuccess).toHaveBeenCalled();
  });

  describe('dry-run testing', () => {
    async function selectTestSource(name = 'main') {
      const option = await screen.findByText('main') as HTMLOptionElement;
      const select = option.closest('select') as HTMLSelectElement;
      fireEvent.change(select, { target: { value: name } });
    }

    it('lists the M3U sources in the test source select', async () => {
      renderDialog([]);
      await screen.findByText('main');
      expect(screen.getByText('backup')).toBeInTheDocument();
    });

    it('tests the in-progress patterns without submitting the create request', async () => {
      vi.mocked(api.createFilter).mockClear();
      vi.mocked(api.dryRunFilter).mockResolvedValue({
        no_archive: false, total_lines: 5, matched_count: 3, excluded_count: 2, top_matched: [], top_excluded: [],
      });
      renderDialog([]);

      await selectTestSource();
      fireEvent.change(screen.getByPlaceholderText('E.g.: FRENCH, TRUEFRENCH, VFF'), { target: { value: 'FRENCH' } });
      fireEvent.click(screen.getByText('Test'));

      await waitFor(() => expect(api.dryRunFilter).toHaveBeenCalledWith({
        source_name: 'main',
        attribute: 'group_title',
        include_patterns: 'FRENCH',
        exclude_patterns: '',
      }));
      expect(api.createFilter).not.toHaveBeenCalled();
      expect(await screen.findByText('5 lines scanned')).toBeInTheDocument();
    });

    it('resets a previous test result when the attribute changes', async () => {
      vi.mocked(api.dryRunFilter).mockResolvedValue({
        no_archive: false, total_lines: 5, matched_count: 3, excluded_count: 2, top_matched: [], top_excluded: [],
      });
      renderDialog([]);

      await selectTestSource();
      fireEvent.click(screen.getByText('Test'));
      await screen.findByText('5 lines scanned');

      fireEvent.change(screen.getByDisplayValue('Group Title (E.g.: VOD-FR, SERIES-US)'), { target: { value: 'tvg_name' } });

      expect(screen.queryByText('5 lines scanned')).not.toBeInTheDocument();
    });

    it('resets a previous test result when the include/exclude patterns are edited', async () => {
      vi.mocked(api.dryRunFilter).mockResolvedValue({
        no_archive: false, total_lines: 5, matched_count: 3, excluded_count: 2, top_matched: [], top_excluded: [],
      });
      renderDialog([]);

      await selectTestSource();
      fireEvent.click(screen.getByText('Test'));
      await screen.findByText('5 lines scanned');

      fireEvent.change(screen.getByPlaceholderText('E.g.: FRENCH, TRUEFRENCH, VFF'), { target: { value: 'FRENCH' } });

      expect(screen.queryByText('5 lines scanned')).not.toBeInTheDocument();
    });

    it('resets a previous test result when the test source changes', async () => {
      vi.mocked(api.dryRunFilter).mockResolvedValue({
        no_archive: false, total_lines: 5, matched_count: 3, excluded_count: 2, top_matched: [], top_excluded: [],
      });
      renderDialog([]);

      await selectTestSource('main');
      fireEvent.click(screen.getByText('Test'));
      await screen.findByText('5 lines scanned');

      await selectTestSource('backup');

      expect(screen.queryByText('5 lines scanned')).not.toBeInTheDocument();
    });
  });
});
