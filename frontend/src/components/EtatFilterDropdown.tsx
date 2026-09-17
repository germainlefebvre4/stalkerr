import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import { EtatFilter, EtatStatus } from '../types';

export interface EtatFilterDropdownOption {
  value: EtatStatus;
  label: string;
}

interface EtatFilterDropdownProps {
  label: string;
  value: EtatFilter;
  onChange: (value: EtatFilter) => void;
  options: EtatFilterDropdownOption[];
  noneLabel: string;
}

// A small trigger + checkbox-menu control for the État (Monitored/Unmonitored/
// Missing) cumulative multi-select filter, shared by the Films and Séries
// sections. Selecting an option never closes the menu (onSelect is
// prevented), matching a standard multi-select checkbox group's behavior.
export function EtatFilterDropdown({ label, value, onChange, options, noneLabel }: EtatFilterDropdownProps) {
  const toggle = (option: EtatStatus, checked: boolean) => {
    const next = new Set(value);
    if (checked) next.add(option);
    else next.delete(option);
    onChange(next);
  };

  const selectedLabels = options.filter(o => value.has(o.value)).map(o => o.label);
  const triggerText = `${label}: ${selectedLabels.length > 0 ? selectedLabels.join(', ') : noneLabel}`;

  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <button type="button" className="custom-select etat-filter-trigger">
          {triggerText}
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content className="etat-filter-content" align="start" sideOffset={4}>
          {options.map(option => (
            <DropdownMenu.CheckboxItem
              key={option.value}
              className="etat-filter-item"
              checked={value.has(option.value)}
              onCheckedChange={(checked) => toggle(option.value, checked === true)}
              onSelect={(e) => e.preventDefault()}
            >
              <DropdownMenu.ItemIndicator className="etat-filter-item-check">✓</DropdownMenu.ItemIndicator>
              {option.label}
            </DropdownMenu.CheckboxItem>
          ))}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
