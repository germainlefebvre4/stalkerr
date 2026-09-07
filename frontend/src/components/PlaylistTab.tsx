import React from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import { useTranslation } from 'react-i18next';
import { PlaylistItem, MediaGroupItem } from '../types';
import { useIsMobile } from '../hooks/useMediaQuery';
import { PlaylistItemsTable } from './PlaylistItemsTable';
import { PlaylistGroupedView } from './PlaylistGroupedView';
import { Pagination } from './Pagination';
import { MediaOccurrenceDrawer } from './MediaOccurrenceDrawer';
import { api } from '../services/api';

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
  onOpenOverride: (item: PlaylistItem, onSuccess?: () => void) => void;
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
  const { t } = useTranslation('playlist');
  const isMobile = useIsMobile();
  const [selectedItem, setSelectedItem] = React.useState<PlaylistItem | null>(null);
  const [advancedFiltersOpen, setAdvancedFiltersOpen] = React.useState(false);

  const handleDrawerOpenOverride = (item: PlaylistItem) => {
    onOpenOverride(item, () => {
      api.getItem(item.id).then(setSelectedItem).catch(() => {});
    });
  };

  const activeAdvancedFilterCount = [
    playlistSearchName,
    playlistSearch,
    playlistTMDBFilter !== 'all',
    playlistStateFilter !== 'all',
  ].filter(Boolean).length;

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
          onRowClick={setSelectedItem}
        />
      )}

      <MediaOccurrenceDrawer
        item={selectedItem}
        onOpenChange={(open) => !open && setSelectedItem(null)}
        onOpenOverride={handleDrawerOpenOverride}
      />
    </Tabs.Content>
  );
}
