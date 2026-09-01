import React from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { PlaylistItem, MediaGroupItem } from '../types';
import { formatDate } from '../utils/date';
import {
  getProcessingStatus,
  getDownloadStatus,
  getProcessingStatusBadgeClass,
  getDownloadStatusBadgeClass,
} from '../utils/pipelineState';
import { useIsMobile } from '../hooks/useMediaQuery';
import { PlaylistItemsTable } from './PlaylistItemsTable';
import { PlaylistGroupedView } from './PlaylistGroupedView';
import { Pagination } from './Pagination';
import { api, ApiError } from '../services/api';

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
  playlistSort: string;
  playlistOrder: string;
  setPlaylistSort: (column: string) => void;
  playlistLoading: boolean;
  onOpenOverride: (item: PlaylistItem) => void;
  onResetPipeline: (id: number, contentType: string) => void;
  playlistView: 'items' | 'grouped';
  setPlaylistView: (view: 'items' | 'grouped') => void;
  groups: MediaGroupItem[];
  groupsLoading: boolean;
  groupsTotal: number;
  groupsPage: number;
  setGroupsPage: React.Dispatch<React.SetStateAction<number>>;
  groupsLimit: number;
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
  playlistSort,
  playlistOrder,
  setPlaylistSort,
  playlistLoading,
  onOpenOverride,
  onResetPipeline,
  playlistView,
  setPlaylistView,
  groups,
  groupsLoading,
  groupsTotal,
  groupsPage,
  setGroupsPage,
  groupsLimit,
}: PlaylistTabProps) {
  const { t, i18n } = useTranslation('playlist');
  const isMobile = useIsMobile();
  const [selectedItem, setSelectedItem] = React.useState<PlaylistItem | null>(null);
  const [copiedText, setCopiedText] = React.useState<'content' | 'url' | null>(null);
  const [advancedFiltersOpen, setAdvancedFiltersOpen] = React.useState(false);
  const [forceDownloadStatus, setForceDownloadStatus] = React.useState<'idle' | 'loading' | 'queued' | 'error'>('idle');
  const [forceDownloadError, setForceDownloadError] = React.useState<string | null>(null);

  React.useEffect(() => {
    setForceDownloadStatus('idle');
    setForceDownloadError(null);
  }, [selectedItem?.id]);

  const isForceDownloadEligible = !!selectedItem
    && !!(selectedItem.movie || selectedItem.tvshow)
    && selectedItem.state !== 'downloaded'
    && selectedItem.state !== 'downloading';

  const handleForceDownload = async () => {
    if (!selectedItem) return;
    setForceDownloadStatus('loading');
    setForceDownloadError(null);
    try {
      await api.forceDownload(selectedItem.id);
      setForceDownloadStatus('queued');
    } catch (err) {
      setForceDownloadStatus('error');
      setForceDownloadError(err instanceof ApiError ? err.message : t('drawer.forceDownload.genericError'));
    }
  };

  const activeAdvancedFilterCount = [
    playlistSearchName,
    playlistSearch,
    playlistTMDBFilter !== 'all',
    playlistStateFilter !== 'all',
  ].filter(Boolean).length;

  const handleCopy = (text: string, type: 'content' | 'url') => {
    navigator.clipboard.writeText(text).then(() => {
      setCopiedText(type);
      setTimeout(() => setCopiedText(null), 2000);
    });
  };

  return (
    <Tabs.Content value="playlist" className="card tab-panel">
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem', marginBottom: '1.5rem' }}>
        {/* Block Supérieur : Boutons de Type de Contenu + Bascule Items / Films & Séries */}
        <div style={{ display: 'flex', gap: isMobile ? '0.35rem' : '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <div style={{ display: 'flex', gap: isMobile ? '0.35rem' : '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
            <button
              onClick={() => { setPlaylistFilter('all'); setPlaylistPage(1); }}
              className={playlistFilter === 'all' ? 'btn-primary' : 'btn-secondary'}
              style={{ padding: isMobile ? '0.35rem 0.6rem' : '0.45rem 1rem' }}
            >
              {t('contentFilter.all')}
            </button>
            <button
              onClick={() => { setPlaylistFilter('movies'); setPlaylistPage(1); }}
              className={playlistFilter === 'movies' ? 'btn-primary' : 'btn-secondary'}
              style={{ padding: isMobile ? '0.35rem 0.6rem' : '0.45rem 1rem' }}
            >
              {t('contentFilter.movies')}
            </button>
            <button
              onClick={() => { setPlaylistFilter('tvshows'); setPlaylistPage(1); }}
              className={playlistFilter === 'tvshows' ? 'btn-primary' : 'btn-secondary'}
              style={{ padding: isMobile ? '0.35rem 0.6rem' : '0.45rem 1rem' }}
            >
              {t('contentFilter.tvshows')}
            </button>
          </div>

          <label
            className="view-toggle-switch"
            style={{ marginLeft: 'auto' }}
            title={playlistView === 'grouped' ? t('view.switchToItems') : t('view.switchToGrouped')}
          >
            <span className={`view-toggle-switch-icon${playlistView === 'items' ? ' is-active' : ''}`} aria-hidden="true">☰</span>
            <input
              type="checkbox"
              className="view-toggle-switch-input"
              checked={playlistView === 'grouped'}
              onChange={e => setPlaylistView(e.target.checked ? 'grouped' : 'items')}
            />
            <span className="view-toggle-switch-track">
              <span className="view-toggle-switch-knob" />
            </span>
            <span className={`view-toggle-switch-icon${playlistView === 'grouped' ? ' is-active' : ''}`} aria-hidden="true">🎬</span>
          </label>
        </div>

        {/* Block Inférieur : Grille de 4 colonnes pour filtres avancés */}
        {isMobile ? (
          <div>
            <button
              type="button"
              className="advanced-filters-toggle"
              onClick={() => setAdvancedFiltersOpen(open => !open)}
              aria-expanded={advancedFiltersOpen}
            >
              <span className="advanced-filters-toggle-label">
                {t('advancedFilters.toggle')}
                {activeAdvancedFilterCount > 0 && (
                  <span className="advanced-filters-count-badge" title={t('advancedFilters.activeCount', { count: activeAdvancedFilterCount })}>
                    {activeAdvancedFilterCount}
                  </span>
                )}
              </span>
              <span className={`advanced-filters-chevron${advancedFiltersOpen ? ' is-open' : ''}`}>▾</span>
            </button>

            {advancedFiltersOpen && (
              <div className="advanced-filters-grid" style={{
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
            )}
          </div>
        ) : (
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
        )}
      </div>

      {playlistView === 'items' ? (
        <>
          <PlaylistItemsTable
            items={playlist}
            loading={playlistLoading}
            onRowClick={setSelectedItem}
            onOpenOverride={onOpenOverride}
            onResetPipeline={onResetPipeline}
            sort={playlistSort}
            order={playlistOrder}
            onSort={setPlaylistSort}
          />
          <Pagination
            total={playlistTotal}
            page={playlistPage}
            setPage={setPlaylistPage}
            limit={playlistLimit}
            setLimit={setPlaylistLimit}
            limitOptions={[10, 50, 100]}
          />
        </>
      ) : (
        <PlaylistGroupedView
          groups={groups}
          groupsLoading={groupsLoading}
          groupsTotal={groupsTotal}
          groupsPage={groupsPage}
          setGroupsPage={setGroupsPage}
          groupsLimit={groupsLimit}
          onOpenOverride={onOpenOverride}
          onResetPipeline={onResetPipeline}
        />
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
                    const legacySeasonEpisode = selectedItem as PlaylistItem & { season?: number; episode?: number };
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
                              {isMovie && selectedItem.movie?.duration && (
                                <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                                  <strong>{t('drawer.duration')}</strong> {t('drawer.durationMinutes', { count: selectedItem.movie.duration })}
                                </div>
                              )}
                              {!isMovie && (selectedItem.tvshow?.season !== undefined || legacySeasonEpisode.season !== undefined) && (
                                <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                                  <strong>{t('drawer.position')}</strong> {t('drawer.seasonEpisode', {
                                    season: selectedItem.tvshow?.season ?? legacySeasonEpisode.season,
                                    episode: selectedItem.tvshow?.episode ?? legacySeasonEpisode.episode,
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
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.processingStatus')}</span>
                      <div>
                        <span className={`badge ${getProcessingStatusBadgeClass(getProcessingStatus(selectedItem.state))}`} style={{ fontSize: '0.75rem' }}>
                          {getProcessingStatus(selectedItem.state)}
                        </span>
                      </div>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.downloadStatus')}</span>
                      <div>
                        <span className={`badge ${getDownloadStatusBadgeClass(getDownloadStatus(selectedItem.state))}`} style={{ fontSize: '0.75rem' }}>
                          {getDownloadStatus(selectedItem.state) === 'not_downloaded' ? t('pipelineStatus.notDownloaded') : getDownloadStatus(selectedItem.state)}
                        </span>
                      </div>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.importDate')}</span>
                      <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{formatDate(selectedItem.created_at, i18n.language)}</span>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.downloadDate')}</span>
                      <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{selectedItem.downloaded_at ? formatDate(selectedItem.downloaded_at, i18n.language) : '—'}</span>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem', gridColumn: '1 / -1' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>{t('drawer.contentType')}</span>
                      <span style={{ fontWeight: 600, color: 'var(--primary-slate)', textTransform: 'capitalize' }}>{selectedItem.content_type}</span>
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
                        !(selectedItem.movie || selectedItem.tvshow)
                          ? t('drawer.forceDownload.unavailableNotMatched')
                          : t('drawer.forceDownload.unavailableAlreadyHandled')
                      ) : undefined}
                    >
                      {forceDownloadStatus === 'loading' ? t('drawer.forceDownload.loading') : t('drawer.forceDownload.action')}
                    </button>
                    {!isForceDownloadEligible && (
                      <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontStyle: 'italic' }}>
                        {!(selectedItem.movie || selectedItem.tvshow)
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
