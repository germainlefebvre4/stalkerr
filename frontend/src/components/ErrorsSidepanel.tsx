import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { DownloadEnriched } from '../types';
import { formatDate } from '../utils/date';
import { getErrorReasons } from './ErrorsTable';

interface ErrorsSidepanelProps {
  item: DownloadEnriched | null;
  onOpenChange: (open: boolean) => void;
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

export function ErrorsSidepanel({ item, onOpenChange }: ErrorsSidepanelProps) {
  const { t, i18n } = useTranslation('errors');

  const reasons = item ? getErrorReasons(item, t) : [];
  const hasYearReason = reasons.some(r => r.key === 'missing_year' || r.key === 'year_mismatch');
  const title = item ? (item.content?.title || item.file_info?.file_name || item.file_info?.folder_name || item.url) : '';
  const year = item?.content?.year ? `(${item.content.year})` : '';
  const typeIcon = item?.content?.type === 'movies' ? '🎬' : item?.content?.type === 'tvshows' ? '📺' : '🔗';

  return (
    <Dialog.Root open={!!item} onOpenChange={onOpenChange}>
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
