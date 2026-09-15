import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import * as Tooltip from '@radix-ui/react-tooltip';
import { RotateCcw } from 'lucide-react';
import { SettingsField } from '../types';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';
import { ToggleSwitch } from './ToggleSwitch';

interface SettingsFieldRowProps {
  label: string;
  field: SettingsField;
  type?: 'text' | 'number' | 'boolean' | 'select';
  /** Used only when `type === 'select'`: the fixed set of technical values, each with a localized label key. */
  options?: { value: string; labelKey: string }[];
  /** Controlled draft value (raw string, `'true'`/`'false'` for booleans). Defaults to the field's effective value. */
  value?: string;
  /** Whether this field currently has an unsaved staged change (owned by the enclosing group card). */
  pending?: boolean;
  onChange?: (raw: string) => void;
  onClear?: () => Promise<unknown>;
}

function draftFromField(field: SettingsField): string {
  return field.sensitive ? '' : String(field.value ?? '');
}

function OriginIndicator({ isOverridden }: { isOverridden: boolean }) {
  const { t } = useTranslation('settings');
  const label = isOverridden ? t('origin.interface') : t('origin.config');

  return (
    <Tooltip.Provider delayDuration={150}>
      <Tooltip.Root>
        <Tooltip.Trigger asChild>
          <button
            type="button"
            className={`settings-origin-dot${isOverridden ? ' is-interface' : ' is-config'}`}
            aria-label={label}
          />
        </Tooltip.Trigger>
        <Tooltip.Portal>
          <Tooltip.Content className="settings-origin-tooltip" sideOffset={4}>
            {label}
            <Tooltip.Arrow className="settings-origin-tooltip-arrow" />
          </Tooltip.Content>
        </Tooltip.Portal>
      </Tooltip.Root>
    </Tooltip.Provider>
  );
}

// A single overridable settings field: shows its effective (or staged
// pending) value, a compact origin indicator revealing "Interface"/"Config"
// on hover/focus, and an icon-only reset-to-config control. Saving is owned
// by the enclosing group card (see SettingsGroupCard) - this component only
// reports draft changes upward via onChange. See frontend-app-settings-management.
export function SettingsFieldRow({ label, field, type = 'text', options, value, pending, onChange, onClear }: SettingsFieldRowProps) {
  const { t } = useTranslation('settings');
  const translateApiError = useApiErrorMessage();
  const [clearing, setClearing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isOverridden = field.origin === 'interface';
  const restartRequired = field.restart_required;
  const draft = value ?? draftFromField(field);

  const handleClear = async () => {
    if (!onClear) return;
    setClearing(true);
    setError(null);
    try {
      await onClear();
    } catch (err) {
      setError(translateApiError(err));
    } finally {
      setClearing(false);
    }
  };

  return (
    <div className={`settings-field${pending ? ' is-pending' : ''}`}>
      <div className="settings-field-row-header">
        <label className="settings-field-label" style={{ marginBottom: 0 }}>{label}</label>
        <div className="settings-field-indicators">
          {restartRequired && (
            <span className="badge badge-warning" style={{ fontSize: '0.65rem' }} title={t('restartRequired.hint')}>
              {t('restartRequired.label')}
            </span>
          )}
          <OriginIndicator isOverridden={isOverridden} />
          {isOverridden && onClear && (
            <button
              type="button"
              className="settings-reset-icon-btn"
              disabled={clearing}
              onClick={handleClear}
              title={t('resetToConfig')}
              aria-label={t('resetToConfig')}
            >
              <RotateCcw size={14} />
            </button>
          )}
        </div>
      </div>

      {type === 'boolean' ? (
        <ToggleSwitch
          checked={draft === 'true'}
          onChange={checked => onChange?.(checked ? 'true' : 'false')}
          ariaLabel={label}
        />
      ) : type === 'select' ? (
        <select
          className="custom-select"
          value={draft}
          aria-label={label}
          onChange={e => onChange?.(e.target.value)}
        >
          {options?.map(option => (
            <option key={option.value} value={option.value}>{t(option.labelKey)}</option>
          ))}
        </select>
      ) : (
        <input
          type={field.sensitive ? 'password' : type === 'number' ? 'number' : 'text'}
          className="custom-input"
          value={draft}
          placeholder={field.sensitive ? (field.is_set ? t('sensitive.set') : t('sensitive.notSet')) : undefined}
          onChange={e => onChange?.(e.target.value)}
        />
      )}

      {error && (
        <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>{error}</p>
      )}
    </div>
  );
}
