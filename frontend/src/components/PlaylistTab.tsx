import React from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { PlaylistItem } from '../types';
import { formatDate } from '../utils/date';
import { getPipelineStateBadgeClass } from '../utils/pipelineState';
import { useIsMobile } from '../hooks/useMediaQuery';

interface PlaylistTabProps {
  playlist: PlaylistItem[];
  playlistSearch: string;
  setPlaylistSearch: (search: string) => void;
  playlistSearchName: string;
  setPlaylistSearchName: (searchName: string) => void;
  playlistTMDBFilter: 'all' | 'yes' | 'no';
  setPlaylistTMDBFilter: (filter: 'all' | 'yes' | 'no') => void;
  playlistFilter: 'all' | 'movies' | 'tvshows';
  setPlaylistFilter: (filter: 'all' | 'movies' | 'tvshows') => void;
  playlistStateFilter: string;
  setPlaylistStateFilter: (state: string) => void;
  playlistTotal: number;
  playlistPage: number;
  setPlaylistPage: React.Dispatch<React.SetStateAction<number>>;
  playlistLimit: number;
  setPlaylistLimit: (limit: number) => void;
  playlistLoading: boolean;
  onOpenOverride: (item: PlaylistItem) => void;
  onResetPipeline: (id: number, contentType: string) => void;
}

