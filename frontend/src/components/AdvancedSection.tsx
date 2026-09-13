import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { SettingsField, BootstrapField } from '../types';
import { SettingsFieldRow } from './SettingsFieldRow';

interface FieldSpec {
  key: string;
  labelKey: string;
  type?: 'text' | 'number' | 'boolean';
}

const DOWNLOADS_FIELDS: FieldSpec[] = [
  { key: 'downloads.movies_path', labelKey: 'moviesPath' },
  { key: 'downloads.tvshows_path', labelKey: 'tvshowsPath' },
  { key: 'downloads.temp_dir', labelKey: 'tempDir' },
  { key: 'downloads.max_parallel', labelKey: 'maxParallel', type: 'number' },
  { key: 'downloads.timeout', labelKey: 'timeout', type: 'number' },
  { key: 'downloads.retry_attempts', labelKey: 'retryAttempts', type: 'number' },
  { key: 'downloads.resume_enabled', labelKey: 'resumeEnabled', type: 'boolean' },
  { key: 'downloads.progress_interval_mb', labelKey: 'progressIntervalMb', type: 'number' },
  { key: 'downloads.progress_interval_seconds', labelKey: 'progressIntervalSeconds', type: 'number' },
  { key: 'downloads.lock_timeout_minutes', labelKey: 'lockTimeoutMinutes', type: 'number' },
  { key: 'downloads.max_retry_attempts', labelKey: 'maxRetryAttempts', type: 'number' },
  { key: 'downloads.min_file_size_mb', labelKey: 'minFileSizeMb', type: 'number' },
  { key: 'downloads.force_tier_probability', labelKey: 'forceTierProbability', type: 'number' },
];

const LOGGING_FIELDS: FieldSpec[] = [
  { key: 'logging.app.level', labelKey: 'appLevel' },
  { key: 'logging.database.level', labelKey: 'databaseLevel' },
  { key: 'logging.format', labelKey: 'format' },
];

const M3U_FIELDS: FieldSpec[] = [
  { key: 'm3u.update_interval', labelKey: 'updateInterval', type: 'number' },
];

const BOOTSTRAP_FIELDS: { key: string; labelKey: string }[] = [
  { key: 'database.host', labelKey: 'databaseHost' },
  { key: 'database.port', labelKey: 'databasePort' },
  { key: 'database.user', labelKey: 'databaseUser' },
  { key: 'database.password', labelKey: 'databasePassword' },
  { key: 'database.dbname', labelKey: 'databaseName' },
  { key: 'database.sslmode', labelKey: 'databaseSslMode' },
  { key: 'api.port', labelKey: 'apiPort' },
  { key: 'metrics.enabled', labelKey: 'metricsEnabled' },
  { key: 'metrics.port', labelKey: 'metricsPort' },
  { key: 'metrics.path', labelKey: 'metricsPath' },
];

function BootstrapFieldRow({ label, field }: { label: string; field: BootstrapField }) {
  const { t } = useTranslation('settings');
  const displayValue = field.sensitive
    ? (field.is_set ? t('sensitive.set') : t('sensitive.notSet'))
    : String(field.value ?? '');

  return (
    <div className="settings-field">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span className="settings-field-label" style={{ marginBottom: 0 }}>{label}</span>
        <span className="badge badge-muted" style={{ fontSize: '0.65rem' }}>{t('origin.config')}</span>
      </div>
      <div
        className="custom-input"
        style={{ opacity: 0.7, cursor: 'not-allowed', userSelect: 'text' }}
        aria-readonly="true"
      >
        {displayValue}
      </div>
    </div>
  );
}

interface AdvancedSectionProps {
  isExpanded: boolean;
  settings: SettingsField[];
  bootstrap: BootstrapField[];
  loading: boolean;
  onFetchSettings: () => void;
  onSetSetting: (key: string, value: unknown) => Promise<unknown>;
  onClearSetting: (key: string) => Promise<unknown>;
}

export function AdvancedSection({
  isExpanded, settings, bootstrap, loading, onFetchSettings, onSetSetting, onClearSetting,
}: AdvancedSectionProps) {
  const { t } = useTranslation('settings');

  useEffect(() => {
    if (!isExpanded) return;
    onFetchSettings();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isExpanded]);

  if (!isExpanded) return null;

  if (loading && settings.length === 0) {
    return <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>{t('loading')}</div>;
  }

  const renderFields = (fields: FieldSpec[], prefix: string) => fields.map(({ key, labelKey, type }) => {
    const field = settings.find(f => f.key === key);
    if (!field) return null;
    return (
      <SettingsFieldRow
        key={key}
        label={t(`fields.${prefix}.${labelKey}`)}
        field={field}
        type={type}
        onSave={value => onSetSetting(key, value)}
        onClear={() => onClearSetting(key)}
      />
    );
  });

  return (
    <div>
      <div style={{ marginBottom: '1.5rem' }}>
        <h4 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--primary-slate)', marginBottom: '0.5rem' }}>{t('groups.downloads')}</h4>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>{renderFields(DOWNLOADS_FIELDS, 'downloads')}</div>
      </div>

      <div style={{ marginBottom: '1.5rem' }}>
        <h4 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--primary-slate)', marginBottom: '0.5rem' }}>{t('groups.logging')}</h4>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>{renderFields(LOGGING_FIELDS, 'logging')}</div>
      </div>

      <div style={{ marginBottom: '1.5rem' }}>
        <h4 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--primary-slate)', marginBottom: '0.5rem' }}>{t('groups.m3u')}</h4>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>{renderFields(M3U_FIELDS, 'm3u')}</div>
      </div>

      <div>
        <h4 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--primary-slate)', marginBottom: '0.5rem' }}>{t('groups.bootstrap')}</h4>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', marginBottom: '0.75rem' }}>{t('bootstrapHint')}</p>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {BOOTSTRAP_FIELDS.map(({ key, labelKey }) => {
            const field = bootstrap.find(f => f.key === key);
            if (!field) return null;
            return <BootstrapFieldRow key={key} label={t(`fields.bootstrap.${labelKey}`)} field={field} />;
          })}
        </div>
      </div>
    </div>
  );
}
