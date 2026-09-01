import React from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { PlaylistItem } from '../types';
import { formatDate } from '../utils/date';
import {
  getProcessingStatus,
  getDownloadStatus,
  getProcessingStatusBadgeClass,
  getDownloadStatusBadgeClass,
} from '../utils/pipelineState';
import { api, ApiError } from '../services/api';

interface MediaOccurrenceDrawerBodyProps {
  item: PlaylistItem;
}

/**
 * The item-detail content (TMDB, pipeline, force-download, M3U provenance, raw line/URL).
 * Exported separately so it can be embedded without its own `Dialog.Root` (see mobile usage
 * in `RadarrSonarrTab`, where a second dialog/overlay must never be mounted).
 */
export function MediaOccurrenceDrawerBody({ item }: MediaOccurrenceDrawerBodyProps) {
  const { t, i18n } = useTranslation('playlist');
  const [copiedText, setCopiedText] = React.useState<'content' | 'url' | null>(null);
  const [forceDownloadStatus, setForceDownloadStatus] = React.useState<'idle' | 'loading' | 'queued' | 'error'>('idle');
  const [forceDownloadError, setForceDownloadError] = React.useState<string | null>(null);

  React.useEffect(() => {
    setForceDownloadStatus('idle');
    setForceDownloadError(null);
  }, [item.id]);

  const isForceDownloadEligible = !!(item.movie || item.tvshow)
    && item.state !== 'downloaded'
    && item.state !== 'downloading';

  const handleForceDownload = async () => {
    setForceDownloadStatus('loading');
    setForceDownloadError(null);
    try {
      await api.forceDownload(item.id);
      setForceDownloadStatus('queued');
    } catch (err) {
      setForceDownloadStatus('error');
      setForceDownloadError(err instanceof ApiError ? err.message : t('drawer.forceDownload.genericError'));
    }
  };

  const handleCopy = (text: string, type: 'content' | 'url') => {
    navigator.clipboard.writeText(text).then(() => {
      setCopiedText(type);
      setTimeout(() => setCopiedText(null), 2000);
    });
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', flex: 1, paddingBottom: '1rem' }}>

      {/* Section 1: Enrichissement Métadonnées (TMDB) */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <h3 className="drawer-section-title">
          {t('drawer.tmdbSection')}
        </h3>
        {(() => {
          const isMovie = item.content_type === 'movies';
          const tmdb = isMovie ? item.movie : item.tvshow;
          const legacySeasonEpisode = item as PlaylistItem & { season?: number; episode?: number };
          if (tmdb) {
            const posterUrl = tmdb.poster_path ? `https://image.tmdb.org/t/p/w342${tmdb.poster_path}` : null;
            const tmdbUrl = tmdb.tmdb_id ? `https://www.themoviedb.org/${isMovie ? 'movie' : 'tv'}/${tmdb.tmdb_id}` : null;
            const imdbUrl = tmdb.imdb_id ? `https://www.imdb.com/title/${tmdb.imdb_id}` : null;
            const tvdbUrl = tmdb.tvdb_id ? `https://thetvdb.com/?tab=series&id=${tmdb.tvdb_id}` : null;
            return (
              <div style={{ backgroundColor: 'rgba(99, 102, 241, 0.03)', border: '1px solid rgba(99, 102, 241, 0.1)', borderRadius: 'var(--radius-md)', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                <div style={{ display: 'flex', gap: '1rem' }}>
                  {posterUrl && (
                    <img
                      src={posterUrl}
                      alt={t('drawer.posterAlt', { title: tmdb.tmdb_title })}
                      style={{ width: '80px', height: 'auto', borderRadius: 'var(--radius-sm)', flexShrink: 0, objectFit: 'cover' }}
                    />
                  )}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: 'var(--primary-accent)' }}>
                      {tmdb.tmdb_title} <span style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', fontWeight: 600 }}>({tmdb.tmdb_year})</span>
                    </div>
                    {tmdb.genres && (
                      <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                        <strong>{t('drawer.genres')}</strong> {tmdb.genres}
                      </div>
                    )}
                    {isMovie && item.movie?.duration && (
                      <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                        <strong>{t('drawer.duration')}</strong> {t('drawer.durationMinutes', { count: item.movie.duration })}
                      </div>
                    )}
                    {!isMovie && (item.tvshow?.season !== undefined || legacySeasonEpisode.season !== undefined) && (
                      <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                        <strong>{t('drawer.position')}</strong> {t('drawer.seasonEpisode', {
                          season: item.tvshow?.season ?? legacySeasonEpisode.season,
                          episode: item.tvshow?.episode ?? legacySeasonEpisode.episode,
                        })}
                      </div>
                    )}
                    <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontWeight: 500 }}>
                      <strong>{t('drawer.tmdbId')}</strong> {tmdb.tmdb_id}
                    </div>
                  </div>
                </div>
                {tmdb.overview && (
                  <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
                    <strong>{t('drawer.synopsis')}</strong> {tmdb.overview}
                  </div>
                )}
                {(tmdbUrl || imdbUrl || tvdbUrl) && (
                  <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                    {tmdbUrl && (
                      <a href={tmdbUrl} target="_blank" rel="noopener noreferrer" className="btn-secondary" style={{ fontSize: '0.8rem', padding: '0.3rem 0.6rem', textDecoration: 'none' }}>
                        {t('drawer.viewOnTmdb')}
                      </a>
                    )}
                    {imdbUrl && (
                      <a href={imdbUrl} target="_blank" rel="noopener noreferrer" className="btn-secondary" style={{ fontSize: '0.8rem', padding: '0.3rem 0.6rem', textDecoration: 'none' }}>
                        {t('drawer.viewOnImdb')}
                      </a>
                    )}
                    {tvdbUrl && (
                      <a href={tvdbUrl} target="_blank" rel="noopener noreferrer" className="btn-secondary" style={{ fontSize: '0.8rem', padding: '0.3rem 0.6rem', textDecoration: 'none' }}>
                        {t('drawer.viewOnTvdb')}
                      </a>
                    )}
                  </div>
                )}
              </div>
            );
          } else {
            return (
              <div style={{ backgroundColor: 'var(--bg-app)', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)', padding: '1rem', textAlign: 'center', color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem' }}>
                {t('drawer.noTmdbAssociation', { type: item.content_type })}
              </div>
            );
          }
        })()}
      </div>

      {/* Section 2: État du Pipeline d'Ingestion */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <h3 className="drawer-section-title">
          {t('drawer.pipelineSection')}
        </h3>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem', fontSize: '0.85rem' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.processingStatus')}</span>
            <div>
              <span className={`badge ${getProcessingStatusBadgeClass(getProcessingStatus(item.state))}`} style={{ fontSize: '0.75rem' }}>
                {getProcessingStatus(item.state)}
              </span>
            </div>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.downloadStatus')}</span>
            <div>
              <span className={`badge ${getDownloadStatusBadgeClass(getDownloadStatus(item.state))}`} style={{ fontSize: '0.75rem' }}>
                {getDownloadStatus(item.state) === 'not_downloaded' ? t('pipelineStatus.notDownloaded') : getDownloadStatus(item.state)}
              </span>
            </div>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.importDate')}</span>
            <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{formatDate(item.created_at, i18n.language)}</span>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
            <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.downloadDate')}</span>
            <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{item.downloaded_at ? formatDate(item.downloaded_at, i18n.language) : '—'}</span>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', gridColumn: '1 / -1' }}>
            <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.contentType')}</span>
            <span style={{ fontWeight: 600, color: 'var(--primary-slate)', textTransform: 'capitalize' }}>{item.content_type}</span>
          </div>
          {item.override_by && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.overriddenBy')}</span>
              <span style={{ fontWeight: 600, color: 'var(--primary-accent)' }}>{item.override_by}</span>
            </div>
          )}
          {item.override_at && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.overriddenAt')}</span>
              <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{formatDate(item.override_at, i18n.language)}</span>
            </div>
          )}
        </div>
      </div>

      {/* Section 2.5: Forcer le Téléchargement (occurrence unique) */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
        <h3 className="drawer-section-title">
          {t('drawer.forceDownload.sectionTitle')}
        </h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', alignItems: 'flex-start' }}>
          <button
            type="button"
            className="btn-secondary"
            disabled={!isForceDownloadEligible || forceDownloadStatus === 'loading'}
            onClick={handleForceDownload}
            title={!isForceDownloadEligible ? (
              !(item.movie || item.tvshow)
                ? t('drawer.forceDownload.unavailableNotMatched')
                : t('drawer.forceDownload.unavailableAlreadyHandled')
            ) : undefined}
          >
            {forceDownloadStatus === 'loading' ? t('drawer.forceDownload.loading') : t('drawer.forceDownload.action')}
          </button>
          {!isForceDownloadEligible && (
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontStyle: 'italic' }}>
              {!(item.movie || item.tvshow)
                ? t('drawer.forceDownload.unavailableNotMatched')
                : t('drawer.forceDownload.unavailableAlreadyHandled')}
            </span>
          )}
          {forceDownloadStatus === 'queued' && (
            <span className="badge badge-success" style={{ fontSize: '0.8rem' }}>
              {t('drawer.forceDownload.queued')}
            </span>
          )}
          {forceDownloadStatus === 'error' && (
            <span className="badge badge-failed" style={{ fontSize: '0.8rem' }}>
              {t('drawer.forceDownload.errorPrefix')} {forceDownloadError}
            </span>
          )}
        </div>
      </div>

      {/* Section 3: Informations de Provenance M3U */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <h3 className="drawer-section-title">
          {t('drawer.m3uSection')}
        </h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem', fontSize: '0.85rem' }}>
          <div>
            <strong style={{ color: 'var(--text-secondary)' }}>{t('drawer.originalName')}</strong>{' '}
            <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>{item.tvg_name}</span>
          </div>
          <div>
            <strong style={{ color: 'var(--text-secondary)' }}>{t('drawer.originalCategory')}</strong>{' '}
            <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{item.group_title}</span>
          </div>
          <div>
            <strong style={{ color: 'var(--text-secondary)' }}>{t('drawer.m3uLineNumber')}</strong>{' '}
            <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>
              {item.line_number > 0 ? (
                <span style={{ backgroundColor: 'rgba(99, 102, 241, 0.08)', padding: '0.15rem 0.4rem', borderRadius: '4px', border: '1px solid rgba(99, 102, 241, 0.15)' }}>
                  {t('drawer.line', { number: item.line_number })}
                </span>
              ) : (
                <span style={{ fontStyle: 'italic', color: 'var(--text-muted)' }}>{t('drawer.unknownLegacyImport')}</span>
              )}
            </span>
          </div>
        </div>
      </div>

      {/* Section 4: Ligne M3U d'origine complète (Copiable) */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
        <h3 className="drawer-section-title">
          {t('drawer.rawLineSection')}
        </h3>
        <div className="technical-block-container">
          <pre className="technical-block">
            {item.line_content}
          </pre>
          <button
            type="button"
            className="copy-btn"
            onClick={() => handleCopy(item.line_content, 'content')}
          >
            {copiedText === 'content' ? t('drawer.copiedBang') : t('drawer.copyIcon')}
          </button>
        </div>
      </div>

      {/* Section 5: URL Brute de streaming (Copiable) */}
      {item.line_url && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <h3 className="drawer-section-title">
            {t('drawer.rawUrlSection')}
          </h3>
          <div className="technical-block-container">
            <pre className="technical-block" style={{ maxHeight: '80px' }}>
              {item.line_url}
            </pre>
            <button
              type="button"
              className="copy-btn"
              onClick={() => handleCopy(item.line_url!, 'url')}
            >
              {copiedText === 'url' ? t('drawer.copiedBang') : t('drawer.copyIcon')}
            </button>
          </div>
        </div>
      )}

    </div>
  );
}

interface MediaOccurrenceDrawerProps {
  item: PlaylistItem | null;
  onOpenChange: (open: boolean) => void;
  contentClassName?: string;
  withOverlay?: boolean;
  modal?: boolean;
}

export function MediaOccurrenceDrawer({
  item,
  onOpenChange,
  contentClassName = 'drawer-content',
  withOverlay = true,
  modal = true,
}: MediaOccurrenceDrawerProps) {
  const { t } = useTranslation('playlist');

  return (
    <Dialog.Root open={!!item} onOpenChange={onOpenChange} modal={modal}>
      <Dialog.Portal>
        {withOverlay && <Dialog.Overlay className="drawer-overlay" />}
        <Dialog.Content className={contentClassName}>
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

          {item && <MediaOccurrenceDrawerBody item={item} />}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
