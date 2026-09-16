import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent, within } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { FiltersSection } from './FiltersSection';
import { FilterConfig, FilterOriginEntry, M3uSource } from '../types';
import { api } from '../services/api';

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

const filterOrigin: FilterOriginEntry[] = [
  { attribute: 'group_title', include_patterns: ['FRENCH'], exclude_patterns: [] },
  { attribute: 'tvg_name', include_patterns: [], exclude_patterns: [] },
];

const sources: M3uSource[] = [{ name: 'main', is_runtime: false } as M3uSource];

function renderSection(filters: FilterConfig[], searchQuery?: string) {
  const onDeleteFilter = vi.fn();
  const onOpenCreate = vi.fn();
  render(
    <I18nextProvider i18n={i18n}>
      <FiltersSection
        isExpanded
        filters={filters}
        filterOrigin={filterOrigin}
        filtersLoading={false}
        onDeleteFilter={onDeleteFilter}
        onOpenCreate={onOpenCreate}
        sources={sources}
        searchQuery={searchQuery}
      />
    </I18nextProvider>
  );
  return { onDeleteFilter, onOpenCreate };
}

describe('FiltersSection', () => {
  afterEach(() => cleanup());

  it('shows only the origin patterns for an attribute with no active override', () => {
    renderSection([]);

    expect(screen.getByText('Group Title')).toBeInTheDocument();
    expect(screen.getByText('FRENCH')).toBeInTheDocument();
    expect(screen.getAllByText('Origin (config.yml)').length).toBe(2);
    expect(screen.queryByText('Active override')).not.toBeInTheDocument();
  });

  it('shows the active override alongside the origin patterns, visually distinguished', () => {
    const filters: FilterConfig[] = [
      { id: 1, name: 'My Override', attribute: 'group_title', include_patterns: 'OVERRIDE', is_runtime: true },
    ];
    renderSection(filters);

    expect(screen.getByText('Active override')).toBeInTheDocument();
    expect(screen.getByText('My Override')).toBeInTheDocument();
    expect(screen.getByText('OVERRIDE')).toBeInTheDocument();
    // The origin patterns for that same attribute remain visible too.
    expect(screen.getByText('FRENCH')).toBeInTheDocument();
  });

  it('deletes the active override via its delete button', () => {
    const filters: FilterConfig[] = [
      { id: 1, name: 'My Override', attribute: 'group_title', is_runtime: true },
    ];
    const { onDeleteFilter } = renderSection(filters);

    fireEvent.click(screen.getByTitle('Delete this filter'));
    expect(onDeleteFilter).toHaveBeenCalledWith(1);
  });

  it('opens the create dialog from the section header trigger', () => {
    const { onOpenCreate } = renderSection([]);
    fireEvent.click(screen.getByText('Configure a Filter'));
    expect(onOpenCreate).toHaveBeenCalled();
  });

  it('renders nothing when collapsed', () => {
    const { container } = render(
      <I18nextProvider i18n={i18n}>
        <FiltersSection
          isExpanded={false}
          filters={[]}
          filterOrigin={filterOrigin}
          filtersLoading={false}
          onDeleteFilter={vi.fn()}
          onOpenCreate={vi.fn()}
          sources={sources}
        />
      </I18nextProvider>
    );
    expect(container).toBeEmptyDOMElement();
  });

  describe('dry-run testing', () => {
    afterEach(() => vi.mocked(api.dryRunFilter).mockReset());

    function testButtonFor(card: HTMLElement) {
      return Array.from(card.querySelectorAll('button')).find(b => b.textContent === 'Test') as HTMLButtonElement;
    }

    it("opens the shared drawer with the origin card's own patterns, without a save/delete request", async () => {
      const { onDeleteFilter } = renderSection([]);

      const card = screen.getByText('Group Title').closest('.filter-card') as HTMLElement;
      fireEvent.click(testButtonFor(card));

      expect(await screen.findByText('Test: Group Title — Origin (config.yml)')).toBeInTheDocument();
      expect(onDeleteFilter).not.toHaveBeenCalled();
    });

    it("opens the shared drawer with the override card's own patterns, distinct from the origin's", async () => {
      const filters: FilterConfig[] = [
        { id: 1, name: 'My Override', attribute: 'group_title', include_patterns: 'OVERRIDE_INC', exclude_patterns: 'OVERRIDE_EXC', is_runtime: true },
      ];
      const { onDeleteFilter } = renderSection(filters);

      const overrideBadge = screen.getByText('Active override');
      const overrideBlock = overrideBadge.closest('div')!.parentElement as HTMLElement;
      fireEvent.click(testButtonFor(overrideBlock));

      expect(await screen.findByText('Test: Group Title — Active override')).toBeInTheDocument();
      expect(onDeleteFilter).not.toHaveBeenCalled();
    });

    it('reuses the same drawer when testing a different card', async () => {
      const filters: FilterConfig[] = [
        { id: 1, name: 'My Override', attribute: 'group_title', include_patterns: 'OVERRIDE_INC', is_runtime: true },
      ];
      renderSection(filters);

      const groupCard = screen.getByText('Group Title').closest('.filter-card') as HTMLElement;
      fireEvent.click(testButtonFor(groupCard));
      expect(await screen.findByText('Test: Group Title — Origin (config.yml)')).toBeInTheDocument();

      const tvgCard = screen.getByText('TVG Name').closest('.filter-card') as HTMLElement;
      fireEvent.click(testButtonFor(tvgCard));

      expect(await screen.findByText('Test: TVG Name — Origin (config.yml)')).toBeInTheDocument();
      expect(screen.queryByText('Test: Group Title — Origin (config.yml)')).not.toBeInTheDocument();
      // Still exactly one drawer/dialog, not a second one stacked on top.
      expect(screen.getAllByRole('dialog')).toHaveLength(1);
    });

    it('opens the combined "test everything" drawer using the effective patterns (override if present, else origin)', async () => {
      vi.mocked(api.dryRunFilter).mockResolvedValue({
        no_archive: false, total_lines: 4, kept_count: 4,
        excluded_by_group_title_only: 0, excluded_by_tvg_name_only: 0, excluded_by_both: 0,
        group_title_top_matched: [], group_title_top_excluded: [],
        tvg_name_top_matched: [], tvg_name_top_excluded: [],
      });
      const filters: FilterConfig[] = [
        { id: 1, name: 'My Override', attribute: 'group_title', include_patterns: 'OVERRIDE_INC', exclude_patterns: 'OVERRIDE_EXC', is_runtime: true },
      ];
      renderSection(filters);

      fireEvent.click(screen.getByText('Test everything'));

      const dialog = await screen.findByRole('dialog');
      expect(within(dialog).getByText('Test everything')).toBeInTheDocument();

      fireEvent.change(within(dialog).getByRole('combobox'), { target: { value: 'main' } });
      fireEvent.click(within(dialog).getByText('Test'));

      // group_title uses the active override, tvg_name falls back to origin
      // (empty patterns since no origin entry has any in this fixture).
      expect(api.dryRunFilter).toHaveBeenCalledWith({
        source_name: 'main',
        attributes: ['group_title', 'tvg_name'],
        group_title_include_patterns: 'OVERRIDE_INC',
        group_title_exclude_patterns: 'OVERRIDE_EXC',
        tvg_name_include_patterns: '',
        tvg_name_exclude_patterns: '',
      });
    });
  });
});
