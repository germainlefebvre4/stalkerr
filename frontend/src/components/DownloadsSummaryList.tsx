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

function getStatusInfo(item: DownloadEnriched, t: TFunction) {
  if (item.status === 'completed') {
    return { label: t('status.completed'), badgeClass: 'badge-success' };
  }
  if (item.status === 'downloading') {
    return { label: t('status.downloading'), badgeClass: 'badge-progress' };
  }
  if (item.status === 'failed') {
    return {
      label: item.retry_count > 0 ? t('status.failedWithRetry', { count: item.retry_count }) : t('status.failed'),
      badgeClass: 'badge-failed',
    };
  }
  if (item.status === 'retrying') {
    return { label: t('status.retrying'), badgeClass: 'badge-pending' };
  }
  return { label: t('status.pending'), badgeClass: 'badge-pending' };
}

function buildRow(item: DownloadEnriched, t: TFunction) {
  const total = item.total_bytes || 0;
  const downloaded = item.bytes_downloaded || 0;
  const progress = total > 0 ? Math.round((downloaded / total) * 100) : 0;
  const isProgressStatus = item.status === 'downloading' || item.status === 'retrying';
  const title = item.content?.title || (item.download_path ? filepathBase(item.download_path) : item.url);
  const year = item.content?.year ? `(${item.content.year})` : '';
  const typeIcon = item.content?.type === 'movies' ? '🎬' : item.content?.type === 'tvshows' ? '📺' : '🔗';
  const { label: statusLabel, badgeClass: statusBadgeClass } = getStatusInfo(item, t);

  return { item, title, year, typeIcon, statusLabel, statusBadgeClass, isProgressStatus, progress };
}

export function DownloadsSummaryList({ downloads, loading, onRowClick }: DownloadsSummaryListProps) {
  const { t } = useTranslation('downloads');
  const isMobile = useIsMobile();

  const rows = downloads.map(item => buildRow(item, t));

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
