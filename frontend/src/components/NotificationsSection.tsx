import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { SettingsField } from '../types';
import { SettingsFieldRow } from './SettingsFieldRow';

interface FieldSpec {
  key: string;
  labelKey: string;
  type?: 'text' | 'number' | 'boolean';
}

const NOTIFICATIONS_FIELDS: FieldSpec[] = [
  { key: 'notifications.enabled', labelKey: 'enabled', type: 'boolean' },
  { key: 'notifications.ntfy.enabled', labelKey: 'ntfyEnabled', type: 'boolean' },
  { key: 'notifications.ntfy.server_url', labelKey: 'ntfyServerUrl' },
  { key: 'notifications.ntfy.topic', labelKey: 'ntfyTopic' },
  { key: 'notifications.ntfy.auth_token', labelKey: 'ntfyAuthToken' },
];

interface NotificationsSectionProps {
  isExpanded: boolean;
  settings: SettingsField[];
  loading: boolean;
  onFetchSettings: () => void;
  onSetSetting: (key: string, value: unknown) => Promise<unknown>;
  onClearSetting: (key: string) => Promise<unknown>;
}

export function NotificationsSection({
  isExpanded, settings, loading, onFetchSettings, onSetSetting, onClearSetting,
}: NotificationsSectionProps) {
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

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
      {NOTIFICATIONS_FIELDS.map(({ key, labelKey, type }) => {
        const field = settings.find(f => f.key === key);
        if (!field) return null;
        return (
          <SettingsFieldRow
            key={key}
            label={t(`fields.notifications.${labelKey}`)}
            field={field}
            type={type}
            onSave={value => onSetSetting(key, value)}
            onClear={() => onClearSetting(key)}
          />
        );
      })}
    </div>
  );
}
