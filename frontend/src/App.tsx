import { useState, useEffect, useRef } from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import { useTranslation } from 'react-i18next';
import { useToast } from './hooks/useToast';
import { useApiErrorMessage } from './hooks/useApiErrorMessage';
import { useHealthAndStats } from './hooks/useHealthAndStats';
import { usePlaylist } from './hooks/usePlaylist';
import { usePlaylistView } from './hooks/usePlaylistView';
import { usePlaylistGroups } from './hooks/usePlaylistGroups';
import { useFilters } from './hooks/useFilters';
import { useLogs } from './hooks/useLogs';
import { useDownloads } from './hooks/useDownloads';
import { useErrorsTab } from './hooks/useErrorsTab';
import { useRadarrSonarr } from './hooks/useRadarrSonarr';
import { useURLState, URLStateSchema } from './hooks/useURLState';
import { useIsMobile } from './hooks/useMediaQuery';
import { api } from './services/api';
import { DownloadEnriched, PlaylistItem, ProcessingLog } from './types';
import { resolveRenameFolderName } from './utils/renameDialog';

const VALID_TABS = ['playlist', 'filters', 'logs', 'downloads', 'radarr-sonarr', 'errors'];

// The Erreurs tab is desktop-only: it is never part of the mobile bottom tab
// bar, so falling back off it while narrowing the viewport lands here.
const MOBILE_FALLBACK_TAB = 'playlist';

const TAB_URL_SCHEMA = {
  tab: {
    default: 'playlist',
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: (v: string) => VALID_TABS.includes(v),
  },
} satisfies URLStateSchema;

function readInitialActiveTab(): string {
  const urlTab = new URLSearchParams(window.location.search).get('tab');
  if (urlTab && VALID_TABS.includes(urlTab)) return urlTab;

  const storedTab = localStorage.getItem('stalkeer_active_tab');
  return storedTab && VALID_TABS.includes(storedTab) ? storedTab : 'playlist';
}

import { FloatingHeader } from './components/FloatingHeader';
import { StatsKPICards } from './components/StatsKPICards';
import { PlaylistTab } from './components/PlaylistTab';
import { FiltersTab } from './components/FiltersTab';
import { LogsTab } from './components/LogsTab';
import { DownloadsTab } from './components/DownloadsTab';
import { ErrorsTab } from './components/ErrorsTab';
import { RadarrSonarrTab } from './components/RadarrSonarrTab';

import { CreateFilterDialog } from './components/CreateFilterDialog';
import { MoveFolderDialog } from './components/MoveFolderDialog';
import { RenameFolderDialog } from './components/RenameFolderDialog';
import { ManualOverrideDialog } from './components/ManualOverrideDialog';
import { RunItemsDialog } from './components/RunItemsDialog';

