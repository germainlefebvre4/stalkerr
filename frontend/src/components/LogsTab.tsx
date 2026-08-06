import * as Tabs from '@radix-ui/react-tabs';
import { useTranslation } from 'react-i18next';
import { ProcessingLog } from '../types';
import { useIsMobile } from '../hooks/useMediaQuery';

interface LogsTabProps {
  logs: ProcessingLog[];
  logsLoading: boolean;
  onFetchLogs: () => void;
}

function getLogStatusBadgeClass(status: string): string {
  return status === 'success' ? 'badge-success' : status === 'failed' ? 'badge-failed' : 'badge-progress';
}

export function LogsTab({
  logs,
  logsLoading,
  onFetchLogs
}: LogsTabProps) {
  const { t, i18n } = useTranslation('logs');
  const isMobile = useIsMobile();
  return (
    <Tabs.Content value="logs" className="card tab-panel">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>{t('heading')}</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('subtitle')}</p>
        </div>
        <button onClick={onFetchLogs} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>{t('refresh')}</button>
      </div>

      {isMobile ? (
        <div>
          {logsLoading && logs.length === 0 ? (
            <div className="mobile-list-empty">{t('loading')}</div>
          ) : logs.length === 0 ? (
            <div className="mobile-list-empty">{t('empty')}</div>
          ) : (
            logs.map(log => (
              <div key={log.id} className="mobile-list-card">
                <div className="mobile-list-card-main">
                  <span className="mobile-list-card-title">{log.action}</span>
                  <span className="mobile-list-card-subtitle">{new Date(log.started_at).toLocaleString(i18n.language)}</span>
                </div>
                <span className={`badge ${getLogStatusBadgeClass(log.status)}`}>
                  {log.status === 'in_progress' ? t('status.inProgress') : log.status === 'success' ? t('status.success') : t('status.failed')}
                </span>
              </div>
            ))
          )}
        </div>
      ) : (
        <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
          <table className="custom-table">
            <thead>
              <tr>
                <th>{t('table.action')}</th>
                <th>{t('table.itemCount')}</th>
                <th>{t('table.status')}</th>
                <th>{t('table.startedAt')}</th>
                <th>{t('table.errorDetails')}</th>
              </tr>
            </thead>
            <tbody>
              {logsLoading && logs.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('loading')}</td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('empty')}</td>
                </tr>
              ) : (
                logs.map(log => (
                  <tr key={log.id}>
                    <td style={{ fontWeight: 700, color: 'var(--primary-slate)' }}>{log.action}</td>
                    <td style={{ fontWeight: 600 }}>{t('itemsCount', { count: log.item_count })}</td>
                    <td>
                      <span className={`badge ${getLogStatusBadgeClass(log.status)}`}>
                        {log.status === 'in_progress' ? t('status.inProgress') : log.status === 'success' ? t('status.success') : t('status.failed')}
                      </span>
                    </td>
                    <td style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{new Date(log.started_at).toLocaleString(i18n.language)}</td>
                    <td style={{ maxWidth: '300px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {log.error_message ? (
                        <span style={{ color: 'var(--status-failed-text)', fontWeight: 600 }} title={log.error_message}>{log.error_message}</span>
                      ) : (
                        <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>{t('noError')}</span>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}
    </Tabs.Content>
  );
}
