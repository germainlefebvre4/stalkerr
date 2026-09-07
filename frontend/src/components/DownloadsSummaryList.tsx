import * as Progress from '@radix-ui/react-progress';
import { useTranslation } from 'react-i18next';
import { TFunction } from 'i18next';
import { DownloadEnriched } from '../types';
import { useIsMobile } from '../hooks/useMediaQuery';

interface DownloadsSummaryListProps {
  downloads: DownloadEnriched[];
  loading: boolean;
  onRowClick: (item: DownloadEnriched) => void;
}

function filepathBase(path: string) {
  const parts = path.split('/');
  return parts[parts.length - 1];
}

function getStatusInfo(item: DownloadEnriched, t: TFunction, isMobile: boolean) {
  if (item.status === 'completed') {
    const label = isMobile ? t('status.completed') : `${t('status.completed')} ${t('status.completedText')}`;
    return { label, badgeClass: 'badge-success' };
  }
  if (item.status === 'downloading') {
    const label = isMobile ? t('status.downloading') : `${t('status.downloading')} ${t('status.downloadingText')}`;
    return { label, badgeClass: 'badge-progress' };
  }
  if (item.status === 'failed') {
    if (item.retry_count > 0) {
      const label = isMobile
        ? t('status.failedWithRetry', { count: item.retry_count })
        : `${t('status.failed')} ${t('status.failedText')} (${item.retry_count}×)`;
      return { label, badgeClass: 'badge-failed' };
    }
    const label = isMobile ? t('status.failed') : `${t('status.failed')} ${t('status.failedText')}`;
    return { label, badgeClass: 'badge-failed' };
  }
  if (item.status === 'retrying') {
    return { label: t('status.retrying'), badgeClass: 'badge-pending' };
  }
  if (item.status === 'cancelled') {
    return { label: t('status.cancelled'), badgeClass: 'badge-neutral' };
  }
  const label = isMobile ? t('status.pending') : `${t('status.pending')} ${t('status.pendingText')}`;
  return { label, badgeClass: 'badge-pending' };
}

function buildRow(item: DownloadEnriched, t: TFunction, isMobile: boolean) {
  const total = item.total_bytes || 0;
  const downloaded = item.bytes_downloaded || 0;
  const progress = total > 0 ? Math.round((downloaded / total) * 100) : 0;
  const isProgressStatus = item.status === 'downloading' || item.status === 'retrying';
  const title = item.content?.title || (item.download_path ? filepathBase(item.download_path) : item.url);
  const year = item.content?.year ? `(${item.content.year})` : '';
  const typeIcon = item.content?.type === 'movies' ? '🎬' : item.content?.type === 'tvshows' ? '📺' : '🔗';
  const { label: statusLabel, badgeClass: statusBadgeClass } = getStatusInfo(item, t, isMobile);

  return { item, title, year, typeIcon, statusLabel, statusBadgeClass, isProgressStatus, progress };
}

export function DownloadsSummaryList({ downloads, loading, onRowClick }: DownloadsSummaryListProps) {
  const { t } = useTranslation('downloads');
  const isMobile = useIsMobile();

  const rows = downloads.map(item => buildRow(item, t, isMobile));

  if (isMobile) {
    return (
      <div>
        {loading ? (
          <div className="mobile-list-empty">{t('loading')}</div>
        ) : downloads.length === 0 ? (
          <div className="mobile-list-empty">{t('empty')}</div>
        ) : (
          rows.map(({ item, title, year, typeIcon, statusLabel, statusBadgeClass, isProgressStatus, progress }) => (
            <div key={item.id} className="mobile-list-card" onClick={() => onRowClick(item)}>
              <div className="mobile-list-card-main">
                <span className="mobile-list-card-title">{typeIcon} {title} {year}</span>
                {isProgressStatus && (
                  <span className="mobile-list-card-subtitle">{t('progress', { percent: progress })}</span>
                )}
              </div>
              <span className={`badge ${statusBadgeClass}`}>{statusLabel}</span>
            </div>
          ))
        )}
      </div>
    );
  }

  return (
    <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
      <table className="custom-table">
        <thead>
          <tr>
            <th>{t('summary.headers.title')}</th>
            <th>{t('summary.headers.status')}</th>
            <th>{t('summary.headers.progress')}</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={3} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
                <span style={{ fontWeight: 600 }}>{t('loading')}</span>
              </td>
            </tr>
          ) : downloads.length === 0 ? (
            <tr>
              <td colSpan={3} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('empty')}</td>
            </tr>
          ) : (
            rows.map(({ item, title, year, typeIcon, statusLabel, statusBadgeClass, isProgressStatus, progress }) => (
              <tr key={item.id} className="clickable-row" onClick={() => onRowClick(item)}>
                <td style={{ fontWeight: 600, color: 'var(--primary-slate)' }}>
                  <span style={{ marginRight: '0.5rem' }}>{typeIcon}</span>{title} {year}
                </td>
                <td>
                  <span className={`badge ${statusBadgeClass}`}>{statusLabel}</span>
                </td>
                <td style={{ minWidth: '160px' }}>
                  {isProgressStatus && (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.2rem' }}>
                      <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', fontWeight: 600 }}>
                        {t('progress', { percent: progress })}
                      </span>
                      <Progress.Root value={progress} className="progress-root">
                        <Progress.Indicator className="progress-indicator" style={{ width: `${progress}%` }} />
                      </Progress.Root>
                    </div>
                  )}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
