import { useTranslation } from 'react-i18next';
import { SettingsField } from '../types';
import { RADARR_FIELDS, SONARR_FIELDS, TMDB_FIELDS, JELLYFIN_FIELDS } from '../settingsFieldGroups';
import { SettingsGroupCard, SettingsGroupFieldSpec } from './SettingsGroupCard';

interface IntegrationsSectionProps {
  isExpanded: boolean;
  settings: SettingsField[];
  loading: boolean;
  onSetSetting: (key: string, value: unknown) => Promise<unknown>;
  onClearSetting: (key: string) => Promise<unknown>;
  searchQuery?: string;
}

// Renders the Radarr/Sonarr/TMDB/Jellyfin field groups as cards under the
// "Intégrations" tab. See the Settings Field Layout requirement.
export function IntegrationsSection({
  isExpanded, settings, loading, onSetSetting, onClearSetting, searchQuery,
}: IntegrationsSectionProps) {
  const { t } = useTranslation('settings');

  if (!isExpanded) return null;

  if (loading && settings.length === 0) {
    return <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>{t('loading')}</div>;
  }

  const localizedFields = (fields: SettingsGroupFieldSpec[]) =>
    fields.map(f => ({ ...f, label: t(`fields.${f.label}`) }));

  return (
    <div className="settings-cards-grid">
      <SettingsGroupCard
        title={t('groups.radarr')}
        fields={localizedFields(RADARR_FIELDS)}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        searchQuery={searchQuery}
      />
      <SettingsGroupCard
        title={t('groups.sonarr')}
        fields={localizedFields(SONARR_FIELDS)}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        searchQuery={searchQuery}
      />
      <SettingsGroupCard
        title={t('groups.tmdb')}
        fields={localizedFields(TMDB_FIELDS)}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        searchQuery={searchQuery}
      />
      <SettingsGroupCard
        title={t('groups.jellyfin')}
        fields={localizedFields(JELLYFIN_FIELDS)}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        searchQuery={searchQuery}
      />
    </div>
  );
}
