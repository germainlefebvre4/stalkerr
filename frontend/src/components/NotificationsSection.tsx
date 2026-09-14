import { useTranslation } from 'react-i18next';
import { SettingsField } from '../types';
import { NOTIFICATIONS_FIELDS } from '../settingsFieldGroups';
import { SettingsGroupCard } from './SettingsGroupCard';

interface NotificationsSectionProps {
  isExpanded: boolean;
  settings: SettingsField[];
  loading: boolean;
  onSetSetting: (key: string, value: unknown) => Promise<unknown>;
  onClearSetting: (key: string) => Promise<unknown>;
  searchQuery?: string;
}

export function NotificationsSection({
  isExpanded, settings, loading, onSetSetting, onClearSetting, searchQuery,
}: NotificationsSectionProps) {
  const { t } = useTranslation('settings');

  if (!isExpanded) return null;

  if (loading && settings.length === 0) {
    return <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>{t('loading')}</div>;
  }

  const fields = NOTIFICATIONS_FIELDS.map(f => ({ ...f, label: t(`fields.notifications.${f.label}`) }));

  return (
    <div className="settings-cards-grid">
      <SettingsGroupCard
        fields={fields}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        searchQuery={searchQuery}
      />
    </div>
  );
}
