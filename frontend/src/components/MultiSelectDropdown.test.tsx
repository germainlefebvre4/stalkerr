import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import { MultiSelectDropdown } from './MultiSelectDropdown';

type Suivi = 'monitored' | 'unmonitored' | 'extra';

afterEach(() => {
  cleanup();
});

const OPTIONS = [
  { value: 'monitored' as const, label: 'Monitored' },
  { value: 'unmonitored' as const, label: 'Unmonitored' },
  { value: 'extra' as const, label: 'Extra' },
];

function openMenu(trigger: HTMLElement) {
  fireEvent.keyDown(trigger, { key: 'Enter' });
}

describe('MultiSelectDropdown', () => {
  it('reflects the current selection in the trigger label', () => {
    const value: Set<Suivi> = new Set(['monitored', 'extra']);
    render(<MultiSelectDropdown label="Filter" value={value} onChange={() => {}} options={OPTIONS} noneLabel="None" />);

    expect(screen.getByText('Filter: Monitored, Extra')).toBeInTheDocument();
  });

  it('shows the "none selected" label when nothing is selected', () => {
    const value: Set<Suivi> = new Set();
    render(<MultiSelectDropdown label="Filter" value={value} onChange={() => {}} options={OPTIONS} noneLabel="None" />);

    expect(screen.getByText('Filter: None')).toBeInTheDocument();
  });

  // Radix's DropdownMenu mounts a real portal/focus-scope/dismissable-layer
  // stack, which is measurably slower under CI/parallel-worker load than a
  // plain render - bump the timeout so these two don't flake under load
  // (they otherwise pass in well under a second locally).
  it('toggling a checkbox reports the updated selection without closing the menu', () => {
    const value: Set<Suivi> = new Set(['monitored']);
    const onChange = vi.fn();
    render(<MultiSelectDropdown label="Filter" value={value} onChange={onChange} options={OPTIONS} noneLabel="None" />);

    openMenu(screen.getByText('Filter: Monitored'));

    const missingItem = screen.getByText('Extra');
    fireEvent.click(missingItem);

    expect(onChange).toHaveBeenCalledWith(new Set(['monitored', 'extra']));
    // The menu content is still present/queryable right after the click,
    // i.e. selecting an option did not close the menu.
    expect(screen.getByText('Unmonitored')).toBeInTheDocument();
  }, 20000);

  it('unchecking a selected option removes it from the reported selection', () => {
    const value: Set<Suivi> = new Set(['monitored', 'extra']);
    const onChange = vi.fn();
    render(<MultiSelectDropdown label="Filter" value={value} onChange={onChange} options={OPTIONS} noneLabel="None" />);

    openMenu(screen.getByText('Filter: Monitored, Extra'));

    fireEvent.click(screen.getByText('Extra'));

    expect(onChange).toHaveBeenCalledWith(new Set(['monitored']));
  }, 20000);
});
