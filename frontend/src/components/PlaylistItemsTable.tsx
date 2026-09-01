import React from 'react';
import { useTranslation } from 'react-i18next';
import { PlaylistItem } from '../types';
import { formatDate, getDateGroupLabel, getDateGroupStarts } from '../utils/date';
import {
  getProcessingStatus,
  getDownloadStatus,
  getProcessingStatusBadgeClass,
  getDownloadStatusBadgeClass,
} from '../utils/pipelineState';
import { useIsMobile } from '../hooks/useMediaQuery';

interface PlaylistItemsTableProps {
  items: PlaylistItem[];
  loading: boolean;
  onOpenOverride: (item: PlaylistItem) => void;
  onResetPipeline: (id: number, contentType: string) => void;
  onRowClick?: (item: PlaylistItem) => void;
  /** Defaults to true: renders "Aujourd'hui"/"Hier"/date headers above created_at day boundaries. */
  showDateGroups?: boolean;
  sort?: string;
  order?: string;
  onSort?: (column: string) => void;
}

export function PlaylistItemsTable({
  items,
  loading,
  onOpenOverride,
  onResetPipeline,
  onRowClick,
  showDateGroups = true,
  sort,
  order,
  onSort,
}: PlaylistItemsTableProps) {
  const { t, i18n } = useTranslation('playlist');
  const isMobile = useIsMobile();

  const dateGroupStarts = showDateGroups && (!sort || sort === 'created_at') ? getDateGroupStarts(items) : [];

  const renderSortableHeader = (column: string, label: string, style?: React.CSSProperties) => {
    if (!onSort) {
      return <th style={style}>{label}</th>;
    }
    const isActive = sort === column;
    return (
      <th
        onClick={() => onSort(column)}
        style={{ cursor: 'pointer', userSelect: 'none', ...style }}
        title={t('table.sortBy', { column: label })}
      >
        {label}
        {isActive && (
          <span style={{ marginLeft: '0.35rem', display: 'inline-block' }}>
            {order === 'asc' ? '▲' : '▼'}
          </span>
        )}
      </th>
    );
  };

  if (isMobile) {
    return (
      <div>
        {loading ? (
          <div className="mobile-list-empty">{t('table.loading')}</div>
        ) : items.length === 0 ? (
          <div className="mobile-list-empty">{t('table.empty')}</div>
        ) : (
          items.map((item, index) => {
            const processingStatus = getProcessingStatus(item.state);
            const downloadStatus = getDownloadStatus(item.state);
            const processingLabel = t(`stateFilter.${processingStatus}`);
            const downloadLabel = downloadStatus === 'not_downloaded'
              ? t('pipelineStatus.notDownloaded')
              : t(`stateFilter.${downloadStatus}`);
            return (
              <React.Fragment key={item.id}>
                {dateGroupStarts[index] && (
                  <div className="date-group-header">
                    {getDateGroupLabel(new Date(item.created_at), t, i18n.language)}
                  </div>
                )}
                <div className="mobile-list-card" onClick={() => onRowClick?.(item)}>
                  <div className="mobile-list-card-main">
                    <span className="mobile-list-card-title">{item.tvg_name}</span>
                    <span className="mobile-list-card-subtitle">{item.group_title}</span>
                  </div>
                  <div style={{ display: 'flex', gap: '0.4rem', alignItems: 'center' }}>
                    <span
                      className={`status-dot ${getProcessingStatusBadgeClass(processingStatus)}`}
                      role="img"
                      title={processingLabel}
                      aria-label={processingLabel}
                    />
                    <span
                      className={`status-dot ${getDownloadStatusBadgeClass(downloadStatus)}`}
                      role="img"
                      title={downloadLabel}
                      aria-label={downloadLabel}
                    />
                  </div>
                </div>
              </React.Fragment>
            );
          })
        )}
      </div>
    );
  }

  return (
    <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
      <table className="custom-table">
        <thead>
          <tr>
            {renderSortableHeader('tvg_name', t('table.headers.mediaName'))}
            {renderSortableHeader('group_title', t('table.headers.groupCategory'))}
            {renderSortableHeader('tmdb_title', t('table.headers.tmdbEnrichment'))}
            {renderSortableHeader('state', t('table.headers.pipelineState'))}
            {renderSortableHeader('created_at', t('table.headers.createdAt'))}
            {renderSortableHeader('downloaded_at', t('table.headers.downloadedAt'))}
            <th style={{ textAlign: 'right' }}>{t('table.headers.actions')}</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={7} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
                <span style={{ fontWeight: 600 }}>{t('table.loading')}</span>
              </td>
            </tr>
          ) : items.length === 0 ? (
            <tr>
              <td colSpan={7} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('table.empty')}</td>
            </tr>
          ) : (
            items.map((item, index) => {
              const isMovie = item.content_type === 'movies';
              const tmdb = isMovie ? item.movie : item.tvshow;
              const processingStatus = getProcessingStatus(item.state);
              const downloadStatus = getDownloadStatus(item.state);
              const downloadLabel = downloadStatus === 'not_downloaded' ? t('pipelineStatus.notDownloaded') : downloadStatus;
              return (
                <React.Fragment key={item.id}>
                {dateGroupStarts[index] && (
                  <tr>
                    <td colSpan={7} className="date-group-header">
                      {getDateGroupLabel(new Date(item.created_at), t, i18n.language)}
                    </td>
                  </tr>
                )}
                <tr className={onRowClick ? 'clickable-row' : undefined} onClick={() => onRowClick?.(item)}>
                  <td style={{ fontWeight: 600, color: 'var(--primary-slate)' }}>{item.tvg_name}</td>
                  <td style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{item.group_title}</td>
                  <td>
                    {tmdb ? (
                      <div>
                        <strong style={{ color: 'var(--primary-accent)' }}>{tmdb.tmdb_title}</strong>{' '}
                        <span style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', fontWeight: 500 }}>({tmdb.tmdb_year})</span>
                      </div>
                    ) : (
                      <span style={{ fontStyle: 'italic', color: 'var(--text-muted)' }}>{t('table.notEnriched')}</span>
                    )}
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: '0.35rem', flexWrap: 'wrap' }}>
                      <span className={`badge ${getProcessingStatusBadgeClass(processingStatus)}`}>
                        {processingStatus}
                      </span>
                      <span className={`badge ${getDownloadStatusBadgeClass(downloadStatus)}`}>
                        {downloadLabel}
                      </span>
                    </div>
                  </td>
                  <td style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{formatDate(item.created_at, i18n.language)}</td>
                  <td style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>
                    {item.downloaded_at ? formatDate(item.downloaded_at, i18n.language) : '—'}
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    <div style={{ display: 'flex', gap: '0.35rem', justifyContent: 'flex-end' }}>
                      <button
                        onClick={(e) => { e.stopPropagation(); onOpenOverride(item); }}
                        className="btn-primary"
                        style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}
                        title={t('table.actions.correctTitle')}
                      >
                        {item.override_by ? t('table.actions.correct') : t('table.actions.associate')}
                      </button>
                      <button
                        onClick={(e) => { e.stopPropagation(); onResetPipeline(item.id, item.content_type); }}
                        className="btn-secondary"
                        style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}
                        title={t('table.actions.resetTitle')}
                      >
                        {t('table.actions.reset')}
                      </button>
                    </div>
                  </td>
                </tr>
                </React.Fragment>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
}
