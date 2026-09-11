import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useSystemStatus } from '../hooks/useSystemStatus';
import { ServiceStatus, ServiceStatusState, DiskUsageEntry } from '../types';

interface SystemStatusSectionProps {
  isExpanded: boolean;
}

const BADGE_CLASS: Record<ServiceStatusState, string> = {
  ok: 'badge-success',
  ko: 'badge-failed',
  not_configured: 'badge-pending',
};

function formatBytes(bytes: number): string {
  if (bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / Math.pow(1024, exp);
  return `${exp === 0 ? value : value.toFixed(1)} ${units[exp]}`;
}

function ServiceRow({ label, status, t }: { label: string; status: ServiceStatus; t: (key: string) => string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.6rem 0', borderBottom: '1px solid var(--border-color)' }}>
      <span style={{ fontWeight: 600 }}>{label}</span>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        {status.status === 'ko' && status.reason && (
          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            {t(`systemStatus.reasons.${status.reason}`)}
          </span>
        )}
        <span className={`badge ${BADGE_CLASS[status.status]}`}>{t(`systemStatus.states.${status.status}`)}</span>
      </div>
    </div>
  );
}

function DiskUsageRow({ entry, t }: { entry: DiskUsageEntry; t: (key: string) => string }) {
  return (
    <div style={{ padding: '0.6rem 0', borderBottom: '1px solid var(--border-color)', fontSize: '0.85rem' }}>
      <div style={{ fontWeight: 600 }}>{entry.paths.join(', ')}</div>
      {entry.unavailable || entry.total == null || entry.available == null ? (
        <div style={{ color: 'var(--status-failed-text)' }}>
          {t(`systemStatus.reasons.${entry.reason || 'unavailable'}`)}
        </div>
      ) : (
        <div style={{ color: 'var(--text-secondary)' }}>
          {t('systemStatus.disk.used')}: {formatBytes(entry.total - entry.available)} / {formatBytes(entry.total)}
          {' '}({formatBytes(entry.available)} {t('systemStatus.disk.available')})
        </div>
      )}
    </div>
  );
}

// Content for the drawer's "Système" disclosure. Fetches only while expanded:
// no request before the first expand, and a fresh one on every re-expand.
export function SystemStatusSection({ isExpanded }: SystemStatusSectionProps) {
  const { t } = useTranslation('dialogs');
  const { status, loading, error, fetchStatus } = useSystemStatus();

  useEffect(() => {
    if (!isExpanded) return;
    void Promise.resolve().then(fetchStatus);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isExpanded]);

  if (!isExpanded) return null;

  return (
    <div style={{ paddingTop: '0.75rem' }}>
      {loading && (
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>{t('systemStatus.loading')}</p>
      )}

      {error && !loading && (
        <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600, marginBottom: '1rem', border: '1px solid var(--status-failed-border)' }}>
          ⚠️ {t('systemStatus.error')}
        </div>
      )}

      {status && !loading && (
        <>
          <div style={{ marginBottom: '1.5rem' }}>
            <ServiceRow label={t('systemStatus.rows.database')} status={status.database} t={t} />
            <ServiceRow label={t('systemStatus.rows.radarr')} status={status.radarr} t={t} />
            <ServiceRow label={t('systemStatus.rows.sonarr')} status={status.sonarr} t={t} />
            <ServiceRow label={t('systemStatus.rows.tmdb')} status={status.tmdb} t={t} />
          </div>

          <div style={{ marginBottom: '1.5rem' }}>
            <h4 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>
              {t('systemStatus.disk.title')}
            </h4>
            {status.disk.length === 0 ? (
              <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>{t('systemStatus.disk.empty')}</p>
            ) : (
              status.disk.map((entry, i) => <DiskUsageRow key={i} entry={entry} t={t} />)
            )}
          </div>

          <div>
            <h4 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>
              {t('systemStatus.build.title')}
            </h4>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
              <div>{t('systemStatus.build.version')}: {status.version}</div>
              <div>{t('systemStatus.build.commit')}: {status.commit}</div>
              <div>{t('systemStatus.build.date')}: {status.date}</div>
            </div>
          </div>
        </>
      )}

      <div style={{ display: 'flex', justifyContent: 'flex-end', paddingTop: '1rem' }}>
        <button onClick={fetchStatus} disabled={loading} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
          {t('systemStatus.refresh')}
        </button>
      </div>
    </div>
  );
}
