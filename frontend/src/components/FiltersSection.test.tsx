import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { FiltersSection } from './FiltersSection';
import { FilterConfig, FilterOriginEntry } from '../types';

const filterOrigin: FilterOriginEntry[] = [
  { attribute: 'group_title', include_patterns: ['FRENCH'], exclude_patterns: [] },
  { attribute: 'tvg_name', include_patterns: [], exclude_patterns: [] },
];

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
        />
      </I18nextProvider>
    );
    expect(container).toBeEmptyDOMElement();
  });
});
