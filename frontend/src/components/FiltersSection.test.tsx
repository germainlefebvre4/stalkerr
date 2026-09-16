import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react';
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

    function testerControlsFor(card: HTMLElement) {
      const select = card.querySelector('select') as HTMLSelectElement;
      const button = Array.from(card.querySelectorAll('button')).find(b => b.textContent === 'Test') as HTMLButtonElement;
      return { select, button };
    }

    it("tests the origin card's own patterns without a save/delete request", async () => {
      vi.mocked(api.dryRunFilter).mockResolvedValue({
        no_archive: false, total_lines: 4, matched_count: 4, excluded_count: 0, top_matched: [], top_excluded: [],
      });
      const { onDeleteFilter } = renderSection([]);

      const card = screen.getByText('Group Title').closest('.filter-card') as HTMLElement;
      const { select, button } = testerControlsFor(card);
      fireEvent.change(select, { target: { value: 'main' } });
      fireEvent.click(button);

      await waitFor(() => expect(api.dryRunFilter).toHaveBeenCalledWith({
        source_name: 'main',
        attribute: 'group_title',
        include_patterns: 'FRENCH',
        exclude_patterns: '',
      }));
      expect(onDeleteFilter).not.toHaveBeenCalled();
    });

    it("tests the override card's own patterns, distinct from the origin's", async () => {
      vi.mocked(api.dryRunFilter).mockResolvedValue({
        no_archive: false, total_lines: 4, matched_count: 4, excluded_count: 0, top_matched: [], top_excluded: [],
      });
      const filters: FilterConfig[] = [
        { id: 1, name: 'My Override', attribute: 'group_title', include_patterns: 'OVERRIDE_INC', exclude_patterns: 'OVERRIDE_EXC', is_runtime: true },
      ];
      const { onDeleteFilter } = renderSection(filters);

      const overrideBadge = screen.getByText('Active override');
      const overrideBlock = overrideBadge.closest('div')!.parentElement as HTMLElement;
      const { select, button } = testerControlsFor(overrideBlock);
      fireEvent.change(select, { target: { value: 'main' } });
      fireEvent.click(button);

      await waitFor(() => expect(api.dryRunFilter).toHaveBeenCalledWith({
        source_name: 'main',
        attribute: 'group_title',
        include_patterns: 'OVERRIDE_INC',
        exclude_patterns: 'OVERRIDE_EXC',
      }));
      expect(onDeleteFilter).not.toHaveBeenCalled();
    });
  });
});
