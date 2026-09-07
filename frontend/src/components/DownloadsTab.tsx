import React, { useEffect, useState } from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import * as Dialog from '@radix-ui/react-dialog';
import * as Progress from '@radix-ui/react-progress';
import { useTranslation } from 'react-i18next';
import { DownloadEnriched } from '../types';
import { formatDate } from '../utils/date';
import { DownloadsSummaryList } from './DownloadsSummaryList';
import { Pagination } from './Pagination';
import { useIsMobile } from '../hooks/useMediaQuery';

interface DownloadsTabProps {
  downloads: DownloadEnriched[];
  downloadsLoading: boolean;
  statusFilter: string;
  setStatusFilter: (status: string) => void;
  typeFilter: string;
  setTypeFilter: (type: string) => void;
  problemFilter: string;
  setProblemFilter: (problem: string) => void;
  downloadsTotal: number;
  downloadsPage: number;
  setDownloadsPage: React.Dispatch<React.SetStateAction<number>>;
  downloadsLimit: number;
  setDownloadsLimit: (limit: number) => void;
  onFetchDownloads: () => void;
  onOpenMoveDialog: (item: DownloadEnriched) => void;
  onOpenRenameDialog: (item: DownloadEnriched) => void;
  onResyncPath: (item: DownloadEnriched) => Promise<void>;
  onCancelDownload: (item: DownloadEnriched) => Promise<void>;
}

function filepathBase(path: string) {
  const parts = path.split('/');
  return parts[parts.length - 1];
}