export function PlaylistTab({
  playlist,
  playlistSearch,
  setPlaylistSearch,
  playlistSearchName,
  setPlaylistSearchName,
  playlistTMDBFilter,
  setPlaylistTMDBFilter,
  playlistFilter,
  setPlaylistFilter,
  playlistStateFilter,
  setPlaylistStateFilter,
  playlistTotal,
  playlistPage,
  setPlaylistPage,
  playlistLimit,
  setPlaylistLimit,
  playlistLoading,
  onOpenOverride,
  onResetPipeline
}: PlaylistTabProps) {
  const { t, i18n } = useTranslation('playlist');
  const isMobile = useIsMobile();
  const [selectedItem, setSelectedItem] = React.useState<PlaylistItem | null>(null);
  const [copiedText, setCopiedText] = React.useState<'content' | 'url' | 'hash' | null>(null);
  const [gotoPageInput, setGotoPageInput] = React.useState('');

  const totalPages = Math.ceil(playlistTotal / playlistLimit);

  const handleGotoPage = () => {
    const parsed = parseInt(gotoPageInput, 10);
    if (!isNaN(parsed)) {
      setPlaylistPage(Math.min(Math.max(1, parsed), totalPages));
    }
    setGotoPageInput('');
  };

  const handleCopy = (text: string, type: 'content' | 'url' | 'hash') => {
    navigator.clipboard.writeText(text).then(() => {
      setCopiedText(type);
      setTimeout(() => setCopiedText(null), 2000);
    });
  };

  return (
    <Tabs.Content value="playlist" className="card" style={{ padding: '2rem' }}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem', marginBottom: '1.5rem' }}>
        {/* Block Supérieur : Boutons de Type de Contenu */}
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <button
            onClick={() => { setPlaylistFilter('all'); setPlaylistPage(1); }}
            className={playlistFilter === 'all' ? 'btn-primary' : 'btn-secondary'}
            style={{ padding: '0.45rem 1rem' }}
          >
            {t('contentFilter.all')}
          </button>
          <button
            onClick={() => { setPlaylistFilter('movies'); setPlaylistPage(1); }}
            className={playlistFilter === 'movies' ? 'btn-primary' : 'btn-secondary'}
            style={{ padding: '0.45rem 1rem' }}
          >
            {t('contentFilter.movies')}
          </button>
          <button
            onClick={() => { setPlaylistFilter('tvshows'); setPlaylistPage(1); }}
            className={playlistFilter === 'tvshows' ? 'btn-primary' : 'btn-secondary'}
            style={{ padding: '0.45rem 1rem' }}
          >
            {t('contentFilter.tvshows')}
          </button>
        </div>

        {/* Block Inférieur : Grille de 4 colonnes pour filtres avancés */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
          gap: '1rem',
          alignItems: 'center'
        }}>
          {/* 1. Recherche VOD (Nom du Média) */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
            <label style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>{t('fields.mediaName')}</label>
            <input
              type="text"
              placeholder={t('search.byTitlePlaceholder')}
              value={playlistSearchName}
              onChange={e => { setPlaylistSearchName(e.target.value); setPlaylistPage(1); }}
              className="custom-input"
              style={{ width: '100%' }}
            />
          </div>

          {/* 2. Recherche Groupe / Catégorie */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
            <label style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>{t('fields.groupCategory')}</label>
            <input
              type="text"
              placeholder={t('search.byGroupPlaceholder')}
              value={playlistSearch}
              onChange={e => { setPlaylistSearch(e.target.value); setPlaylistPage(1); }}
              className="custom-input"
              style={{ width: '100%' }}
            />
          </div>

          {/* 3. Filtrage d'enrichissement TMDB */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
            <label style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>{t('fields.tmdbEnrichment')}</label>
            <select
              value={playlistTMDBFilter}
              onChange={e => { setPlaylistTMDBFilter(e.target.value as 'all' | 'yes' | 'no'); setPlaylistPage(1); }}
              className="custom-select"
              style={{ width: '100%', padding: '0.4rem 1.5rem 0.4rem 1rem' }}
            >
              <option value="all">{t('tmdbFilter.all')}</option>
              <option value="yes">{t('tmdbFilter.yes')}</option>
              <option value="no">{t('tmdbFilter.no')}</option>
            </select>
          </div>

          {/* 4. État Pipeline */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
            <label style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>{t('fields.pipelineState')}</label>
            <select
              value={playlistStateFilter}
              onChange={e => { setPlaylistStateFilter(e.target.value); setPlaylistPage(1); }}
              className="custom-select"
              style={{ width: '100%', padding: '0.4rem 1.5rem 0.4rem 1rem' }}
            >
              <option value="all">{t('stateFilter.all')}</option>
              <option value="processed">{t('stateFilter.processed')}</option>
              <option value="pending">{t('stateFilter.pending')}</option>
              <option value="downloading">{t('stateFilter.downloading')}</option>
              <option value="organizing">{t('stateFilter.organizing')}</option>
              <option value="downloaded">{t('stateFilter.downloaded')}</option>
              <option value="failed">{t('stateFilter.failed')}</option>
            </select>
          </div>
        </div>
      </div>

      {isMobile ? (
        <div>
          {playlistLoading ? (
            <div className="mobile-list-empty">{t('table.loading')}</div>
          ) : playlist.length === 0 ? (
            <div className="mobile-list-empty">{t('table.empty')}</div>
          ) : (
            playlist.map(item => (
              <div key={item.id} className="mobile-list-card" onClick={() => setSelectedItem(item)}>
                <div className="mobile-list-card-main">
                  <span className="mobile-list-card-title">{item.tvg_name}</span>
                  <span className="mobile-list-card-subtitle">{item.group_title}</span>
                </div>
                <span className={`badge ${getPipelineStateBadgeClass(item.state)}`}>
                  {item.state}
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
                <th>{t('table.headers.mediaName')}</th>
                <th>{t('table.headers.groupCategory')}</th>
                <th>{t('table.headers.tmdbEnrichment')}</th>
                <th>{t('table.headers.pipelineState')}</th>
                <th>{t('table.headers.createdAt')}</th>
                <th style={{ textAlign: 'right' }}>{t('table.headers.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {playlistLoading ? (
                <tr>
                  <td colSpan={6} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
                    <span style={{ fontWeight: 600 }}>{t('table.loading')}</span>
                  </td>
                </tr>
              ) : playlist.length === 0 ? (
                <tr>
                  <td colSpan={6} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('table.empty')}</td>
                </tr>
              ) : (
                playlist.map(item => {
                  const isMovie = item.content_type === 'movies';
                  const tmdb = isMovie ? item.movie : item.tvshow;
                  return (
                    <tr key={item.id} className="clickable-row" onClick={() => setSelectedItem(item)}>
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
                        <span className={`badge ${getPipelineStateBadgeClass(item.state)}`}>
                          {item.state}
                        </span>
                      </td>
                      <td style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{formatDate(item.created_at, i18n.language)}</td>
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
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination & Limit Selector */}
      {playlistTotal > 0 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
              {t('pagination.showing', {
                from: playlistTotal === 0 ? 0 : (playlistPage - 1) * playlistLimit + 1,
                to: Math.min(playlistPage * playlistLimit, playlistTotal),
                total: playlistTotal,
              })}
            </span>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>{t('pagination.show')}</span>
              <select
                value={playlistLimit}
                onChange={e => {
                  const limit = parseInt(e.target.value, 10);
                  setPlaylistLimit(limit);
                }}
                className="custom-select"
                style={{ padding: '0.2rem 1.5rem 0.2rem 0.5rem', fontSize: '0.8rem', width: 'auto', height: 'auto' }}
              >
                <option value="10">10</option>
                <option value="50">50</option>
                <option value="100">100</option>
              </select>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>{t('pagination.perPage')}</span>
            </div>
          </div>

          {totalPages > 1 && (
            <div style={{ display: 'flex', gap: '0.35rem', alignItems: 'center' }}>
              <button
                disabled={playlistPage === 1}
                onClick={() => setPlaylistPage(1)}
                className="btn-secondary"
                style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: playlistPage === 1 ? 0.5 : 1 }}
                title={t('pagination.firstPage')}
              >
                &lt;&lt;
              </button>
              <button
                disabled={playlistPage === 1}
                onClick={() => setPlaylistPage(p => Math.max(1, p - 1))}
                className="btn-secondary"
                style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: playlistPage === 1 ? 0.5 : 1 }}
                title={t('pagination.prevPage')}
              >
                &lt;
              </button>

              {getPaginationRange(playlistPage, totalPages).map((page, index) => {
                if (page === '...') {
                  return (
                    <span key={`ellipsis-${index}`} style={{ padding: '0.4rem 0.6rem', color: 'var(--text-muted)', fontWeight: 'bold' }}>
                      ...
                    </span>
                  );
                }
                return (
                  <button
                    key={`page-${page}`}
                    onClick={() => setPlaylistPage(page as number)}
                    className={playlistPage === page ? 'btn-primary' : 'btn-secondary'}
                    style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem' }}
                  >
                    {page}
                  </button>
                );
              })}

              <button
                disabled={playlistPage === totalPages}
                onClick={() => setPlaylistPage(p => Math.min(totalPages, p + 1))}
                className="btn-secondary"
                style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: playlistPage === totalPages ? 0.5 : 1 }}
                title={t('pagination.nextPage')}
              >
                &gt;
              </button>
              <button
                disabled={playlistPage === totalPages}
                onClick={() => setPlaylistPage(totalPages)}
                className="btn-secondary"
                style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: playlistPage === totalPages ? 0.5 : 1 }}
                title={t('pagination.lastPage')}
              >
                &gt;&gt;
              </button>

              <div style={{ display: 'flex', alignItems: 'center', gap: '0.3rem', marginLeft: '0.4rem' }}>
                <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>{t('pagination.goTo')}</span>
                <input
                  type="number"
                  min={1}
                  max={totalPages}
                  value={gotoPageInput}
                  onChange={e => setGotoPageInput(e.target.value)}
                  onKeyDown={e => { if (e.key === 'Enter') handleGotoPage(); }}
                  placeholder={String(playlistPage)}
                  className="custom-input"
                  style={{ width: '4rem', padding: '0.35rem 0.5rem', fontSize: '0.8rem' }}
                />
                <button
                  onClick={handleGotoPage}
                  className="btn-secondary"
                  style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem' }}
                >
                  {t('pagination.go')}
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Sidepanel de Détails Interactif (Drawer) */}
      <Dialog.Root open={!!selectedItem} onOpenChange={(open) => !open && setSelectedItem(null)}>
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

            {selectedItem && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', flex: 1, paddingBottom: '1rem' }}>
                
                {/* Section 1: Enrichissement Métadonnées (TMDB) */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <h3 className="drawer-section-title">
                    {t('drawer.tmdbSection')}
                  </h3>
                  {(() => {
                    const isMovie = selectedItem.content_type === 'movies';
                    const tmdb = isMovie ? selectedItem.movie : selectedItem.tvshow;
                    if (tmdb) {
                      return (
                        <div style={{ backgroundColor: 'rgba(99, 102, 241, 0.03)', border: '1px solid rgba(99, 102, 241, 0.1)', borderRadius: 'var(--radius-md)', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                          <div style={{ fontSize: '1.1rem', fontWeight: 800, color: 'var(--primary-accent)' }}>
                            {tmdb.tmdb_title} <span style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', fontWeight: 600 }}>({tmdb.tmdb_year})</span>
                          </div>
                          {tmdb.genres && (
                            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                              <strong>{t('drawer.genres')}</strong> {tmdb.genres}
                            </div>
                          )}
                          {isMovie && selectedItem.movie?.duration && (
                            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                              <strong>{t('drawer.duration')}</strong> {t('drawer.durationMinutes', { count: selectedItem.movie.duration })}
                            </div>
                          )}
                          {!isMovie && (selectedItem.tvshow?.season !== undefined || (selectedItem as any).season !== undefined) && (
                            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                              <strong>{t('drawer.position')}</strong> {t('drawer.seasonEpisode', {
                                season: selectedItem.tvshow?.season ?? (selectedItem as any).season,
                                episode: selectedItem.tvshow?.episode ?? (selectedItem as any).episode,
                              })}
                            </div>
                          )}
                          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontWeight: 500, marginTop: '0.25rem' }}>
                            <strong>{t('drawer.tmdbId')}</strong> {tmdb.tmdb_id}
                          </div>
                        </div>
                      );
                    } else {
                      return (
                        <div style={{ backgroundColor: 'var(--bg-app)', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)', padding: '1rem', textAlign: 'center', color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem' }}>
                          {t('drawer.noTmdbAssociation', { type: selectedItem.content_type })}
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
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.currentStatus')}</span>
                      <div>
                        <span className={`badge ${getPipelineStateBadgeClass(selectedItem.state)}`} style={{ fontSize: '0.75rem' }}>
                          {selectedItem.state}
                        </span>
                      </div>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.contentType')}</span>
                      <span style={{ fontWeight: 600, color: 'var(--primary-slate)', textTransform: 'capitalize' }}>{selectedItem.content_type}</span>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.importDate')}</span>
                      <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{formatDate(selectedItem.created_at, i18n.language)}</span>
                    </div>
                    {selectedItem.override_by && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                        <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.overriddenBy')}</span>
                        <span style={{ fontWeight: 600, color: 'var(--primary-accent)' }}>{selectedItem.override_by}</span>
                      </div>
                    )}
                    {selectedItem.override_at && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                        <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.overriddenAt')}</span>
                        <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{formatDate(selectedItem.override_at, i18n.language)}</span>
                      </div>
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
                      <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>{selectedItem.tvg_name}</span>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--text-secondary)' }}>{t('drawer.originalCategory')}</strong>{' '}
                      <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{selectedItem.group_title}</span>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--text-secondary)' }}>{t('drawer.m3uLineNumber')}</strong>{' '}
                      <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>
                        {selectedItem.line_number > 0 ? (
                          <span style={{ backgroundColor: 'rgba(99, 102, 241, 0.08)', padding: '0.15rem 0.4rem', borderRadius: '4px', border: '1px solid rgba(99, 102, 241, 0.15)' }}>
                            {t('drawer.line', { number: selectedItem.line_number })}
                          </span>
                        ) : (
                          <span style={{ fontStyle: 'italic', color: 'var(--text-muted)' }}>{t('drawer.unknownLegacyImport')}</span>
                        )}
                      </span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                      <strong style={{ color: 'var(--text-secondary)' }}>{t('drawer.uniqueHash')}</strong>{' '}
                      <code style={{ fontFamily: 'monospace', fontSize: '0.75rem', backgroundColor: 'var(--bg-app)', padding: '0.15rem 0.35rem', borderRadius: '4px', border: '1px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                        {selectedItem.line_hash}
                      </code>
                      <button
                        type="button"
                        onClick={() => handleCopy(selectedItem.line_hash, 'hash')}
                        style={{ fontSize: '0.75rem', color: 'var(--primary-accent)', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '0.15rem', cursor: 'pointer', border: 'none', background: 'none', padding: 0 }}
                      >
                        {copiedText === 'hash' ? t('drawer.copied') : t('drawer.copy')}
                      </button>
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
                      {selectedItem.line_content}
                    </pre>
                    <button
                      type="button"
                      className="copy-btn"
                      onClick={() => handleCopy(selectedItem.line_content, 'content')}
                    >
                      {copiedText === 'content' ? t('drawer.copiedBang') : t('drawer.copyIcon')}
                    </button>
                  </div>
                </div>

                {/* Section 5: URL Brute de streaming (Copiable) */}
                {selectedItem.line_url && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    <h3 className="drawer-section-title">
                      {t('drawer.rawUrlSection')}
                    </h3>
                    <div className="technical-block-container">
                      <pre className="technical-block" style={{ maxHeight: '80px' }}>
                        {selectedItem.line_url}
                      </pre>
                      <button
                        type="button"
                        className="copy-btn"
                        onClick={() => handleCopy(selectedItem.line_url!, 'url')}
                      >
                        {copiedText === 'url' ? t('drawer.copiedBang') : t('drawer.copyIcon')}
                      </button>
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

function getPaginationRange(current: number, total: number): (number | string)[] {
  const range: number[] = [];
  const delta = 1;

  for (let i = 1; i <= total; i++) {
    if (i === 1 || i === total || (i >= current - delta && i <= current + delta)) {
      range.push(i);
    }
  }

  const result: (number | string)[] = [];
  let prev: number | null = null;
  for (const i of range) {
    if (prev !== null) {
      if (i - prev === 2) {
        result.push(prev + 1);
      } else if (i - prev > 2) {
        result.push('...');
      }
    }
    result.push(i);
    prev = i;
  }
  return result;
}
