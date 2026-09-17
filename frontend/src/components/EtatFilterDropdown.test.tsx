import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import { EtatFilterDropdown } from './EtatFilterDropdown';
import { EtatFilter } from '../types';

afterEach(() => {
  cleanup();
});

const OPTIONS = [
  { value: 'monitored' as const, label: 'Monitored' },
  { value: 'unmonitored' as const, label: 'Unmonitored' },
  { value: 'missing' as const, label: 'Missing' },
];

function openMenu(trigger: HTMLElement) {
  fireEvent.keyDown(trigger, { key: 'Enter' });
}

describe('EtatFilterDropdown', () => {
  it('reflects the current selection in the trigger label', () => {
    const value: EtatFilter = new Set(['monitored', 'missing']);
    render(<EtatFilterDropdown label="État" value={value} onChange={() => {}} options={OPTIONS} noneLabel="None" />);

    expect(screen.getByText('État: Monitored, Missing')).toBeInTheDocument();
  });

  it('shows the "none selected" label when nothing is selected', () => {
    const value: EtatFilter = new Set();
    render(<EtatFilterDropdown label="État" value={value} onChange={() => {}} options={OPTIONS} noneLabel="None" />);

    expect(screen.getByText('État: None')).toBeInTheDocument();
  });

  // Radix's DropdownMenu mounts a real portal/focus-scope/dismissable-layer
  // stack, which is measurably slower under CI/parallel-worker load than a
  // plain render - bump the timeout so these two don't flake under load
  // (they otherwise pass in well under a second locally).
  it('toggling a checkbox reports the updated selection without closing the menu', () => {
    const value: EtatFilter = new Set(['monitored']);
    const onChange = vi.fn();
    render(<EtatFilterDropdown label="État" value={value} onChange={onChange} options={OPTIONS} noneLabel="None" />);

    openMenu(screen.getByText('État: Monitored'));

    const missingItem = screen.getByText('Missing');
    fireEvent.click(missingItem);

    expect(onChange).toHaveBeenCalledWith(new Set(['monitored', 'missing']));
    // The menu content is still present/queryable right after the click,
    // i.e. selecting an option did not close the menu.
    expect(screen.getByText('Unmonitored')).toBeInTheDocument();
  }, 20000);

  it('unchecking a selected option removes it from the reported selection', () => {
    const value: EtatFilter = new Set(['monitored', 'missing']);
    const onChange = vi.fn();
    render(<EtatFilterDropdown label="État" value={value} onChange={onChange} options={OPTIONS} noneLabel="None" />);

    openMenu(screen.getByText('État: Monitored, Missing'));

    fireEvent.click(screen.getByText('Missing'));

    expect(onChange).toHaveBeenCalledWith(new Set(['monitored']));
  }, 20000);
});
