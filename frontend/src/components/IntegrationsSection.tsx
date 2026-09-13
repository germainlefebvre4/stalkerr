import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { SettingsField } from '../types';
import { SettingsFieldRow } from './SettingsFieldRow';

interface FieldSpec {
  key: string;
  labelKey: string;
  type?: 'text' | 'number' | 'boolean';
}

const RADARR_FIELDS: FieldSpec[] = [
  { key: 'radarr.enabled', labelKey: 'radarr.enabled', type: 'boolean' },
  { key: 'radarr.url', labelKey: 'radarr.url' },
  { key: 'radarr.api_key', labelKey: 'radarr.apiKey' },
  { key: 'radarr.sync_interval', labelKey: 'radarr.syncInterval', type: 'number' },
  { key: 'radarr.quality_profile_id', labelKey: 'radarr.qualityProfileId', type: 'number' },
];

const SONARR_FIELDS: FieldSpec[] = [
  { key: 'sonarr.enabled', labelKey: 'sonarr.enabled', type: 'boolean' },
  { key: 'sonarr.url', labelKey: 'sonarr.url' },
  { key: 'sonarr.api_key', labelKey: 'sonarr.apiKey' },
  { key: 'sonarr.sync_interval', labelKey: 'sonarr.syncInterval', type: 'number' },
  { key: 'sonarr.quality_profile_id', labelKey: 'sonarr.qualityProfileId', type: 'number' },
];

const TMDB_FIELDS: FieldSpec[] = [
  { key: 'tmdb.enabled', labelKey: 'tmdb.enabled', type: 'boolean' },
  { key: 'tmdb.api_key', labelKey: 'tmdb.apiKey' },
  { key: 'tmdb.language', labelKey: 'tmdb.language' },
  { key: 'tmdb.requests_per_second', labelKey: 'tmdb.requestsPerSecond', type: 'number' },
];

const JELLYFIN_FIELDS: FieldSpec[] = [
  { key: 'jellyfin.enabled', labelKey: 'jellyfin.enabled', type: 'boolean' },
  { key: 'jellyfin.url', labelKey: 'jellyfin.url' },
  { key: 'jellyfin.api_key', labelKey: 'jellyfin.apiKey' },
];

interface IntegrationsSectionProps {
  isExpanded: boolean;
  settings: SettingsField[];
  loading: boolean;
  onFetchSettings: () => void;
  onSetSetting: (key: string, value: unknown) => Promise<unknown>;
  onClearSetting: (key: string) => Promise<unknown>;
}

export function IntegrationsSection({
  isExpanded, settings, loading, onFetchSettings, onSetSetting, onClearSetting,
}: IntegrationsSectionProps) {
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

  const renderGroup = (title: string, fields: FieldSpec[]) => {
    const rows = fields
      .map(({ key, labelKey, type }) => {
        const field = settings.find(f => f.key === key);
        if (!field) return null;
        return (
          <SettingsFieldRow
            key={key}
            label={t(`fields.${labelKey}`)}
            field={field}
            type={type}
            onSave={value => onSetSetting(key, value)}
            onClear={() => onClearSetting(key)}
          />
        );
      })
      .filter(Boolean);

    if (rows.length === 0) return null;

    return (
      <div style={{ marginBottom: '1.5rem' }}>
        <h4 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--primary-slate)', marginBottom: '0.5rem' }}>{title}</h4>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>{rows}</div>
      </div>
    );
  };

  return (
    <div>
      {renderGroup(t('groups.radarr'), RADARR_FIELDS)}
      {renderGroup(t('groups.sonarr'), SONARR_FIELDS)}
      {renderGroup(t('groups.tmdb'), TMDB_FIELDS)}
      {renderGroup(t('groups.jellyfin'), JELLYFIN_FIELDS)}
    </div>
  );
}
