import { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { DownloadEnriched, DuplicateCandidate, KeepDuplicateResponse, ResolveDuplicateOccurrence } from '../types';
import { formatDate } from '../utils/date';
import { getErrorReasons } from '../utils/errorReasons';
import { api } from '../services/api';

interface ErrorsSidepanelProps {
  item: DownloadEnriched | null;
  onOpenChange: (open: boolean) => void;
  /** Called when the panel closes after a keep-one run, so the list can refresh. */
  onChanged?: () => void;
}

type KeepStep = 'idle' | 'loading' | 'picking' | 'confirming' | 'running' | 'done';

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  const units = ['KB', 'MB', 'GB', 'TB'];
  let value = bytes / 1024;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024;
    i++;
  }
  return `${value.toFixed(1)} ${units[i]}`;
}

function getStatusLabel(item: DownloadEnriched, t: (key: string, options?: Record<string, unknown>) => string): string {
  switch (item.status) {
    case 'completed':
      return t('status.completed');
    case 'downloading':
      return t('status.downloading');
    case 'failed':
      return item.retry_count > 0 ? t('status.failedWithRetry', { count: item.retry_count }) : t('status.failed');
    case 'retrying':
      return t('status.retrying');
    default:
      return t('status.pending');
  }
}

export function ErrorsSidepanel({ item, onOpenChange, onChanged }: ErrorsSidepanelProps) {
  const { t, i18n } = useTranslation('errors');

  const [resolving, setResolving] = useState(false);
  const [resolveOutcome, setResolveOutcome] = useState<ResolveDuplicateOccurrence[] | null>(null);
  const [resolveError, setResolveError] = useState<string | null>(null);

  const [keepStep, setKeepStep] = useState<KeepStep>('idle');
  const [candidates, setCandidates] = useState<DuplicateCandidate[]>([]);
  const [selectedPath, setSelectedPath] = useState<string | null>(null);
  const [keepResult, setKeepResult] = useState<KeepDuplicateResponse | null>(null);
  const [keepError, setKeepError] = useState<string | null>(null);
  const [keepRan, setKeepRan] = useState(false);

  // Reset the resolve-action state when a different item is opened, following
  // React's "adjusting state when a prop changes" pattern (setState during
  // render, not in an effect) rather than an effect that would run after the
  // stale outcome has already painted once.
  const [lastItemId, setLastItemId] = useState(item?.id);
  if (item?.id !== lastItemId) {
    setLastItemId(item?.id);
    setResolveOutcome(null);
    setResolveError(null);
    setResolving(false);
    setKeepStep('idle');
    setCandidates([]);
    setSelectedPath(null);
    setKeepResult(null);
    setKeepError(null);
    setKeepRan(false);
  }

  const reasons = item ? getErrorReasons(item, t) : [];
  const hasYearReason = reasons.some(r => r.key === 'missing_year' || r.key === 'year_mismatch');
  const isDuplicate = reasons.some(r => r.key === 'duplicate_download');
  const title = item ? (item.content?.title || item.file_info?.file_name || item.file_info?.folder_name || item.url) : '';
  const year = item?.content?.year ? `(${item.content.year})` : '';
  const typeIcon = item?.content?.type === 'movies' ? '🎬' : item?.content?.type === 'tvshows' ? '📺' : '🔗';

  const handleResolveDuplicate = () => {
    if (!item) return;
    setResolving(true);
    setResolveError(null);
    api.resolveDuplicateDownload(item.id)
      .then(res => setResolveOutcome(res.occurrences))
      .catch(() => setResolveError(t('duplicate.resolveError')))
      .finally(() => setResolving(false));
  };

  const handleOpenChange = (open: boolean) => {
    if (!open && keepRan) {
      setKeepRan(false);
      onChanged?.();
    }
    onOpenChange(open);
  };

  const handleOpenKeepPicker = () => {
    if (!item) return;
    setKeepStep('loading');
    setKeepError(null);
    api.getDuplicateCandidates(item.id)
      .then(res => {
        setCandidates(res.candidates);
        setSelectedPath(res.candidates.find(c => c.suggested && c.exists)?.download_path ?? null);
        setKeepStep('picking');
      })
      .catch(() => {
        setKeepError(t('duplicate.keepLoadError'));
        setKeepStep('idle');
      });
  };

  const selectedCandidate = candidates.find(c => c.download_path === selectedPath) ?? null;
  const deletedCandidates = candidates.filter(c => c.exists && c.download_path !== selectedPath);

  const handleConfirmKeep = () => {
    if (!selectedCandidate) return;
    setKeepStep('running');
    setKeepError(null);
    api.keepDuplicateDownload(selectedCandidate.download_info_ids[0])
      .then(res => {
        setKeepRan(true);
        setKeepResult(res);
        setKeepStep('done');
      })
      .catch(() => {
        setKeepRan(true); // a partial run may have changed data
        setKeepError(t('duplicate.keepError'));
        setKeepStep('confirming');
      });
  };

  const keepOutcomeLabel = (outcome: string) => {
    switch (outcome) {
      case 'kept': return t('duplicate.outcomeKept');
      case 'deleted': return t('duplicate.outcomeDeleted');
      default: return t('duplicate.outcomeCancelledFileMissing');
    }
  };

  const refreshLabel = (refresh: KeepDuplicateResponse['refresh']) => {
    switch (refresh) {
      case 'requested': return t('duplicate.refreshRequested');
      case 'not_found_in_library': return t('duplicate.refreshNotFound');
      default: return t('duplicate.refreshNotRefreshed');
    }
  };

  return (
    <Dialog.Root open={!!item} onOpenChange={handleOpenChange}>
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

          {item && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', flex: 1, paddingBottom: '1rem' }}>

              <h3 style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--primary-slate)', display: 'flex', alignItems: 'center', gap: '0.5rem', margin: 0 }}>
                <span>{typeIcon}</span>
                <span>{title} {year}</span>
              </h3>

              {/* Section: Diagnostic (leads, per this tab's diagnostic-first hierarchy) */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                <h3 className="drawer-section-title">{t('drawer.diagnosticSection')}</h3>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.35rem', alignItems: 'center' }}>
                  {reasons.map(reason => (
                    <span key={reason.key} className="badge badge-failed" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                      {reason.label}
                    </span>
                  ))}
                </div>
                {hasYearReason && (
                  <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                    <div>{t('year.detected', { year: item.file_info?.detected_year ?? t('year.unknown') })}</div>
                    <div>{t('year.expected', { year: item.content?.year ?? t('year.unknown') })}</div>
                  </div>
                )}
                {reasons.some(r => r.key === 'unknown_format') && (
                  <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                    {item.file_info?.extension || t('unknownExtension')}
                  </div>
                )}
                {isDuplicate && item.duplicate_siblings && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    <div style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-secondary)' }}>
                      {t('duplicate.siblingsTitle')}
                    </div>
                    {item.duplicate_siblings.map(sibling => (
                      <div
                        key={sibling.download_info_id}
                        className="file-info"
                        style={{ backgroundColor: 'var(--bg-app)', padding: '0.5rem 0.75rem', borderRadius: 'var(--radius-sm)', fontFamily: "'Courier New', monospace", fontSize: '0.8rem', lineHeight: '1.4', border: '1px solid var(--border-color)' }}
                      >
                        <div style={{ wordBreak: 'break-all' }}>{sibling.download_path}</div>
                        {sibling.completed_at && (
                          <div style={{ color: 'var(--text-muted)' }}>{t('completedAt', { date: formatDate(sibling.completed_at, i18n.language) })}</div>
                        )}
                      </div>
                    ))}

                    <div>
                      <button
                        onClick={handleResolveDuplicate}
                        disabled={resolving}
                        className="btn-secondary"
                        style={{ fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
                      >
                        {resolving ? t('duplicate.resolveInProgress') : t('duplicate.resolveAction')}
                      </button>
                    </div>

                    <div>
                      {keepStep === 'idle' && (
                        <button
                          onClick={handleOpenKeepPicker}
                          className="btn-secondary"
                          style={{ fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
                        >
                          {t('duplicate.keepAction')}
                        </button>
                      )}
                      {keepStep === 'loading' && (
                        <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{t('duplicate.keepLoading')}</div>
                      )}
                    </div>

                    {keepError && (
                      <div style={{ fontSize: '0.8rem', color: 'var(--error-color, #dc2626)', fontWeight: 500 }}>{keepError}</div>
                    )}

                    {keepStep === 'picking' && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                        <div style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('duplicate.keepPickerTitle')}</div>
                        {candidates.map(cand => (
                          <label
                            key={cand.download_path}
                            style={{ display: 'flex', gap: '0.5rem', alignItems: 'flex-start', fontSize: '0.8rem', opacity: cand.exists ? 1 : 0.6 }}
                          >
                            <input
                              type="radio"
                              name="keep-duplicate-file"
                              checked={selectedPath === cand.download_path}
                              disabled={!cand.exists}
                              onChange={() => setSelectedPath(cand.download_path)}
                            />
                            <span style={{ display: 'flex', flexDirection: 'column', gap: '0.15rem' }}>
                              <span style={{ wordBreak: 'break-all', fontFamily: "'Courier New', monospace" }}>{cand.download_path}</span>
                              <span style={{ color: 'var(--text-muted)' }}>
                                {cand.exists
                                  ? (cand.size_bytes != null ? t('duplicate.keepSize', { size: formatSize(cand.size_bytes) }) : '')
                                  : t('duplicate.keepMissing')}
                                {cand.completed_at && ` · ${t('completedAt', { date: formatDate(cand.completed_at, i18n.language) })}`}
                              </span>
                              <span style={{ display: 'flex', gap: '0.35rem' }}>
                                {cand.suggested && <span className="badge badge-pending">{t('duplicate.keepSuggested')}</span>}
                                {cand.matches_library_name && <span className="badge badge-pending">{t('duplicate.keepMatchesName')}</span>}
                              </span>
                            </span>
                          </label>
                        ))}
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                          <button
                            onClick={() => setKeepStep('confirming')}
                            disabled={!selectedCandidate}
                            className="btn-secondary"
                            style={{ fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
                          >
                            {t('duplicate.keepContinue')}
                          </button>
                          <button
                            onClick={() => setKeepStep('idle')}
                            className="btn-secondary"
                            style={{ fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
                          >
                            {t('duplicate.keepCancel')}
                          </button>
                        </div>
                      </div>
                    )}

                    {(keepStep === 'confirming' || keepStep === 'running') && selectedCandidate && (
                      <div role="alertdialog" style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', padding: '0.75rem' }}>
                        <div style={{ fontSize: '0.85rem', fontWeight: 700 }}>{t('duplicate.keepConfirmTitle')}</div>
                        <div style={{ fontSize: '0.8rem' }}>
                          <div style={{ fontWeight: 700 }}>{t('duplicate.keepConfirmKeep')}</div>
                          <div style={{ wordBreak: 'break-all', fontFamily: "'Courier New', monospace" }}>{selectedCandidate.download_path}</div>
                        </div>
                        <div style={{ fontSize: '0.8rem' }}>
                          {deletedCandidates.length > 0 ? (
                            <>
                              <div style={{ fontWeight: 700 }}>{t('duplicate.keepConfirmDelete')}</div>
                              {deletedCandidates.map(c => (
                                <div key={c.download_path} style={{ wordBreak: 'break-all', fontFamily: "'Courier New', monospace" }}>{c.download_path}</div>
                              ))}
                            </>
                          ) : (
                            <div>{t('duplicate.keepConfirmNone')}</div>
                          )}
                        </div>
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                          <button
                            onClick={handleConfirmKeep}
                            disabled={keepStep === 'running'}
                            className="btn-secondary"
                            style={{ fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
                          >
                            {keepStep === 'running' ? t('duplicate.keepInProgress') : t('duplicate.keepConfirmAction')}
                          </button>
                          <button
                            onClick={() => setKeepStep('picking')}
                            disabled={keepStep === 'running'}
                            className="btn-secondary"
                            style={{ fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
                          >
                            {t('duplicate.keepCancel')}
                          </button>
                        </div>
                      </div>
                    )}

                    {keepStep === 'done' && keepResult && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                        {keepResult.occurrences.map(occ => (
                          <div key={occ.download_info_id} style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                            <span style={{ wordBreak: 'break-all' }}>{occ.download_path}</span>
                            {': '}
                            <span style={{ fontWeight: 700 }}>{keepOutcomeLabel(occ.outcome)}</span>
                          </div>
                        ))}
                        <div
                          style={{ fontSize: '0.8rem', fontWeight: 600, color: keepResult.refresh === 'requested' ? 'var(--text-secondary)' : 'var(--error-color, #dc2626)' }}
                        >
                          {refreshLabel(keepResult.refresh)}
                        </div>
                      </div>
                    )}

                    {resolveError && (
                      <div style={{ fontSize: '0.8rem', color: 'var(--error-color, #dc2626)', fontWeight: 500 }}>{resolveError}</div>
                    )}

                    {resolveOutcome && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                        {resolveOutcome.map(occ => (
                          <div key={occ.download_info_id} style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                            <span style={{ wordBreak: 'break-all' }}>{occ.download_path}</span>
                            {': '}
                            <span style={{ fontWeight: 700 }}>
                              {occ.outcome === 'cancelled' ? t('duplicate.outcomeCancelled') : t('duplicate.outcomeUnchanged')}
                            </span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>

              {/* Section: File */}
              {item.file_info && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <h3 className="drawer-section-title">{t('drawer.fileSection')}</h3>
                  <div className="file-info" style={{ backgroundColor: 'var(--bg-app)', padding: '0.5rem 0.75rem', borderRadius: 'var(--radius-sm)', fontFamily: "'Courier New', monospace", fontSize: '0.85rem', lineHeight: '1.4', border: '1px solid var(--border-color)' }}>
                    <div style={{ fontWeight: 600, color: 'var(--text-secondary)' }}>{t('folder', { name: item.file_info.folder_name })}</div>
                    {item.file_info.file_name && (
                      <div style={{ color: 'var(--text-muted)', paddingLeft: '1rem' }}>└─ {item.file_info.file_name}</div>
                    )}
                  </div>
                </div>
              )}

              {/* Section: Matched Content */}
              {item.content && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <h3 className="drawer-section-title">{t('drawer.contentSection')}</h3>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', alignItems: 'center', color: 'var(--primary-slate)', fontWeight: 500, fontSize: '0.85rem' }}>
                    <span>{typeIcon} {item.content.title} {year}</span>
                    {item.content.season != null && item.content.episode != null && (
                      <>
                        <span>•</span>
                        <span>{t('season', { season: item.content.season, episode: item.content.episode })}</span>
                      </>
                    )}
                  </div>
                  {item.content.genres && (
                    <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '0.35rem', fontWeight: 500 }}>
                      <span>{t('genres')}</span>
                      <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>{item.content.genres}</span>
                    </div>
                  )}
                </div>
              )}

              {/* Section: Status & Dates */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                <h3 className="drawer-section-title">{t('drawer.statusSection')}</h3>
                <div>
                  <span className="badge badge-pending" style={{ fontSize: '0.8rem', padding: '0.25rem 0.6rem', fontWeight: 600 }}>
                    {getStatusLabel(item, t)}
                  </span>
                </div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                  {item.completed_at && <span>{t('completedAt', { date: formatDate(item.completed_at, i18n.language) })}</span>}
                  {item.updated_at && <span>{t('updatedAt', { date: formatDate(item.updated_at, i18n.language) })}</span>}
                </div>
              </div>

              {/* Section: Source */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                <h3 className="drawer-section-title">{t('drawer.sourceSection')}</h3>
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', wordBreak: 'break-all', fontWeight: 500, margin: 0 }}>
                  {item.url}
                </p>
              </div>

            </div>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
