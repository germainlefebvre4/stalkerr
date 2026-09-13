import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { SettingsField } from '../types';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';

interface SettingsFieldRowProps {
  label: string;
  field: SettingsField;
  type?: 'text' | 'number' | 'boolean';
  onSave: (value: unknown) => Promise<unknown>;
  onClear: () => Promise<unknown>;
}

function draftFromField(field: SettingsField): string {
  return field.sensitive ? '' : String(field.value ?? '');
}

// A single overridable settings field: shows its effective value, an
// "Interface"/"Config" origin badge, and (for sensitive fields) a masked
// input that only submits a change when the user explicitly types one. See
// the frontend-app-settings-management spec.
export function SettingsFieldRow({ label, field, type = 'text', onSave, onClear }: SettingsFieldRowProps) {
  const { t } = useTranslation('settings');
  const translateApiError = useApiErrorMessage();
  const [draft, setDraft] = useState<string>(() => draftFromField(field));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Adjust the draft during render when the field's effective value changes
  // from under us (e.g. another tab cleared the override) - the React-
  // recommended alternative to a useEffect+setState pair for this exact
  // "reset local state to match a changed prop" case.
  const [lastFieldValue, setLastFieldValue] = useState(field.value);
  if (!field.sensitive && field.value !== lastFieldValue) {
    setLastFieldValue(field.value);
    setDraft(draftFromField(field));
  }

  const isOverridden = field.origin === 'interface';
  const saveDisabled = saving || (field.sensitive && draft === '');

  const handleSave = async () => {
    // Sensitive fields must never submit a change unless the user typed
    // into them: an empty draft here means "left unchanged".
    if (field.sensitive && draft === '') return;

    setSaving(true);
    setError(null);
    try {
      let value: unknown = draft;
      if (type === 'number') value = Number(draft);
      if (type === 'boolean') value = draft === 'true';
      await onSave(value);
      if (field.sensitive) setDraft('');
    } catch (err) {
      setError(translateApiError(err));
    } finally {
      setSaving(false);
    }
  };

  const handleClear = async () => {
    setSaving(true);
    setError(null);
    try {
      await onClear();
      setDraft('');
    } catch (err) {
      setError(translateApiError(err));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="settings-field">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.35rem' }}>
        <label className="settings-field-label" style={{ marginBottom: 0 }}>{label}</label>
        <div style={{ display: 'flex', gap: '0.35rem' }}>
          <span
            className={`badge ${isOverridden ? 'badge-progress' : 'badge-muted'}`}
            style={{ fontSize: '0.65rem' }}
          >
            {isOverridden ? t('origin.interface') : t('origin.config')}
          </span>
          {field.restart_required && (
            <span className="badge badge-warning" style={{ fontSize: '0.65rem' }} title={t('restartRequired.hint')}>
              {t('restartRequired.label')}
            </span>
          )}
        </div>
      </div>

      {type === 'boolean' ? (
        <select className="custom-select" value={draft || 'false'} onChange={e => setDraft(e.target.value)}>
          <option value="true">{t('boolean.true')}</option>
          <option value="false">{t('boolean.false')}</option>
        </select>
      ) : (
        <input
          type={field.sensitive ? 'password' : type === 'number' ? 'number' : 'text'}
          className="custom-input"
          value={draft}
          placeholder={field.sensitive ? (field.is_set ? t('sensitive.set') : t('sensitive.notSet')) : undefined}
          onChange={e => setDraft(e.target.value)}
        />
      )}

      {error && (
        <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>{error}</p>
      )}

      <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.35rem' }}>
        <button
          type="button"
          className="btn-primary"
          disabled={saveDisabled}
          onClick={handleSave}
          style={{ padding: '0.3rem 0.75rem', fontSize: '0.8rem' }}
        >
          {saving ? t('saving') : t('save')}
        </button>
        {isOverridden && (
          <button
            type="button"
            className="btn-secondary"
            disabled={saving}
            onClick={handleClear}
            style={{ padding: '0.3rem 0.75rem', fontSize: '0.8rem' }}
          >
            {t('resetToConfig')}
          </button>
        )}
      </div>
    </div>
  );
}
