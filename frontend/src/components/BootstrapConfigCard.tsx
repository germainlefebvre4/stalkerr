import { useTranslation } from 'react-i18next';
import { BootstrapField } from '../types';
import { SettingsFieldRow } from './SettingsFieldRow';

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

interface BootstrapConfigCardProps {
  bootstrap: BootstrapField[];
  searchQuery?: string;
}

// Read-only card for the bootstrap configuration (database connection, API
// port, metrics port/enabled), moved here from the former "Avancé" grouping.
// See the Bootstrap Configuration Display requirement.
export function BootstrapConfigCard({ bootstrap, searchQuery }: BootstrapConfigCardProps) {
  const { t } = useTranslation('settings');

  const query = (searchQuery ?? '').trim().toLowerCase();
  const rows = BOOTSTRAP_FIELDS
    .map(({ key, labelKey }) => ({ key, label: t(`fields.bootstrap.${labelKey}`), field: bootstrap.find(f => f.key === key) }))
    .filter((entry): entry is { key: string; label: string; field: BootstrapField } => !!entry.field)
    .filter(entry => !query || entry.label.toLowerCase().includes(query));

  if (rows.length === 0) return null;

  return (
    <div className="settings-group-card settings-group-card--full">
      <h4 className="settings-group-card-title">{t('groups.bootstrap')}</h4>
      <p className="settings-field-hint" style={{ marginBottom: '0.75rem' }}>{t('bootstrapHint')}</p>
      <div className="settings-group-card-fields">
        {rows.map(({ key, label, field }) => (
          <div key={key} className="settings-field-full">
            <SettingsFieldRow label={label} field={field} readOnly />
          </div>
        ))}
      </div>
    </div>
  );
}
