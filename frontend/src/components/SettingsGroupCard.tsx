import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { IntegrationTestService, SettingsField } from '../types';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';
import { SettingsFieldRow } from './SettingsFieldRow';
import { api } from '../services/api';

export interface SettingsGroupFieldSpec {
  key: string;
  label: string;
  type?: 'text' | 'number' | 'boolean';
}

export interface SettingsGroupCardTestConfig {
  service: IntegrationTestService;
  urlKey: string;
  apiKeyKey: string;
}

type TestState = 'idle' | 'testing' | 'ok' | 'ko';

interface SettingsGroupCardProps {
  title?: string;
  fields: SettingsGroupFieldSpec[];
  settings: SettingsField[];
  onSetSetting: (key: string, value: unknown) => Promise<unknown>;
  onClearSetting: (key: string) => Promise<unknown>;
  /** Case-insensitive substring filter over each field's label; matching fields (and their card) hide when nothing matches. */
  searchQuery?: string;
  /** When set, renders a "Test connection" action (while the card has an unsaved change) that checks the in-progress url/api_key values live. */
  testConfig?: SettingsGroupCardTestConfig;
}

function draftFromField(field: SettingsField): string {
  return field.sensitive ? '' : String(field.value ?? '');
}

function applyType(raw: string, type?: 'text' | 'number' | 'boolean'): unknown {
  if (type === 'number') return Number(raw);
  if (type === 'boolean') return raw === 'true';
  return raw;
}

// Renders a logical group of overridable fields as one card, owning a
// pending-changes map and showing a save/discard bar only while it is
// non-empty. See the Editing and Clearing a Settings Field requirement and
// the Settings Field Layout requirement (960px-capped, compact-vs-full-width
// grid) in frontend-app-settings-management.
export function SettingsGroupCard({ title, fields, settings, onSetSetting, onClearSetting, searchQuery, testConfig }: SettingsGroupCardProps) {
  const { t } = useTranslation('settings');
  const translateApiError = useApiErrorMessage();
  const [pending, setPending] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [testState, setTestState] = useState<TestState>('idle');
  const [testReason, setTestReason] = useState<string | undefined>(undefined);

  const query = (searchQuery ?? '').trim().toLowerCase();
  const visibleFields = query
    ? fields.filter(f => f.label.toLowerCase().includes(query))
    : fields;

  const resolvedFields = visibleFields
    .map(spec => ({ spec, field: settings.find(f => f.key === spec.key) }))
    .filter((entry): entry is { spec: SettingsGroupFieldSpec; field: SettingsField } => !!entry.field);

  if (resolvedFields.length === 0) return null;

  const isFieldPending = (key: string) => Object.prototype.hasOwnProperty.call(pending, key);
  const hasPending = Object.keys(pending).length > 0;

  // Resolves a settings key's current in-form value the same way handleSave
  // resolves what it is about to submit: a pending edit if present, else the
  // field's current draft (a sensitive field's untouched draft is '').
  const resolveFieldValue = (key: string): string => {
    if (isFieldPending(key)) return pending[key];
    const field = settings.find(f => f.key === key);
    return field ? draftFromField(field) : '';
  };

  const handleChange = (spec: SettingsGroupFieldSpec, field: SettingsField, raw: string) => {
    setError(null);
    setTestState('idle');
    setTestReason(undefined);
    setPending(prev => {
      const next = { ...prev };
      const unchanged = field.sensitive ? raw === '' : raw === draftFromField(field);
      if (unchanged) {
        delete next[spec.key];
      } else {
        next[spec.key] = raw;
      }
      return next;
    });
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    const keys = Object.keys(pending);
    const results = await Promise.allSettled(
      keys.map(key => {
        const spec = fields.find(f => f.key === key);
        return onSetSetting(key, applyType(pending[key], spec?.type));
      })
    );

    const failedKeys: string[] = [];
    results.forEach((result, i) => {
      if (result.status === 'rejected') failedKeys.push(keys[i]);
    });

    if (failedKeys.length > 0) {
      const firstFailure = results.find(r => r.status === 'rejected') as PromiseRejectedResult;
      setError(translateApiError(firstFailure.reason));
      setPending(prev => {
        const next: Record<string, string> = {};
        for (const key of failedKeys) next[key] = prev[key];
        return next;
      });
    } else {
      setPending({});
    }
    setSaving(false);
  };

  const handleDiscard = () => {
    setPending({});
    setError(null);
    setTestState('idle');
    setTestReason(undefined);
  };

  const handleTest = async () => {
    if (!testConfig) return;
    setTestState('testing');
    setTestReason(undefined);
    try {
      const result = await api.testIntegration(
        testConfig.service,
        resolveFieldValue(testConfig.urlKey),
        resolveFieldValue(testConfig.apiKeyKey)
      );
      setTestState(result.status === 'ok' ? 'ok' : 'ko');
      setTestReason(result.reason);
    } catch (err) {
      setTestState('idle');
      setError(translateApiError(err));
    }
  };

  const cardClassName = `settings-group-card${fields.length > 6 ? ' settings-group-card--full' : ' settings-group-card--compact'}`;

  return (
    <div className={cardClassName}>
      {title && <h4 className="settings-group-card-title">{title}</h4>}
      <div className="settings-group-card-fields">
        {resolvedFields.map(({ spec, field }) => (
          <div key={spec.key} className={spec.type === 'text' || spec.type === undefined ? 'settings-field-full' : 'settings-field-compact'}>
            <SettingsFieldRow
              label={spec.label}
              field={field}
              type={spec.type}
              value={isFieldPending(spec.key) ? pending[spec.key] : draftFromField(field)}
              pending={isFieldPending(spec.key)}
              onChange={raw => handleChange(spec, field, raw)}
              onClear={() => onClearSetting(spec.key)}
            />
          </div>
        ))}
      </div>

      {error && (
        <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>{error}</p>
      )}

      {hasPending && (
        <div className="settings-group-card-actions">
          {testConfig && (
            <>
              {(testState === 'ok' || testState === 'ko') && (
                <span className={`badge ${testState === 'ok' ? 'badge-success' : 'badge-failed'}`}>
                  {t(`dialogs:systemStatus.states.${testState}`)}
                  {testState === 'ko' && testReason ? ` – ${t(`dialogs:systemStatus.reasons.${testReason}`)}` : ''}
                </span>
              )}
              <button
                type="button"
                className="btn-secondary"
                disabled={saving || testState === 'testing' || resolveFieldValue(testConfig.urlKey) === ''}
                onClick={handleTest}
              >
                {testState === 'testing' ? t('testing') : t('testConnection')}
              </button>
            </>
          )}
          <button type="button" className="btn-secondary" disabled={saving} onClick={handleDiscard}>
            {t('cancel')}
          </button>
          <button type="button" className="btn-primary" disabled={saving} onClick={handleSave}>
            {saving ? t('saving') : t('save')}
          </button>
        </div>
      )}
    </div>
  );
}