export function DownloadsTab({
  downloads,
  downloadsLoading,
  statusFilter,
  setStatusFilter,
  typeFilter,
  setTypeFilter,
  problemFilter,
  setProblemFilter,
  downloadsTotal,
  downloadsPage,
  setDownloadsPage,
  downloadsLimit,
  setDownloadsLimit,
  onFetchDownloads,
  onOpenMoveDialog,
  onOpenRenameDialog,
  onResyncPath,
  onCancelDownload
}: DownloadsTabProps) {
  const { t, i18n } = useTranslation('downloads');
  const isMobile = useIsMobile();
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const selectedItem = downloads.find(d => d.id === selectedId) ?? null;
  const [isResyncing, setIsResyncing] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);

  const handleResyncClick = () => {
    if (!selectedItem) return;
    setIsResyncing(true);
    onResyncPath(selectedItem).finally(() => setIsResyncing(false));
  };

  const handleCancelClick = () => {
    if (!selectedItem) return;
    setIsCancelling(true);
    onCancelDownload(selectedItem).finally(() => setIsCancelling(false));
  };

  useEffect(() => {
    if (selectedId !== null && !selectedItem) {
      setSelectedId(null);
    }
  }, [selectedId, selectedItem]);

  // Derived data for the drawer, computed only when a download is selected.
  let drawerData: {
    total: number;
    downloaded: number;
    progress: number;
    isCompleted: boolean;
    isCancelEligible: boolean;
    isProgressStatus: boolean;
    title: string;
    year: string;
    typeIcon: string;
    statusLabel: string;
    statusBadgeClass: string;
    hasYearIssue: boolean | undefined;
    hasYearMismatch: boolean | undefined;
    hasFormatIssue: boolean | undefined;
    isLowQuality: boolean;
  } | null = null;

  if (selectedItem) {
    const total = selectedItem.total_bytes || 0;
    const downloaded = selectedItem.bytes_downloaded || 0;
    const progress = total > 0 ? Math.round((downloaded / total) * 100) : 0;
    const isCompleted = selectedItem.status === 'completed';
    const isCancelEligible = selectedItem.status === 'pending' || selectedItem.status === 'failed' || selectedItem.status === 'retrying';
    const isProgressStatus = selectedItem.status === 'downloading' || selectedItem.status === 'retrying';

    const title = selectedItem.content?.title || (selectedItem.download_path ? filepathBase(selectedItem.download_path) : selectedItem.url);
    const year = selectedItem.content?.year ? `(${selectedItem.content.year})` : '';
    const typeIcon = selectedItem.content?.type === 'movies' ? '🎬' : selectedItem.content?.type === 'tvshows' ? '📺' : '🔗';

    let statusLabel: string = selectedItem.status;
    let statusBadgeClass = 'badge-pending';
    if (selectedItem.status === 'completed') {
      statusLabel = isMobile ? t('status.completed') : `${t('status.completed')} ${t('status.completedText')}`;
      statusBadgeClass = 'badge-success';
    } else if (selectedItem.status === 'downloading') {
      statusLabel = isMobile ? t('status.downloading') : `${t('status.downloading')} ${t('status.downloadingText')}`;
      statusBadgeClass = 'badge-progress';
    } else if (selectedItem.status === 'failed') {
      if (selectedItem.retry_count > 0) {
        statusLabel = isMobile
          ? t('status.failedWithRetry', { count: selectedItem.retry_count })
          : `${t('status.failed')} ${t('status.failedText')} (${selectedItem.retry_count}×)`;
      } else {
        statusLabel = isMobile ? t('status.failed') : `${t('status.failed')} ${t('status.failedText')}`;
      }
      statusBadgeClass = 'badge-failed';
    } else if (selectedItem.status === 'pending') {
      statusLabel = isMobile ? t('status.pending') : `${t('status.pending')} ${t('status.pendingText')}`;
      statusBadgeClass = 'badge-pending';
    } else if (selectedItem.status === 'retrying') {
      statusLabel = t('status.retrying');
      statusBadgeClass = 'badge-pending';
    } else if (selectedItem.status === 'cancelled') {
      statusLabel = t('status.cancelled');
      statusBadgeClass = 'badge-neutral';
    }

    const hasYearIssue = selectedItem.file_info && !selectedItem.file_info.has_year_in_path;
    const hasYearMismatch = selectedItem.file_info?.year_mismatch;
    const hasFormatIssue = selectedItem.file_info && !selectedItem.file_info.is_valid_format;
    const resolutionLower = selectedItem.file_info?.detected_resolution?.toLowerCase() || '';
    const isLowQuality = resolutionLower === '480p' || resolutionLower === '360p';

    drawerData = {
      total,
      downloaded,
      progress,
      isCompleted,
      isCancelEligible,
      isProgressStatus,
      title,
      year,
      typeIcon,
      statusLabel,
      statusBadgeClass,
      hasYearIssue,
      hasYearMismatch,
      hasFormatIssue,
      isLowQuality,
    };
  }

  return (
    <Tabs.Content value="downloads" className="card tab-panel">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>{t('heading')}</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('subtitle')}</p>
        </div>
        <button onClick={onFetchDownloads} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>{t('refresh')}</button>
      </div>

      {/* Filter Bar */}
      <div className="filters" style={{ display: 'flex', gap: '0.75rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
        <select
          value={statusFilter}
          onChange={e => setStatusFilter(e.target.value)}
          style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: '#fff', fontWeight: 600, color: 'var(--text-secondary)' }}
        >
          <option value="">{t('statusFilter.all')}</option>
          <option value="completed">{t('statusFilter.completed')}</option>
          <option value="downloading">{t('statusFilter.downloading')}</option>
          <option value="failed">{t('statusFilter.failed')}</option>
          <option value="cancelled">{t('statusFilter.cancelled')}</option>
        </select>

        <select
          value={typeFilter}
          onChange={e => setTypeFilter(e.target.value)}
          style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: '#fff', fontWeight: 600, color: 'var(--text-secondary)' }}
        >
          <option value="">{t('typeFilter.all')}</option>
          <option value="movies">{t('typeFilter.movies')}</option>
          <option value="tvshows">{t('typeFilter.tvshows')}</option>
        </select>

        <select
          value={problemFilter}
          onChange={e => setProblemFilter(e.target.value)}
          style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: '#fff', fontWeight: 600, color: 'var(--text-secondary)' }}
        >
          <option value="">{t('problemFilter.none')}</option>
          <option value="missing_year">{t('problemFilter.missingYear')}</option>
          <option value="year_mismatch">{t('problemFilter.yearMismatch')}</option>
          <option value="unknown_format">{t('problemFilter.unknownFormat')}</option>
          <option value="low_quality">{t('problemFilter.lowQuality')}</option>
        </select>
      </div>

      <DownloadsSummaryList
        downloads={downloads}
        loading={downloadsLoading}
        onRowClick={item => setSelectedId(item.id)}
      />

      <Pagination
        total={downloadsTotal}
        page={downloadsPage}
        setPage={setDownloadsPage}
        limit={downloadsLimit}
        setLimit={setDownloadsLimit}
        limitOptions={[10, 50, 100]}
      />

      {/* Sidepanel de Détails Interactif (Drawer) */}
      <Dialog.Root open={!!selectedItem} onOpenChange={(open) => !open && setSelectedId(null)}>
        <Dialog.Portal>
          <Dialog.Overlay className="drawer-overlay" />
          <Dialog.Content className="drawer-content">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1rem' }}>
              <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)', margin: 0, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                {t('drawer.title')}
              </Dialog.Title>
              <Dialog.Close className="btn-secondary" style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem', borderRadius: 'var(--radius-sm)' }}>
                {t('drawer.close')}
              </Dialog.Close>
            </div>

            <Dialog.Description style={{ display: 'none' }}>
              {t('drawer.description')}
            </Dialog.Description>

            {selectedItem && drawerData && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', flex: 1, paddingBottom: '1rem' }}>

                {/* Title */}
                <h3 style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--primary-slate)', display: 'flex', alignItems: 'center', gap: '0.5rem', margin: 0 }}>
                  <span>{drawerData.typeIcon}</span>
                  <span>{drawerData.title} {drawerData.year}</span>
                </h3>

                {/* Section: Status & Progress */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <h3 className="drawer-section-title">{t('drawer.statusSection')}</h3>
                  <div>
                    <span className={`badge ${drawerData.statusBadgeClass}`} style={{ fontSize: '0.8rem', padding: '0.25rem 0.6rem', fontWeight: 600 }}>
                      {drawerData.statusLabel}
                    </span>
                  </div>
                  {drawerData.isProgressStatus && (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>
                        <span>{t('progress', { percent: drawerData.progress })}</span>
                        {selectedItem.file_size && (
                          <span>{t('progressSize', {
                            downloaded: (drawerData.downloaded / 1024 / 1024).toFixed(1),
                            total: (selectedItem.file_size / 1024 / 1024).toFixed(1),
                          })}</span>
                        )}
                      </div>
                      <Progress.Root value={drawerData.progress} className="progress-root">
                        <Progress.Indicator className="progress-indicator" style={{ width: `${drawerData.progress}%` }} />
                      </Progress.Root>
                    </div>
                  )}
                </div>

                {/* Section: File */}
                {(selectedItem.file_info || selectedItem.download_path || selectedItem.url) && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                    <h3 className="drawer-section-title">{t('drawer.fileSection')}</h3>
                    {selectedItem.file_info && (
                      <div className="file-info" style={{ backgroundColor: 'var(--bg-app)', padding: '0.5rem 0.75rem', borderRadius: 'var(--radius-sm)', fontFamily: "'Courier New', monospace", fontSize: '0.85rem', lineHeight: '1.4', border: '1px solid var(--border-color)' }}>
                        <div style={{ fontWeight: 600, color: 'var(--text-secondary)' }}>{t('folder', { name: selectedItem.file_info.folder_name })}</div>
                        {selectedItem.file_info.file_name && (
                          <div style={{ color: 'var(--text-muted)', paddingLeft: '1rem' }}>└─ {selectedItem.file_info.file_name}</div>
                        )}
                      </div>
                    )}
                    {selectedItem.status === 'completed' && selectedItem.download_path ? (
                      <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', wordBreak: 'break-all', fontWeight: 500, margin: 0 }}>
                        {selectedItem.download_path}
                      </p>
                    ) : selectedItem.target_path || selectedItem.staging_path ? (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                        {selectedItem.target_path && (
                          <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', wordBreak: 'break-all', fontWeight: 500, margin: 0 }}>
                            {t('plannedPath', { path: selectedItem.target_path })}
                          </p>
                        )}
                        {selectedItem.staging_path && (
                          <p style={{ color: 'var(--text-muted)', fontSize: '0.8rem', wordBreak: 'break-all', fontWeight: 500, margin: 0 }}>
                            {t('stagingPath', { path: selectedItem.staging_path })}
                          </p>
                        )}
                      </div>
                    ) : (
                      <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', wordBreak: 'break-all', fontWeight: 500, margin: 0 }}>
                        {selectedItem.url}
                      </p>
                    )}
                  </div>
                )}

                {/* Section: Technical Specifications */}
                {selectedItem.file_info && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                    <h3 className="drawer-section-title">{t('drawer.technicalSection')}</h3>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', alignItems: 'center', color: 'var(--primary-slate)', fontWeight: 500, fontSize: '0.85rem' }}>
                      <span>{t('format', { format: selectedItem.file_info.extension.toUpperCase() || t('unknownFormat') })}</span>
                      <span>•</span>
                      <span>{selectedItem.file_info.detected_resolution || t('unknownResolution')}</span>
                      {selectedItem.file_size && (
                        <>
                          <span>•</span>
                          <span>{t('sizeMb', { size: (selectedItem.file_size / 1024 / 1024).toFixed(1) })}</span>
                        </>
                      )}
                      {selectedItem.content?.duration && (
                        <>
                          <span>•</span>
                          <span>{t('durationMin', { count: selectedItem.content.duration })}</span>
                        </>
                      )}
                      {selectedItem.completed_at && (
                        <>
                          <span>•</span>
                          <span>{t('completedAt', { date: formatDate(selectedItem.completed_at, i18n.language) })}</span>
                        </>
                      )}
                    </div>
                  </div>
                )}

                {/* Section: Validation */}
                {selectedItem.file_info && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                    <h3 className="drawer-section-title">{t('drawer.validationSection')}</h3>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.35rem', alignItems: 'center' }}>
                      {drawerData.isLowQuality && (
                        <span className="badge badge-pending" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          {t('lowQualityBadge', { resolution: selectedItem.file_info.detected_resolution })}
                        </span>
                      )}

                      {drawerData.hasYearIssue ? (
                        <span className="badge badge-failed" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          {t('missingYearBadge')}
                        </span>
                      ) : drawerData.hasYearMismatch ? (
                        <span className="badge badge-failed" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }} title={t('incorrectYearTitle')}>
                          {t('incorrectYearBadge')}
                        </span>
                      ) : (
                        <span className="badge badge-success" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          {t('yearOkBadge')}
                        </span>
                      )}

                      {drawerData.hasFormatIssue ? (
                        <span className="badge badge-failed" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          {t('unknownFormatBadge')}
                        </span>
                      ) : (
                        <span className="badge badge-success" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          {t('formatOkBadge')}
                        </span>
                      )}
                    </div>
                  </div>
                )}

                {/* Section: Genres */}
                {selectedItem.content?.genres && (
                  <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '0.35rem', fontWeight: 500 }}>
                    <span>{t('genres')}</span>
                    <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>{selectedItem.content.genres}</span>
                  </div>
                )}

                {/* Section: Error */}
                {selectedItem.status === 'failed' && selectedItem.error_message && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                    <h3 className="drawer-section-title">{t('drawer.errorSection')}</h3>
                    <div style={{ padding: '0.75rem', backgroundColor: 'var(--status-failed-bg)', borderRadius: 'var(--radius-sm)', color: 'var(--status-failed-text)', fontSize: '0.8rem', fontWeight: 600, border: '1px solid var(--status-failed-border)' }}>
                      🔴 {selectedItem.error_message}
                    </div>
                  </div>
                )}

                {/* Section: Actions */}
                {(drawerData.isCompleted || drawerData.isCancelEligible) && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                    <h3 className="drawer-section-title">{t('drawer.actionsSection')}</h3>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      {drawerData.isCompleted && (
                        <>
                          <button onClick={() => onOpenMoveDialog(selectedItem)} className="btn-primary" style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }}>
                            {t('move')}
                          </button>
                          <button onClick={() => onOpenRenameDialog(selectedItem)} className="btn-secondary" style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }}>
                            {t('rename')}
                          </button>
                          <button
                            onClick={handleResyncClick}
                            disabled={isResyncing}
                            className="btn-secondary"
                            style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }}
                          >
                            {isResyncing ? t('resync.inProgress') : t('resync.action')}
                          </button>
                        </>
                      )}
                      {drawerData.isCancelEligible && (
                        <button
                          onClick={handleCancelClick}
                          disabled={isCancelling}
                          className="btn-secondary"
                          style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }}
                        >
                          {isCancelling ? t('cancel.inProgress') : t('cancel.action')}
                        </button>
                      )}
                    </div>
                  </div>
                )}

              </div>
            )}
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </Tabs.Content>
  );
}
