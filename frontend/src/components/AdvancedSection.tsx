import { useTranslation } from 'react-i18next';
import { EffectivePolicy, ScheduleWindow, ScheduleWindowInput, SettingsField } from '../types';
import { DOWNLOADS_FIELDS, LOGGING_FIELDS, M3U_FIELDS } from '../settingsFieldGroups';
import { SettingsGroupCard, SettingsGroupFieldSpec } from './SettingsGroupCard';
import { BandwidthScheduleSection } from './BandwidthScheduleSection';

interface AdvancedSectionProps {
  isExpanded: boolean;
  settings: SettingsField[];
  loading: boolean;
  onSetSetting: (key: string, value: unknown) => Promise<unknown>;
  onClearSetting: (key: string) => Promise<unknown>;
  searchQuery?: string;

  scheduleWindows: ScheduleWindow[];
  scheduleWindowsLoading: boolean;
  onCreateScheduleWindow: (input: ScheduleWindowInput) => Promise<void>;
  onUpdateScheduleWindow: (id: number, input: ScheduleWindowInput) => Promise<void>;
  onDeleteScheduleWindow: (id: number) => Promise<void>;
  effectivePolicy: EffectivePolicy | null;
  onFetchEffectivePolicy: () => void;
}

// Renders the Avancé tab's editable-only groups (Downloads tuning, Logging,
// M3U update interval).
export function AdvancedSection({
  isExpanded, settings, loading, onSetSetting, onClearSetting, searchQuery,
  scheduleWindows, scheduleWindowsLoading, onCreateScheduleWindow, onUpdateScheduleWindow, onDeleteScheduleWindow,
  effectivePolicy, onFetchEffectivePolicy,
}: AdvancedSectionProps) {
  const { t } = useTranslation('settings');

  if (!isExpanded) return null;

  if (loading && settings.length === 0) {
    return <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>{t('loading')}</div>;
  }

  const localizedFields = (prefix: string, fields: SettingsGroupFieldSpec[]) =>
    fields.map(f => ({ ...f, label: t(`fields.${prefix}.${f.label}`) }));

  return (
    <div className="settings-cards-grid">
      <SettingsGroupCard
        title={t('groups.downloads')}
        fields={localizedFields('downloads', DOWNLOADS_FIELDS)}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        searchQuery={searchQuery}
      />
      <SettingsGroupCard
        title={t('groups.logging')}
        fields={localizedFields('logging', LOGGING_FIELDS)}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        searchQuery={searchQuery}
      />
      <SettingsGroupCard
        title={t('groups.m3u')}
        fields={localizedFields('m3u', M3U_FIELDS)}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        searchQuery={searchQuery}
      />
      <div className="settings-group-card settings-group-card--full">
        <BandwidthScheduleSection
          isExpanded
          settings={settings}
          onSetSetting={onSetSetting}
          onClearSetting={onClearSetting}
          windows={scheduleWindows}
          windowsLoading={scheduleWindowsLoading}
          onCreateWindow={onCreateScheduleWindow}
          onUpdateWindow={onUpdateScheduleWindow}
          onDeleteWindow={onDeleteScheduleWindow}
          effectivePolicy={effectivePolicy}
          onFetchEffectivePolicy={onFetchEffectivePolicy}
          searchQuery={searchQuery}
        />
      </div>
    </div>
  );
}
