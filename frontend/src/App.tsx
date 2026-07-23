import { useState, useEffect } from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import { useToast } from './hooks/useToast';
import { useHealthAndStats } from './hooks/useHealthAndStats';
import { usePlaylist } from './hooks/usePlaylist';
import { useFilters } from './hooks/useFilters';
import { useLogs } from './hooks/useLogs';
import { useDownloads } from './hooks/useDownloads';
import { api } from './services/api';
import { DownloadEnriched, PlaylistItem } from './types';

import { FloatingHeader } from './components/FloatingHeader';
import { StatsKPICards } from './components/StatsKPICards';
import { PlaylistTab } from './components/PlaylistTab';
import { FiltersTab } from './components/FiltersTab';
import { LogsTab } from './components/LogsTab';
import { DownloadsTab } from './components/DownloadsTab';

import { CreateFilterDialog } from './components/CreateFilterDialog';
import { MoveFolderDialog } from './components/MoveFolderDialog';
import { ManualOverrideDialog } from './components/ManualOverrideDialog';

export default function App() {
  const [activeTab, setActiveTab] = useState(() => {
    return localStorage.getItem('stalkeer_active_tab') || 'playlist';
  });
  
  const { notification, showToast } = useToast();
  const { healthStatus, stats, fetchStats, getDownloadSuccessRatio } = useHealthAndStats();
  
  const {
    playlist, playlistSearch, setPlaylistSearch,
    playlistFilter, setPlaylistFilter, playlistStateFilter, setPlaylistStateFilter,
    playlistTotal, playlistPage, setPlaylistPage, playlistLoading, fetchPlaylist
  } = usePlaylist();
  
  const { filters, filtersLoading, fetchFilters, deleteFilter } = useFilters();
  const { logs, logsLoading, fetchLogs } = useLogs(activeTab === 'logs');
  
  const {
    downloads, downloadsLoading, statusFilter, setStatusFilter,
    typeFilter, setTypeFilter, problemFilter, setProblemFilter,
    configPaths, fetchDownloads
  } = useDownloads(activeTab === 'downloads');

  const [isCreateFilterOpen, setIsCreateFilterOpen] = useState(false);
  const [isMoveOpen, setIsMoveOpen] = useState(false);
  const [moveItem, setMoveItem] = useState<{ id: number; title: string; type: 'movie' | 'tvshow'; currentPath?: string } | null>(null);
  const [isOverrideOpen, setIsOverrideOpen] = useState(false);
  const [overrideItemData, setOverrideItemData] = useState<PlaylistItem | null>(null);

  useEffect(() => {
    localStorage.setItem('stalkeer_active_tab', activeTab);
  }, [activeTab]);

  useEffect(() => {
    if (activeTab === 'filters') {
      fetchFilters();
    }
  }, [activeTab, fetchFilters]);

  const handleResetPipeline = (id: number, contentType: string) => {
    api.resetPipeline(id, contentType)
      .then(() => {
        showToast('Flux réinitialisé avec succès ! Prêt pour le retraitement.');
        fetchPlaylist();
        fetchStats();
      })
      .catch(err => showToast(err.message, 'error'));
  };

  const handleDeleteFilter = (id: number) => {
    if (!confirm('Êtes-vous sûr de vouloir supprimer définitivement ce filtre de tri ?')) return;
    deleteFilter(id)
      .then(() => showToast('Filtre supprimé avec succès !'))
      .catch(err => showToast(err.message, 'error'));
  };

  const openMoveDialog = (item: DownloadEnriched) => {
    const filename = item.download_path ? item.download_path.split('/').pop() || 'Média' : 'Média';
    const moveMeta = {
      id: item.id,
      title: item.content?.title || filename,
      type: (item.content?.type === 'tvshows' || item.download_path?.includes('tvshows')) ? 'tvshow' as const : 'movie' as const,
      currentPath: item.download_path,
    };
    setMoveItem(moveMeta);
    setIsMoveOpen(true);
  };

  return (
    <div style={{ maxWidth: 1600, margin: '0 auto', padding: '2rem 1.5rem' }}>
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

      <FloatingHeader healthStatus={healthStatus} />
      <StatsKPICards stats={stats} getDownloadSuccessRatio={getDownloadSuccessRatio} />

      <Tabs.Root value={activeTab} onValueChange={setActiveTab}>
        <Tabs.List className="segmented-tabs-list">
          <Tabs.Trigger value="playlist" className="segmented-tabs-trigger">🗒️ Playlist M3U</Tabs.Trigger>
          <Tabs.Trigger value="filters" className="segmented-tabs-trigger">🔍 Filtres de Tri</Tabs.Trigger>
          <Tabs.Trigger value="logs" className="segmented-tabs-trigger">⚙️ Traitements</Tabs.Trigger>
          <Tabs.Trigger value="downloads" className="segmented-tabs-trigger">📥 Téléchargements</Tabs.Trigger>
        </Tabs.List>

        <PlaylistTab
          playlist={playlist} playlistSearch={playlistSearch} setPlaylistSearch={setPlaylistSearch}
          playlistFilter={playlistFilter} setPlaylistFilter={setPlaylistFilter}
          playlistStateFilter={playlistStateFilter} setPlaylistStateFilter={setPlaylistStateFilter}
          playlistTotal={playlistTotal} playlistPage={playlistPage} setPlaylistPage={setPlaylistPage}
          playlistLoading={playlistLoading} onOpenOverride={(item) => { setOverrideItemData(item); setIsOverrideOpen(true); }}
          onResetPipeline={handleResetPipeline}
        />

        <FiltersTab
          filters={filters} filtersLoading={filtersLoading} onDeleteFilter={handleDeleteFilter}
          onOpenCreate={() => setIsCreateFilterOpen(true)}
        />

        <LogsTab logs={logs} logsLoading={logsLoading} onFetchLogs={fetchLogs} />

        <DownloadsTab
          downloads={downloads} downloadsLoading={downloadsLoading} statusFilter={statusFilter} setStatusFilter={setStatusFilter}
          typeFilter={typeFilter} setTypeFilter={setTypeFilter} problemFilter={problemFilter} setProblemFilter={setProblemFilter}
          onFetchDownloads={fetchDownloads} onOpenMoveDialog={openMoveDialog}
        />
      </Tabs.Root>

      <CreateFilterDialog isOpen={isCreateFilterOpen} onOpenChange={setIsCreateFilterOpen} onSuccess={(msg) => { showToast(msg); fetchFilters(); }} />
      
      <MoveFolderDialog isOpen={isMoveOpen} onOpenChange={setIsMoveOpen} moveItem={moveItem} configPaths={configPaths} onSuccess={(msg) => { showToast(msg); fetchDownloads(); fetchStats(); }} />
      
      <ManualOverrideDialog isOpen={isOverrideOpen} onOpenChange={setIsOverrideOpen} overrideItemData={overrideItemData} onSuccess={(msg) => { showToast(msg); fetchPlaylist(); fetchStats(); }} />
    </div>
  );
}
