import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
export interface MultiSelectDropdownOption<T extends string> {
  value: T;
  label: string;
}

interface MultiSelectDropdownProps<T extends string> {
  label: string;
  value: Set<T>;
  onChange: (value: Set<T>) => void;
  options: MultiSelectDropdownOption<T>[];
  noneLabel: string;
}

// A small trigger + checkbox-menu control for a multi-select filter, shared by
// the Suivi and Fichier filters of the Films and Séries sections. Selecting an
// option never closes the menu (onSelect is prevented), matching a standard
// multi-select checkbox group's behavior.
export function MultiSelectDropdown<T extends string>({ label, value, onChange, options, noneLabel }: MultiSelectDropdownProps<T>) {
  const toggle = (option: T, checked: boolean) => {
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
        <button type="button" className="custom-select multi-select-trigger">
          {triggerText}
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content className="multi-select-content" align="start" sideOffset={4}>
          {options.map(option => (
            <DropdownMenu.CheckboxItem
              key={option.value}
              className="multi-select-item"
              checked={value.has(option.value)}
              onCheckedChange={(checked) => toggle(option.value, checked === true)}
              onSelect={(e) => e.preventDefault()}
            >
              <DropdownMenu.ItemIndicator className="multi-select-item-check">✓</DropdownMenu.ItemIndicator>
              {option.label}
            </DropdownMenu.CheckboxItem>
          ))}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