export default function App() {
  const { t, i18n } = useTranslation();
  const isMobile = useIsMobile();
  const [activeTab, setActiveTab] = useState(readInitialActiveTab);
  const [, patchTabURLState] = useURLState(TAB_URL_SCHEMA);

  const { notification, showToast } = useToast();
  const { stats, fetchStats, getDownloadSuccessRatio } = useHealthAndStats();
  
  const {
    playlist, playlistSearch, setPlaylistSearch,
    playlistSearchName, setPlaylistSearchName,
    playlistTMDBFilter, setPlaylistTMDBFilter,
    playlistFilter, setPlaylistFilter, playlistStateFilter, setPlaylistStateFilter,
    playlistTotal, playlistPage, setPlaylistPage, playlistLoading, fetchPlaylist,
    playlistLimit, setPlaylistLimit,
    playlistSort, playlistOrder, setPlaylistSort
  } = usePlaylist();

  const { playlistView, setPlaylistView } = usePlaylistView();
  const {
    groups, groupsLoading, groupsTotal, groupsPage, setGroupsPage, groupsLimit,
  } = usePlaylistGroups(
    activeTab === 'playlist' && playlistView === 'grouped',
    playlistFilter,
    playlistStateFilter,
    playlistSearch,
    playlistSearchName,
    playlistTMDBFilter
  );

  const { filters, filtersLoading, fetchFilters, deleteFilter } = useFilters();
  const { logs, logsLoading, fetchLogs } = useLogs(activeTab === 'logs');
  
  const {
    downloads, downloadsLoading, statusFilter, setStatusFilter,
    typeFilter, setTypeFilter, problemFilter, setProblemFilter,
    downloadsPage, setDownloadsPage, downloadsLimit, setDownloadsLimit, downloadsTotal,
    configPaths, fetchDownloads, updateDownloadPath
  } = useDownloads(activeTab === 'downloads');

  const {
    errors, errorsLoading, reasonFilter, setReasonFilter,
    errorsPage, setErrorsPage, errorsLimit, setErrorsLimit, errorsTotal, fetchErrors
  } = useErrorsTab(activeTab === 'errors');

  const {
    filmsItems, filmsLoading, filmsError, filmsTotal, filmsPage, setFilmsPage, filmsLimit, fetchFilms,
    filmsSearch, setFilmsSearch, filmsFilter, setFilmsFilter,
    seriesItems, seriesLoading, seriesError, seriesTotal, seriesPage, setSeriesPage, seriesLimit, fetchSeries, refreshSeries,
    seriesSearch, setSeriesSearch, seriesFilter, setSeriesFilter,
    stats: radarrSonarrStats, statsLoading: radarrSonarrStatsLoading, statsError: radarrSonarrStatsError, fetchStats: fetchRadarrSonarrStats,
  } = useRadarrSonarr(activeTab === 'radarr-sonarr');

  const [isCreateFilterOpen, setIsCreateFilterOpen] = useState(false);
  const [isMoveOpen, setIsMoveOpen] = useState(false);
  const [moveItem, setMoveItem] = useState<{ id: number; title: string; type: 'movie' | 'tvshow'; currentPath?: string } | null>(null);
  const [isRenameOpen, setIsRenameOpen] = useState(false);
  const [renameItem, setRenameItem] = useState<{ id: number; title: string; folderName: string } | null>(null);
  const [isOverrideOpen, setIsOverrideOpen] = useState(false);
  const [overrideItemData, setOverrideItemData] = useState<PlaylistItem | null>(null);
  const overrideExtraSuccessRef = useRef<(() => void) | null>(null);
  const [isRunItemsOpen, setIsRunItemsOpen] = useState(false);
  const [selectedRunLog, setSelectedRunLog] = useState<ProcessingLog | null>(null);

  useEffect(() => {
    localStorage.setItem('stalkeer_active_tab', activeTab);
    patchTabURLState({ tab: activeTab });
  }, [activeTab, patchTabURLState]);

  useEffect(() => {
    document.documentElement.lang = i18n.language;
    const handleLanguageChanged = (lng: string) => {
      document.documentElement.lang = lng;
    };
    i18n.on('languageChanged', handleLanguageChanged);
    return () => {
      i18n.off('languageChanged', handleLanguageChanged);
    };
  }, [i18n]);

  useEffect(() => {
    if (activeTab === 'filters') {
      fetchFilters();
    }
  }, [activeTab, fetchFilters]);

  // The Erreurs tab is desktop-only (never in the mobile bottom tab bar): if
  // the viewport narrows while it's active, fall back to another tab instead
  // of continuing to render it.
  useEffect(() => {
    if (isMobile && activeTab === 'errors') {
      setActiveTab(MOBILE_FALLBACK_TAB);
    }
  }, [isMobile, activeTab]);

  const translateApiError = useApiErrorMessage();

  const handleResetPipeline = (id: number, contentType: string, onDone?: () => void) => {
    api.resetPipeline(id, contentType)
      .then(() => {
        showToast(t('toasts.resetSuccess'));
        fetchPlaylist();
        fetchStats();
        onDone?.();
      })
      .catch(err => showToast(translateApiError(err), 'error'));
  };

  const handleOpenOverride = (item: PlaylistItem, onSuccess?: () => void) => {
    setOverrideItemData(item);
    setIsOverrideOpen(true);
    overrideExtraSuccessRef.current = onSuccess ?? null;
  };

  const handleDeleteFilter = (id: number) => {
    if (!confirm(t('confirm.deleteFilter'))) return;
    deleteFilter(id)
      .then(() => showToast(t('toasts.filterDeleted')))
      .catch(err => showToast(translateApiError(err), 'error'));
  };

  const openMoveDialog = (item: DownloadEnriched) => {
    const unnamedMedia = t('media.unnamed');
    const filename = item.download_path ? item.download_path.split('/').pop() || unnamedMedia : unnamedMedia;
    const moveMeta = {
      id: item.id,
      title: item.content?.title || filename,
      type: (item.content?.type === 'tvshows' || item.download_path?.includes('tvshows')) ? 'tvshow' as const : 'movie' as const,
      currentPath: item.download_path,
    };
    setMoveItem(moveMeta);
    setIsMoveOpen(true);
  };

  const openRenameDialog = (item: DownloadEnriched) => {
    const unnamedMedia = t('media.unnamed');
    const filename = item.download_path ? item.download_path.split('/').pop() || unnamedMedia : unnamedMedia;
    setRenameItem({
      id: item.id,
      title: item.content?.title || filename,
      folderName: resolveRenameFolderName(item, filename),
    });
    setIsRenameOpen(true);
  };

  const handleRenameSuccess = (id: number, newPath: string, message: string) => {
    showToast(message);
    updateDownloadPath(id, newPath);
  };

  const tabs = [
    { value: 'playlist', icon: '🎬', label: t('tabs.playlist') },
    { value: 'filters', icon: '🔍', label: t('tabs.filters') },
    { value: 'logs', icon: '⚙️', label: t('tabs.logs') },
    { value: 'downloads', icon: '📥', label: t('tabs.downloads') },
    { value: 'radarr-sonarr', icon: '🎯', label: t('tabs.radarrSonarr') },
  ];

  return (
    <div className="app-container" style={{ maxWidth: 1600, margin: '0 auto', paddingTop: '2rem', paddingLeft: '1.5rem', paddingRight: '1.5rem' }}>
      {/* Toast Notification */}
      {notification && (
        <div style={{
          position: 'fixed', bottom: '24px', right: '24px', zIndex: 1000, display: 'flex', alignItems: 'center', gap: '0.5rem',
          backgroundColor: notification.type === 'success' ? 'var(--status-success-bg)' : 'var(--status-failed-bg)',
          color: notification.type === 'success' ? 'var(--status-success-text)' : 'var(--status-failed-text)',
          border: `1px solid ${notification.type === 'success' ? 'var(--status-success-border)' : 'var(--status-failed-border)'}`,
          padding: '1rem 1.5rem', borderRadius: 'var(--radius-md)', boxShadow: 'var(--shadow-lg)', fontWeight: 600, fontSize: '0.875rem',
          animation: 'contentShow 150ms ease-out'
        }}>
          {notification.type === 'success' ? '✓' : '⚠️'} {notification.message}
        </div>
      )}

      <FloatingHeader />
      <StatsKPICards stats={stats} getDownloadSuccessRatio={getDownloadSuccessRatio} />

      <Tabs.Root value={activeTab} onValueChange={setActiveTab}>
        <Tabs.List className="segmented-tabs-list">
          <Tabs.Trigger value="playlist" className="segmented-tabs-trigger">🎬 {t('tabs.playlist')}</Tabs.Trigger>
          <Tabs.Trigger value="filters" className="segmented-tabs-trigger">🔍 {t('tabs.filters')}</Tabs.Trigger>
          <Tabs.Trigger value="logs" className="segmented-tabs-trigger">⚙️ {t('tabs.logs')}</Tabs.Trigger>
          <Tabs.Trigger value="downloads" className="segmented-tabs-trigger">📥 {t('tabs.downloads')}</Tabs.Trigger>
          <Tabs.Trigger value="radarr-sonarr" className="segmented-tabs-trigger">🎯 {t('tabs.radarrSonarr')}</Tabs.Trigger>
          <Tabs.Trigger value="errors" className="segmented-tabs-trigger">🩺 {t('tabs.errors')}</Tabs.Trigger>
        </Tabs.List>

        <PlaylistTab
          playlist={playlist} playlistSearch={playlistSearch} setPlaylistSearch={setPlaylistSearch}
          playlistSearchName={playlistSearchName} setPlaylistSearchName={setPlaylistSearchName}
          playlistTMDBFilter={playlistTMDBFilter} setPlaylistTMDBFilter={setPlaylistTMDBFilter}
          playlistFilter={playlistFilter} setPlaylistFilter={setPlaylistFilter}
          playlistStateFilter={playlistStateFilter} setPlaylistStateFilter={setPlaylistStateFilter}
          playlistTotal={playlistTotal} playlistPage={playlistPage} setPlaylistPage={setPlaylistPage}
          playlistLimit={playlistLimit} setPlaylistLimit={setPlaylistLimit}
          playlistSort={playlistSort} playlistOrder={playlistOrder} setPlaylistSort={setPlaylistSort}
          playlistLoading={playlistLoading} onOpenOverride={handleOpenOverride}
          onResetPipeline={handleResetPipeline}
          playlistView={playlistView} setPlaylistView={setPlaylistView}
          groups={groups} groupsLoading={groupsLoading} groupsTotal={groupsTotal}
          groupsPage={groupsPage} setGroupsPage={setGroupsPage} groupsLimit={groupsLimit}
        />

        <FiltersTab
          filters={filters} filtersLoading={filtersLoading} onDeleteFilter={handleDeleteFilter}
          onOpenCreate={() => setIsCreateFilterOpen(true)}
        />

        <LogsTab
          logs={logs} logsLoading={logsLoading} onFetchLogs={fetchLogs}
          onRowClick={(log) => { setSelectedRunLog(log); setIsRunItemsOpen(true); }}
        />

        <DownloadsTab
          downloads={downloads} downloadsLoading={downloadsLoading} statusFilter={statusFilter} setStatusFilter={setStatusFilter}
          typeFilter={typeFilter} setTypeFilter={setTypeFilter} problemFilter={problemFilter} setProblemFilter={setProblemFilter}
          downloadsTotal={downloadsTotal} downloadsPage={downloadsPage} setDownloadsPage={setDownloadsPage}
          downloadsLimit={downloadsLimit} setDownloadsLimit={setDownloadsLimit}
          onFetchDownloads={fetchDownloads} onOpenMoveDialog={openMoveDialog}
          onOpenRenameDialog={openRenameDialog}
        />

        <ErrorsTab
          errors={errors} errorsLoading={errorsLoading} reasonFilter={reasonFilter} setReasonFilter={setReasonFilter}
          errorsTotal={errorsTotal} errorsPage={errorsPage} setErrorsPage={setErrorsPage}
          errorsLimit={errorsLimit} setErrorsLimit={setErrorsLimit}
          onFetchErrors={fetchErrors}
        />

        <RadarrSonarrTab
          filmsItems={filmsItems} filmsLoading={filmsLoading} filmsError={filmsError}
          filmsTotal={filmsTotal} filmsPage={filmsPage} setFilmsPage={setFilmsPage}
          filmsLimit={filmsLimit} fetchFilms={fetchFilms}
          filmsSearch={filmsSearch} setFilmsSearch={setFilmsSearch}
          filmsFilter={filmsFilter} setFilmsFilter={setFilmsFilter}
          seriesItems={seriesItems} seriesLoading={seriesLoading} seriesError={seriesError}
          seriesTotal={seriesTotal} seriesPage={seriesPage} setSeriesPage={setSeriesPage}
          seriesLimit={seriesLimit} fetchSeries={fetchSeries} refreshSeries={refreshSeries}
          seriesSearch={seriesSearch} setSeriesSearch={setSeriesSearch}
          seriesFilter={seriesFilter} setSeriesFilter={setSeriesFilter}
          stats={radarrSonarrStats} statsLoading={radarrSonarrStatsLoading} statsError={radarrSonarrStatsError} fetchStats={fetchRadarrSonarrStats}
          onOpenOverride={handleOpenOverride}
        />
      </Tabs.Root>

      {isMobile && (
        <nav className="mobile-tab-bar">
          {tabs.map(tab => (
            <button
              key={tab.value}
              type="button"
              className="mobile-tab-bar-btn"
              data-state={activeTab === tab.value ? 'active' : 'inactive'}
              onClick={() => setActiveTab(tab.value)}
            >
              <span className="mobile-tab-bar-icon">{tab.icon}</span>
              <span className="mobile-tab-bar-label">{tab.label}</span>
            </button>
          ))}
        </nav>
      )}

      <CreateFilterDialog isOpen={isCreateFilterOpen} onOpenChange={setIsCreateFilterOpen} onSuccess={(msg) => { showToast(msg); fetchFilters(); }} />
      
      <MoveFolderDialog isOpen={isMoveOpen} onOpenChange={setIsMoveOpen} moveItem={moveItem} configPaths={configPaths} onSuccess={(msg) => { showToast(msg); fetchDownloads(); fetchStats(); }} />

      <RenameFolderDialog isOpen={isRenameOpen} onOpenChange={setIsRenameOpen} renameItem={renameItem} onSuccess={handleRenameSuccess} />

      <ManualOverrideDialog
        isOpen={isOverrideOpen} onOpenChange={setIsOverrideOpen} overrideItemData={overrideItemData}
        onSuccess={(msg) => {
          showToast(msg);
          fetchPlaylist();
          fetchStats();
          overrideExtraSuccessRef.current?.();
          overrideExtraSuccessRef.current = null;
        }}
      />

      <RunItemsDialog
        isOpen={isRunItemsOpen} onOpenChange={setIsRunItemsOpen} log={selectedRunLog}
        onOpenOverride={handleOpenOverride} onResetPipeline={handleResetPipeline}
      />
    </div>
  );
}
