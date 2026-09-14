export interface ToggleSwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  iconOff?: string;
  iconOn?: string;
  ariaLabel: string;
  title?: string;
  disabled?: boolean;
}

// Shared toggle-switch control, used for boolean preferences (Apparence) and
// boolean overridable settings fields (see the Boolean Field Control
// requirement in frontend-app-settings-management).
export function ToggleSwitch({ checked, onChange, iconOff, iconOn, ariaLabel, title, disabled }: ToggleSwitchProps) {
  return (
    <label className="view-toggle-switch" title={title} aria-label={ariaLabel}>
      {iconOff && (
        <span className={`view-toggle-switch-icon${!checked ? ' is-active' : ''}`} aria-hidden="true">{iconOff}</span>
      )}
      <input
        type="checkbox"
        className="view-toggle-switch-input"
        checked={checked}
        disabled={disabled}
        onChange={e => onChange(e.target.checked)}
      />
      <span className="view-toggle-switch-track">
        <span className="view-toggle-switch-knob" />
      </span>
      {iconOn && (
        <span className={`view-toggle-switch-icon${checked ? ' is-active' : ''}`} aria-hidden="true">{iconOn}</span>
      )}
    </label>
  );
}
