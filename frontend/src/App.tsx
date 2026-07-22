import React, { useState, useEffect } from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import * as Dialog from '@radix-ui/react-dialog';
import * as Progress from '@radix-ui/react-progress';

// TypeScript DTO Interfaces
interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
  total_pages: number;
}

interface MovieResponse {
  id: number;
  tmdb_id: number;
  tmdb_title: string;
  tmdb_year: number;
  genres?: string;
  duration?: number;
}

interface TVShowResponse {
  id: number;
  tmdb_id: number;
  tmdb_title: string;
  tmdb_year: number;
  genres?: string;
  season?: number;
  episode?: number;
}

interface PlaylistItem {
  id: number;
  tvg_name: string;
  group_title: string;
  content_type: 'movies' | 'tvshows' | 'channels' | 'uncategorized';
  state: string;
  movie?: MovieResponse;
  tvshow?: TVShowResponse;
  override_by?: string;
  override_at?: string;
  created_at: string;
}

interface ProcessingLog {
  id: number;
  action: string;
  item_count: number;
  status: 'success' | 'failed' | 'in_progress';
  started_at: string;
  completed_at?: string;
  error_message?: string;
}

interface DownloadInfo {
  id: number;
  url: string;
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'failed' | 'retrying';
  download_path?: string;
  file_size?: number;
  bytes_downloaded?: number;
  total_bytes?: number;
  retry_count: number;
  error_message?: string;
  updated_at: string;
}

interface ConfigPaths {
  movies_path: string;
  tvshows_path: string;
}

interface StatsResponse {
  total_items: number;
  by_content_type: Record<string, number>;
  by_state: Record<string, number>;
}

interface FilterConfig {
  id: number;
  name: string;
  attribute: string;
  include_patterns?: string;
  exclude_patterns?: string;
  is_runtime: boolean;
}

export default function App() {
  const [activeTab, setActiveTab] = useState('playlist');
  const [healthStatus, setHealthStatus] = useState<'healthy' | 'unhealthy' | 'checking'>('checking');
  
  // Stats KPI State
  const [stats, setStats] = useState<StatsResponse | null>(null);

  // Playlist State
  const [playlist, setPlaylist] = useState<PlaylistItem[]>([]);
  const [playlistSearch, setPlaylistSearch] = useState('');
  const [playlistFilter, setPlaylistFilter] = useState<'all' | 'movies' | 'tvshows'>('all');
  const [playlistStateFilter, setPlaylistStateFilter] = useState<string>('all');
  const [playlistTotal, setPlaylistTotal] = useState(0);
  const [playlistPage, setPlaylistPage] = useState(1);
  const [playlistLoading, setPlaylistLoading] = useState(false);

  // Processing Logs State
  const [logs, setLogs] = useState<ProcessingLog[]>([]);
  const [logsLoading, setLogsLoading] = useState(false);

  // Downloads State
  const [downloads, setDownloads] = useState<DownloadInfo[]>([]);
  const [downloadsLoading, setDownloadsLoading] = useState(false);

  // Re-organization Modal State
  const [isMoveOpen, setIsMoveOpen] = useState(false);
  const [moveItem, setMoveItem] = useState<{ id: number; title: string; type: 'movie' | 'tvshow'; currentPath?: string } | null>(null);
  const [configPaths, setConfigPaths] = useState<ConfigPaths | null>(null);
  const [targetDir, setTargetDir] = useState('');
  const [customDirInput, setCustomDirInput] = useState('');
  const [isMoving, setIsMoving] = useState(false);
  const [moveError, setMoveError] = useState<string | null>(null);
  const [moveSuccess, setMoveSuccess] = useState<string | null>(null);

  // Manual Override Modal State
  const [isOverrideOpen, setIsOverrideOpen] = useState(false);
  const [overrideItemData, setOverrideItemData] = useState<PlaylistItem | null>(null);
  const [overrideSearchQuery, setOverrideSearchQuery] = useState('');
  const [overrideMediaType, setOverrideMediaType] = useState<'movie' | 'tvshow'>('movie');
  const [overrideYear, setOverrideSearchYear] = useState('');
  const [overrideSeason, setOverrideSeason] = useState('');
  const [overrideEpisode, setOverrideEpisode] = useState('');
  const [overrideSearchResults, setOverrideSearchResults] = useState<any[]>([]);
  const [overrideSearchLoading, setOverrideSearchLoading] = useState(false);
  const [selectedResult, setSelectedResult] = useState<any | null>(null);
  const [isSubmittingOverride, setIsSubmittingOverride] = useState(false);
  const [overrideError, setOverrideError] = useState<string | null>(null);

  // Filters State
  const [filters, setFilters] = useState<FilterConfig[]>([]);
  const [filtersLoading, setFiltersLoading] = useState(false);
  const [isFilterCreateOpen, setIsFilterCreateOpen] = useState(false);
  const [newFilterName, setNewFilterName] = useState('');
  const [newFilterAttribute, setNewFilterAttribute] = useState('group_title');
  const [newFilterIncludes, setNewFilterIncludes] = useState('');
  const [newFilterExcludes, setNewFilterExcludes] = useState('');
  const [isFilterCreating, setIsFilterCreating] = useState(false);
  const [filterError, setFilterError] = useState<string | null>(null);

  // Notifications State
  const [notification, setNotification] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  // Show Toast helper
  const showToast = (message: string, type: 'success' | 'error' = 'success') => {
    setNotification({ message, type });
    setTimeout(() => setNotification(null), 4000);
  };

  // Check Health on Load
  const fetchHealth = () => {
    fetch('/health')
      .then(res => res.json())
      .then(data => setHealthStatus(data.status === 'healthy' ? 'healthy' : 'unhealthy'))
      .catch(() => setHealthStatus('unhealthy'));
  };

  // Fetch Stats KPI on Load
  const fetchStats = () => {
    fetch('/api/v1/stats')
      .then(res => res.json())
      .then((data: StatsResponse) => setStats(data))
      .catch(() => {});
  };

  useEffect(() => {
    fetchHealth();
    fetchStats();
    const interval = setInterval(() => {
      fetchHealth();
      fetchStats();
    }, 10000);
    return () => clearInterval(interval);
  }, []);

  // Fetch Config Paths on Load
  useEffect(() => {
    fetch('/api/v1/config/paths')
      .then(res => res.json())
      .then(data => setConfigPaths(data))
      .catch(() => {});
  }, []);

  // Fetch Playlist Items
  const fetchPlaylist = () => {
    setPlaylistLoading(true);
    let url = `/api/v1/items?limit=15&offset=${(playlistPage - 1) * 15}`;
    if (playlistFilter !== 'all') {
      url += `&content_type=${playlistFilter}`;
    }
    if (playlistStateFilter !== 'all') {
      url += `&state=${playlistStateFilter}`;
    }
    if (playlistSearch) {
      url += `&group_title=${encodeURIComponent(playlistSearch)}`;
    }
    fetch(url)
      .then(res => res.json())
      .then((data: PaginatedResponse<PlaylistItem>) => {
        setPlaylist(data.data || []);
        setPlaylistTotal(data.total || 0);
      })
      .catch(() => {})
      .finally(() => setPlaylistLoading(false));
  };

  useEffect(() => {
    fetchPlaylist();
  }, [playlistPage, playlistFilter, playlistStateFilter, playlistSearch]);

  // Fetch Processing Logs
  const fetchLogs = () => {
    setLogsLoading(true);
    fetch('/api/v1/processing-logs?limit=20')
      .then(res => res.json())
      .then((data: PaginatedResponse<ProcessingLog>) => setLogs(data.data || []))
      .catch(() => {})
      .finally(() => setLogsLoading(false));
  };

  useEffect(() => {
    if (activeTab !== 'logs') return;
    fetchLogs();
    const interval = setInterval(fetchLogs, 5000);
    return () => clearInterval(interval);
  }, [activeTab]);

  // Fetch Downloads Info
  const fetchDownloads = () => {
    setDownloadsLoading(true);
    fetch('/api/v1/downloads?limit=20')
      .then(res => res.json())
      .then((data: PaginatedResponse<DownloadInfo>) => setDownloads(data.data || []))
      .catch(() => {})
      .finally(() => setDownloadsLoading(false));
  };

  useEffect(() => {
    if (activeTab !== 'downloads') return;
    fetchDownloads();
    const interval = setInterval(fetchDownloads, 5000);
    return () => clearInterval(interval);
  }, [activeTab]);

  // Fetch Custom Filters
  const fetchFilters = () => {
    setFiltersLoading(true);
    fetch('/api/v1/filters')
      .then(res => res.json())
      .then((data: any) => setFilters(data.filters || []))
      .catch(() => {})
      .finally(() => setFiltersLoading(false));
  };

  useEffect(() => {
    if (activeTab === 'filters') {
      fetchFilters();
    }
  }, [activeTab]);

  // Handle Pipeline Reset
  const handleResetPipeline = (id: number, contentType: string) => {
    const endpoint = contentType === 'movies'
      ? `/api/v1/movies/${id}/reset`
      : `/api/v1/tvshows/${id}/reset`;

    fetch(endpoint, { method: 'POST' })
      .then(res => {
        if (!res.ok) {
          throw new Error('La réinitialisation a échoué');
        }
        return res.json();
      })
      .then(() => {
        showToast('Flux réinitialisé avec succès ! Prêt pour le retraitement.', 'success');
        fetchPlaylist();
        fetchStats();
      })
      .catch(err => {
        showToast(err.message, 'error');
      });
  };

  // Handle Manual TMDB Override Modal
  const handleOpenOverrideModal = (item: PlaylistItem) => {
    setOverrideItemData(item);
    
    // Clean raw tvg_name
    let cleaned = item.tvg_name;
    cleaned = cleaned.replace(/^[a-zA-Z]{2,4}\s*[:\-]\s*/, '');
    cleaned = cleaned.replace(/\b(FHD|UHD|4K|1080p|720p|480p|HD|SD|MULTI|VF|VOSTFR|WEB|x264|H264|x265|HEVC|AAC|AC3)\b/ig, '');
    cleaned = cleaned.replace(/\bS\d+E\d+\b/ig, '');
    cleaned = cleaned.replace(/\bS\d+\b/ig, '');
    cleaned = cleaned.replace(/\bE\d+\b/ig, '');
    cleaned = cleaned.replace(/[\(\)\-\[\]]/g, ' ');
    cleaned = cleaned.trim().replace(/\s+/g, ' ');

    setOverrideSearchQuery(cleaned);

    const isTV = item.content_type === 'tvshows';
    setOverrideMediaType(isTV ? 'tvshow' : 'movie');
    
    // Extract Season & Episode
    const seMatch = item.tvg_name.match(/S(\d+)E(\d+)/i);
    if (seMatch) {
      setOverrideSeason(parseInt(seMatch[1], 10).toString());
      setOverrideEpisode(parseInt(seMatch[2], 10).toString());
    } else {
      const epMatch = item.tvg_name.match(/E(\d+)/i);
      if (epMatch) {
        setOverrideSeason('1');
        setOverrideEpisode(parseInt(epMatch[1], 10).toString());
      } else {
        setOverrideSeason('');
        setOverrideEpisode('');
      }
    }

    const existingYear = isTV ? item.tvshow?.tmdb_year : item.movie?.tmdb_year;
    setOverrideSearchYear(existingYear ? existingYear.toString() : '');
    setSelectedResult(null);
    setOverrideError(null);
    setOverrideSearchResults([]);
    setIsOverrideOpen(true);

    // Auto-trigger search immediately
    setOverrideSearchLoading(true);
    fetch(`/api/v1/tmdb/search?query=${encodeURIComponent(cleaned)}&type=${isTV ? 'tvshow' : 'movie'}&year=${existingYear || ''}`)
      .then(res => {
        if (!res.ok) {
          throw new Error('La recherche a échoué');
        }
        return res.json();
      })
      .then(data => {
        setOverrideSearchResults(data);
      })
      .catch(err => {
        setOverrideError(err.message);
      })
      .finally(() => {
        setOverrideSearchLoading(false);
      });
  };

  const handleSearchTMDB = (queryOverride?: string, typeOverride?: 'movie' | 'tvshow', yearOverride?: string) => {
    const q = queryOverride !== undefined ? queryOverride : overrideSearchQuery;
    const t = typeOverride !== undefined ? typeOverride : overrideMediaType;
    const y = yearOverride !== undefined ? yearOverride : overrideYear;

    if (!q.trim()) return;
    setOverrideSearchLoading(true);
    setOverrideError(null);
    setSelectedResult(null);

    fetch(`/api/v1/tmdb/search?query=${encodeURIComponent(q)}&type=${t}&year=${y}`)
      .then(res => {
        if (!res.ok) {
          throw new Error('La recherche a échoué');
        }
        return res.json();
      })
      .then(data => {
        setOverrideSearchResults(data);
      })
      .catch(err => {
        setOverrideError(err.message);
      })
      .finally(() => {
        setOverrideSearchLoading(false);
      });
  };

  const handleForceOverride = () => {
    if (!overrideItemData || !selectedResult) return;
    setIsSubmittingOverride(true);
    setOverrideError(null);

    const payload = {
      tmdb_id: selectedResult.id,
      type: overrideMediaType,
      season: overrideMediaType === 'tvshow' && overrideSeason ? parseInt(overrideSeason, 10) : null,
      episode: overrideMediaType === 'tvshow' && overrideEpisode ? parseInt(overrideEpisode, 10) : null,
    };

    fetch(`/api/v1/items/${overrideItemData.id}/override`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    })
      .then(res => {
        if (!res.ok) {
          throw new Error("L'association forcée a échoué");
        }
        return res.json();
      })
      .then(() => {
        showToast('Association TMDB mise à jour avec succès !', 'success');
        setIsOverrideOpen(false);
        fetchPlaylist();
        fetchStats();
      })
      .catch(err => {
        setOverrideError(err.message);
      })
      .finally(() => {
        setIsSubmittingOverride(false);
      });
  };

  // Handle Move Complete parent directory
  const handleMoveFolder = () => {
    if (!moveItem) return;
    setIsMoving(true);
    setMoveError(null);
    setMoveSuccess(null);

    const destDir = customDirInput || targetDir;
    if (!destDir) {
      setMoveError('Veuillez spécifier un dossier de destination');
      setIsMoving(false);
      return;
    }

    const endpoint = moveItem.type === 'movie' 
      ? `/api/v1/movies/${moveItem.id}/move`
      : `/api/v1/tvshows/${moveItem.id}/move`;

    fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ destination_parent_dir: destDir }),
    })
      .then(res => {
        if (!res.ok) {
          return res.json().then(err => { throw new Error(err.message || 'Le déplacement a échoué') });
        }
        return res.json();
      })
      .then(() => {
        setMoveSuccess('Le dossier parent a été déplacé avec succès !');
        fetchDownloads(); // Refresh downloads
        fetchStats();     // Refresh stats
        setTimeout(() => {
          setIsMoveOpen(false);
          setMoveItem(null);
          setCustomDirInput('');
          setMoveSuccess(null);
        }, 2000);
      })
      .catch(err => {
        setMoveError(err.message);
      })
      .finally(() => setIsMoving(false));
  };

  const openMoveDialog = (item: DownloadInfo) => {
    // Find matching movie or tv show from loaded playlist if possible, or build basic metadata
    let moveMeta: { id: number; title: string; type: 'movie' | 'tvshow'; currentPath?: string } = {
      id: 0,
      title: 'Inconnu',
      type: 'movie',
      currentPath: item.download_path,
    };
    
    // Look in playlist for matching items
    const match = playlist.find(p => p.movie && p.id === item.id) || playlist.find(p => p.tvshow && p.id === item.id);
    if (match) {
      if (match.content_type === 'movies' && match.movie) {
        moveMeta = { id: match.movie.id, title: match.movie.tmdb_title, type: 'movie', currentPath: item.download_path };
      } else if (match.content_type === 'tvshows' && match.tvshow) {
        moveMeta = { id: match.tvshow.id, title: match.tvshow.tmdb_title, type: 'tvshow', currentPath: item.download_path };
      }
    } else {
      // Fallback: extract title from download path
      const folderName = item.download_path ? filepathBase(item.download_path) : 'Média';
      moveMeta = { id: item.id, title: folderName, type: item.download_path?.includes('tvshows') ? 'tvshow' : 'movie', currentPath: item.download_path };
    }

    setMoveItem(moveMeta);
    setTargetDir(moveMeta.type === 'movie' ? configPaths?.movies_path || '' : configPaths?.tvshows_path || '');
    setIsMoveOpen(true);
  };

  // Handle Filter Creation
  const handleCreateFilter = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newFilterName.trim()) {
      setFilterError('Le nom du filtre est requis.');
      return;
    }

    setIsFilterCreating(true);
    setFilterError(null);

    const payload = {
      name: newFilterName,
      attribute: newFilterAttribute,
      include_patterns: newFilterIncludes.trim() ? newFilterIncludes : undefined,
      exclude_patterns: newFilterExcludes.trim() ? newFilterExcludes : undefined,
    };

    fetch('/api/v1/filters', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
      .then(res => {
        if (!res.ok) {
          return res.json().then(err => { throw new Error(err.message || 'La création du filtre a échoué') });
        }
        return res.json();
      })
      .then(() => {
        showToast('Nouveau filtre de tri configuré avec succès !', 'success');
        setIsFilterCreateOpen(false);
        setNewFilterName('');
        setNewFilterIncludes('');
        setNewFilterExcludes('');
        fetchFilters();
      })
      .catch(err => {
        setFilterError(err.message);
      })
      .finally(() => setIsFilterCreating(false));
  };

  // Handle Filter Deletion
  const handleDeleteFilter = (id: number) => {
    if (!confirm('Êtes-vous sûr de vouloir supprimer définitivement ce filtre de tri ?')) {
      return;
    }

    fetch(`/api/v1/filters/${id}`, { method: 'DELETE' })
      .then(res => {
        if (!res.ok) {
          throw new Error('La suppression du filtre a échoué');
        }
        return res.json();
      })
      .then(() => {
        showToast('Filtre supprimé avec succès !', 'success');
        fetchFilters();
      })
      .catch(err => {
        showToast(err.message, 'error');
      });
  };

  const filepathBase = (path: string) => {
    const parts = path.split('/');
    return parts[parts.length - 1];
  };

  // Calculate download success percentage
  const getDownloadSuccessRatio = () => {
    if (!stats || !stats.by_state) return '100%';
    const downloaded = stats.by_state.downloaded || 0;
    const failed = stats.by_state.failed || 0;
    const total = downloaded + failed;
    if (total === 0) return '100%';
    return `${Math.round((downloaded / total) * 100)}%`;
  };

  return (
    <div style={{ maxWidth: 1600, margin: '0 auto', padding: '2rem 1.5rem' }}>
      
      {/* Toast Notification */}
      {notification && (
        <div style={{
          position: 'fixed',
          bottom: '24px',
          right: '24px',
          backgroundColor: notification.type === 'success' ? 'var(--status-success-bg)' : 'var(--status-failed-bg)',
          color: notification.type === 'success' ? 'var(--status-success-text)' : 'var(--status-failed-text)',
          border: `1px solid ${notification.type === 'success' ? 'var(--status-success-border)' : 'var(--status-failed-border)'}`,
          padding: '1rem 1.5rem',
          borderRadius: 'var(--radius-md)',
          boxShadow: 'var(--shadow-lg)',
          fontWeight: 600,
          fontSize: '0.875rem',
          zIndex: 1000,
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          animation: 'contentShow 150ms ease-out'
        }}>
          {notification.type === 'success' ? '✓' : '⚠️'} {notification.message}
        </div>
      )}

      {/* Modern Floating Header */}
      <header className="glass-header">
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.75rem', color: 'var(--primary-slate)' }}>
            🛰️ Stalkeer Dashboard
          </h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem', fontWeight: 500 }}>IHM de pilotage, logs, filtres et réorganisation de médias</p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <span className={`badge ${healthStatus === 'healthy' ? 'badge-success' : healthStatus === 'unhealthy' ? 'badge-failed' : 'badge-pending'}`} style={{ gap: '0.5rem' }}>
            <span className="pulse-dot" style={{ display: healthStatus === 'healthy' ? 'inline-block' : 'none' }}></span>
            API: {healthStatus === 'healthy' ? 'En ligne' : healthStatus === 'unhealthy' ? 'Erreur' : 'Vérification...'}
          </span>
        </div>
      </header>

      {/* Global Stats KPI Cards Section */}
      <section className="kpi-grid">
        <div className="kpi-card">
          <span className="kpi-card-title">🗒️ Playlist M3U</span>
          <span className="kpi-card-value">{stats ? stats.total_items.toLocaleString() : '...'}</span>
          <span className="kpi-card-subtitle">Flux enregistrés en DB</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-card-title">🎬 Films identifiés</span>
          <span className="kpi-card-value">{stats && stats.by_content_type ? (stats.by_content_type.movies || 0).toLocaleString() : '...'}</span>
          <span className="kpi-card-subtitle">TMDB enrichis</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-card-title">📺 Séries suivies</span>
          <span className="kpi-card-value">{stats && stats.by_content_type ? (stats.by_content_type.tvshows || 0).toLocaleString() : '...'}</span>
          <span className="kpi-card-subtitle">Saisons et épisodes</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-card-title">📥 Taux de Réussite</span>
          <span className="kpi-card-value">{stats ? getDownloadSuccessRatio() : '...'}</span>
          <span className="kpi-card-subtitle">
            {stats && stats.by_state ? `${stats.by_state.downloaded || 0} réussis / ${stats.by_state.failed || 0} échecs` : 'Téléchargements'}
          </span>
        </div>
      </section>

      {/* Tabs Navigation (Radix UI) */}
      <Tabs.Root value={activeTab} onValueChange={setActiveTab}>
        <Tabs.List className="segmented-tabs-list">
          <Tabs.Trigger value="playlist" className="segmented-tabs-trigger">
            🗒️ Playlist M3U
          </Tabs.Trigger>
          <Tabs.Trigger value="filters" className="segmented-tabs-trigger">
            🔍 Filtres de Tri
          </Tabs.Trigger>
          <Tabs.Trigger value="logs" className="segmented-tabs-trigger">
            ⚙️ Traitements
          </Tabs.Trigger>
          <Tabs.Trigger value="downloads" className="segmented-tabs-trigger">
            📥 Téléchargements
          </Tabs.Trigger>
        </Tabs.List>

        {/* Tab 1: Playlist Manager */}
        <Tabs.Content value="playlist" className="card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
            
            {/* Filter Content Type Buttons */}
            <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
              <button onClick={() => { setPlaylistFilter('all'); setPlaylistPage(1); }} className={playlistFilter === 'all' ? 'btn-primary' : 'btn-secondary'} style={{ padding: '0.45rem 1rem' }}>Tous</button>
              <button onClick={() => { setPlaylistFilter('movies'); setPlaylistPage(1); }} className={playlistFilter === 'movies' ? 'btn-primary' : 'btn-secondary'} style={{ padding: '0.45rem 1rem' }}>Films</button>
              <button onClick={() => { setPlaylistFilter('tvshows'); setPlaylistPage(1); }} className={playlistFilter === 'tvshows' ? 'btn-primary' : 'btn-secondary'} style={{ padding: '0.45rem 1rem' }}>Séries</button>
              
              {/* Vertical Divider */}
              <div style={{ width: '1px', background: 'var(--border-color)', margin: '0 0.5rem' }} />
              
              {/* Filter State Selection */}
              <select value={playlistStateFilter} onChange={e => { setPlaylistStateFilter(e.target.value); setPlaylistPage(1); }} className="custom-select" style={{ width: 'auto', padding: '0.4rem 1.5rem 0.4rem 1rem', height: '100%' }}>
                <option value="all">Tous les Statuts</option>
                <option value="processed">Traité (Processed)</option>
                <option value="pending">En Attente (Pending)</option>
                <option value="downloading">En Cours (Downloading)</option>
                <option value="downloaded">Téléchargé (Downloaded)</option>
                <option value="failed">Échoué (Failed)</option>
              </select>
            </div>

            {/* Title search */}
            <div style={{ display: 'flex', gap: '0.5rem', flex: 1, maxWidth: '400px' }}>
              <input type="text" placeholder="Rechercher par Groupe / VOD..." value={playlistSearch} onChange={e => { setPlaylistSearch(e.target.value); setPlaylistPage(1); }} className="custom-input" />
            </div>
          </div>

          <div style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Nom du Média</th>
                  <th>Groupe / Catégorie</th>
                  <th>Enrichissement TMDB</th>
                  <th>État Pipeline</th>
                  <th>Créé le</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {playlistLoading ? (
                  <tr>
                    <td colSpan={6} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
                      <span style={{ fontWeight: 600 }}>Chargement de la playlist en cours...</span>
                    </td>
                  </tr>
                ) : playlist.length === 0 ? (
                  <tr>
                    <td colSpan={6} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Aucun item trouvé</td>
                  </tr>
                ) : (
                  playlist.map(item => {
                    const isMovie = item.content_type === 'movies';
                    const tmdb = isMovie ? item.movie : item.tvshow;
                    return (
                      <tr key={item.id}>
                        <td style={{ fontWeight: 600, color: 'var(--primary-slate)' }}>{item.tvg_name}</td>
                        <td style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{item.group_title}</td>
                        <td>
                          {tmdb ? (
                            <div>
                              <strong style={{ color: 'var(--primary-accent)' }}>{tmdb.tmdb_title}</strong>{' '}
                              <span style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', fontWeight: 500 }}>({tmdb.tmdb_year})</span>
                            </div>
                          ) : (
                            <span style={{ fontStyle: 'italic', color: 'var(--text-muted)' }}>Non enrichi</span>
                          )}
                        </td>
                        <td>
                          <span className={`badge ${
                            item.state === 'downloaded' ? 'badge-success' : 
                            item.state === 'downloading' ? 'badge-progress' : 
                            item.state === 'failed' ? 'badge-failed' : 'badge-pending'
                          }`}>
                            {item.state}
                          </span>
                        </td>
                        <td style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{new Date(item.created_at).toLocaleDateString()}</td>
                        <td style={{ textAlign: 'right' }}>
                          <div style={{ display: 'flex', gap: '0.35rem', justifyContent: 'flex-end' }}>
                            <button
                              onClick={() => handleOpenOverrideModal(item)}
                              className="btn-primary"
                              style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}
                              title="Corriger ou forcer manuellement l'association TMDB"
                            >
                              {item.override_by ? 'Corriger ✏️' : 'Associer 🔍'}
                            </button>
                            <button 
                              onClick={() => handleResetPipeline(item.id, item.content_type)} 
                              className="btn-secondary" 
                              style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}
                              title="Réinitialiser l'état du média pour forcer le retraitement"
                            >
                              Réinitialiser ↻
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

          {/* Pagination */}
          {playlistTotal > 15 && (
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1.5rem' }}>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>Affichage de 15 sur {playlistTotal} entrées</span>
              <div style={{ display: 'flex', gap: '0.35rem' }}>
                <button disabled={playlistPage === 1} onClick={() => setPlaylistPage(p => p - 1)} className="btn-secondary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem', opacity: playlistPage === 1 ? 0.5 : 1 }}>Précédent</button>
                <button disabled={playlistPage * 15 >= playlistTotal} onClick={() => setPlaylistPage(p => p + 1)} className="btn-secondary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem', opacity: playlistPage * 15 >= playlistTotal ? 0.5 : 1 }}>Suivant</button>
              </div>
            </div>
          )}
        </Tabs.Content>

        {/* Tab 2: Custom Filters Management */}
        <Tabs.Content value="filters" className="card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <div>
              <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>Expressions régulières d'Inclusion & d'Exclusion</h2>
              <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>Gérez les patterns de filtrage appliqués aux flux M3U lors de l'enrichissement</p>
            </div>
            <button onClick={() => setIsFilterCreateOpen(true)} className="btn-primary">
              + Configurer un Filtre
            </button>
          </div>

          {filtersLoading && filters.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>Chargement des filtres...</div>
          ) : filters.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '4rem', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)' }}>
              <p style={{ color: 'var(--text-secondary)', fontWeight: 600 }}>Aucun filtre de tri configuré</p>
              <p style={{ color: 'var(--text-muted)', fontSize: '0.8rem', marginTop: '0.25rem' }}>Tous les flux M3U importés seront acceptés par défaut</p>
            </div>
          ) : (
            <div className="filter-grid">
              {filters.map(filter => (
                <div key={filter.id} className="filter-card">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <div>
                      <h3 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{filter.name}</h3>
                      <span className="badge badge-progress" style={{ marginTop: '0.4rem', fontSize: '0.65rem', padding: '0.15rem 0.5rem' }}>
                        Attribut: {filter.attribute}
                      </span>
                    </div>
                    <button onClick={() => handleDeleteFilter(filter.id)} className="btn-danger" style={{ padding: '0.3rem 0.5rem' }} title="Supprimer ce filtre">
                      Supprimer 🗑️
                    </button>
                  </div>
                  
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.8rem', borderTop: '1px solid var(--border-color)', paddingTop: '0.75rem' }}>
                    <div>
                      <strong style={{ color: 'var(--status-success-text)' }}>🟢 Inclure : </strong>
                      <code style={{ background: 'var(--status-success-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                        {filter.include_patterns || '* (Tout accepter)'}
                      </code>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--status-failed-text)' }}>🔴 Exclure : </strong>
                      <code style={{ background: 'var(--status-failed-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                        {filter.exclude_patterns || '- (Aucun)'}
                      </code>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Tabs.Content>

        {/* Tab 3: Processing Logs */}
        <Tabs.Content value="logs" className="card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <div>
              <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>Dernières Exécutions de Tâches</h2>
              <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>Historique des rafraîchissements M3U, nettoyages de base et enrichissements</p>
            </div>
            <button onClick={fetchLogs} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>Actualiser ↻</button>
          </div>

          <div style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
            <table className="custom-table">
              <thead>
                <tr>
                  <th>Action exécutée</th>
                  <th>Éléments Traités</th>
                  <th>Statut Pipeline</th>
                  <th>Lancement</th>
                  <th>Détails d'Erreur</th>
                </tr>
              </thead>
              <tbody>
                {logsLoading && logs.length === 0 ? (
                  <tr>
                    <td colSpan={5} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Chargement en cours...</td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td colSpan={5} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Aucun log enregistré en base de données</td>
                  </tr>
                ) : (
                  logs.map(log => (
                    <tr key={log.id}>
                      <td style={{ fontWeight: 700, color: 'var(--primary-slate)' }}>{log.action}</td>
                      <td style={{ fontWeight: 600 }}>{log.item_count} items</td>
                      <td>
                        <span className={`badge ${log.status === 'success' ? 'badge-success' : log.status === 'failed' ? 'badge-failed' : 'badge-progress'}`}>
                          {log.status === 'in_progress' ? 'En cours' : log.status === 'success' ? 'Réussi' : 'Échoué'}
                        </span>
                      </td>
                      <td style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{new Date(log.started_at).toLocaleString()}</td>
                      <td style={{ maxWidth: '300px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {log.error_message ? (
                          <span style={{ color: 'var(--status-failed-text)', fontWeight: 600 }} title={log.error_message}>{log.error_message}</span>
                        ) : (
                          <span style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>Aucune erreur</span>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Tabs.Content>

        {/* Tab 4: Downloads Tracker */}
        <Tabs.Content value="downloads" className="card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <div>
              <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>File de Téléchargement et Organisation</h2>
              <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>Suivi dynamique des téléchargements de fichiers vidéo et réorganisations</p>
            </div>
            <button onClick={fetchDownloads} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>Actualiser ↻</button>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {downloadsLoading && downloads.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>Chargement des téléchargements...</div>
            ) : downloads.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '4rem', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)', color: 'var(--text-secondary)' }}>Aucun téléchargement enregistré</div>
            ) : (
              downloads.map(item => {
                const total = item.total_bytes || 0;
                const downloaded = item.bytes_downloaded || 0;
                const progress = total > 0 ? Math.round((downloaded / total) * 100) : 0;
                const isCompleted = item.status === 'completed';

                return (
                  <div key={item.id} style={{ border: '1px solid var(--border-color)', borderRadius: 'var(--radius-md)', padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '0.75rem', backgroundColor: '#fff', boxShadow: 'var(--shadow-sm)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '0.5rem' }}>
                      <div style={{ flex: 1, minWidth: '250px' }}>
                        <h3 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--primary-slate)', wordBreak: 'break-all' }}>{item.download_path ? filepathBase(item.download_path) : item.url}</h3>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '0.75rem', marginTop: '0.25rem', wordBreak: 'break-all', fontWeight: 500 }}>{item.url}</p>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <span className={`badge ${item.status === 'completed' ? 'badge-success' : item.status === 'downloading' ? 'badge-progress' : item.status === 'failed' ? 'badge-failed' : 'badge-pending'}`}>
                          {item.status}
                        </span>
                        {isCompleted && (
                          <button onClick={() => openMoveDialog(item)} className="btn-primary" style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}>
                            Déplacer ⇄
                          </button>
                        )}
                      </div>
                    </div>

                    {/* Progress Bar for Downloading / Retrying */}
                    {!isCompleted && item.status !== 'failed' && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>
                          <span>Progression: {progress}%</span>
                          {item.file_size && (
                            <span>{(downloaded / 1024 / 1024).toFixed(1)} Mo / {(item.file_size / 1024 / 1024).toFixed(1)} Mo</span>
                          )}
                        </div>
                        {/* Radix UI Progress Bar */}
                        <Progress.Root value={progress} className="progress-root">
                          <Progress.Indicator className="progress-indicator" style={{ width: `${progress}%` }} />
                        </Progress.Root>
                      </div>
                    )}

                    {item.error_message && (
                      <div style={{ padding: '0.75rem', backgroundColor: 'var(--status-failed-bg)', borderRadius: 'var(--radius-sm)', color: 'var(--status-failed-text)', fontSize: '0.8rem', fontWeight: 600, border: '1px solid var(--status-failed-border)' }}>
                        {item.error_message}
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </Tabs.Content>
      </Tabs.Root>

      {/* Accessible Move Dialog (Radix UI) */}
      <Dialog.Root open={isMoveOpen} onOpenChange={setIsMoveOpen}>
        <Dialog.Portal>
          <Dialog.Overlay className="dialog-overlay" />
          <Dialog.Content className="dialog-content">
            <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--primary-slate)', marginBottom: '1rem' }}>
              ⇄ Déplacer le dossier de l'œuvre
            </Dialog.Title>
            <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.5, marginBottom: '1.5rem', fontWeight: 500 }}>
              Cette action déplacera l'<strong>intégralité du dossier parent</strong> (films ou séries complètes incluant toutes les saisons) vers un nouveau répertoire de stockage et mettra à jour la base de données.
            </Dialog.Description>

            {moveItem && (
              <div style={{ backgroundColor: 'var(--bg-app)', padding: '1rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', marginBottom: '1.5rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.85rem' }}>
                <div><strong>📁 Œuvre :</strong> {moveItem.title} <span className="badge badge-pending" style={{ padding: '0.1rem 0.4rem', fontSize: '0.65rem' }}>{moveItem.type}</span></div>
                <div><strong>📌 Chemin actuel :</strong> <code style={{ wordBreak: 'break-all', color: 'var(--text-secondary)' }}>{moveItem.currentPath || 'Inconnu'}</code></div>
              </div>
            )}

            {/* Input Selection */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Dossier de destination parent :</label>
                <select value={targetDir} onChange={e => { setTargetDir(e.target.value); setCustomDirInput(''); }} className="custom-select">
                  <option value="">-- Sélectionner un répertoire racine --</option>
                  {configPaths && (
                    <>
                      <option value={configPaths.movies_path}>Dossier des Films ({configPaths.movies_path})</option>
                      <option value={configPaths.tvshows_path}>Dossier des Séries ({configPaths.tvshows_path})</option>
                    </>
                  )}
                  <option value="custom">Autre dossier personnalisé...</option>
                </select>
              </div>

              {(targetDir === 'custom' || !configPaths) && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                  <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Saisir un chemin de destination personnalisé :</label>
                  <input type="text" placeholder="Exemple: /media/enfants-movies" value={customDirInput} onChange={e => setCustomDirInput(e.target.value)} className="custom-input" />
                </div>
              )}
            </div>

            {moveError && (
              <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600, marginBottom: '1.5rem', border: '1px solid var(--status-failed-border)' }}>
                ⚠️ {moveError}
              </div>
            )}

            {moveSuccess && (
              <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-success-bg)', color: 'var(--status-success-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600, marginBottom: '1.5rem', border: '1px solid var(--status-success-border)' }}>
                ✓ {moveSuccess}
              </div>
            )}

            {/* Actions */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
              <button disabled={isMoving} onClick={() => { setIsMoveOpen(false); setMoveItem(null); setCustomDirInput(''); }} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
                Annuler
              </button>
              <button disabled={isMoving} onClick={handleMoveFolder} className="btn-primary" style={{ padding: '0.5rem 1.25rem' }}>
                {isMoving ? 'Déplacement...' : '✓ Confirmer'}
              </button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>

      {/* Accessible Manual Override Dialog (Radix UI) */}
      <Dialog.Root open={isOverrideOpen} onOpenChange={setIsOverrideOpen}>
        <Dialog.Portal>
          <Dialog.Overlay className="dialog-overlay" />
          <Dialog.Content className="dialog-content" style={{ maxWidth: '600px', display: 'flex', flexDirection: 'column', maxHeight: '85vh' }}>
            <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--primary-slate)', marginBottom: '0.5rem' }}>
              🔍 Correction Manuelle TMDB
            </Dialog.Title>
            <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.4, marginBottom: '1rem', fontWeight: 500 }}>
              Recherchez et associez manuellement cet item de la playlist à une œuvre officielle TMDB.
            </Dialog.Description>

            {overrideItemData && (
              <div style={{ backgroundColor: 'var(--bg-app)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', marginBottom: '1rem', display: 'flex', flexDirection: 'column', gap: '0.35rem', fontSize: '0.8rem' }}>
                <div><strong>🏷️ Titre brut :</strong> <span style={{ fontFamily: 'monospace', color: 'var(--text-secondary)' }}>{overrideItemData.tvg_name}</span></div>
                <div><strong>📁 Groupe d'origine :</strong> <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{overrideItemData.group_title}</span></div>
              </div>
            )}

            {/* MediaType Select & Year */}
            <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
              <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <label style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Type de média cible :</label>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button 
                    type="button" 
                    onClick={() => { setOverrideMediaType('movie'); handleSearchTMDB(overrideSearchQuery, 'movie', overrideYear); }} 
                    className={overrideMediaType === 'movie' ? 'btn-primary' : 'btn-secondary'} 
                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem' }}
                  >
                    🎬 Film
                  </button>
                  <button 
                    type="button" 
                    onClick={() => { setOverrideMediaType('tvshow'); handleSearchTMDB(overrideSearchQuery, 'tvshow', overrideYear); }} 
                    className={overrideMediaType === 'tvshow' ? 'btn-primary' : 'btn-secondary'} 
                    style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem' }}
                  >
                    📺 Série TV
                  </button>
                </div>
              </div>
              
              {overrideMediaType === 'movie' && (
                <div style={{ width: '120px', display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                  <label style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Année (film) :</label>
                  <input 
                    type="number" 
                    placeholder="Ex: 2010" 
                    value={overrideYear} 
                    onChange={e => { setOverrideSearchYear(e.target.value); handleSearchTMDB(overrideSearchQuery, overrideMediaType, e.target.value); }} 
                    className="custom-input" 
                    style={{ padding: '0.4rem', fontSize: '0.8rem' }}
                  />
                </div>
              )}
            </div>

            {/* Search Input */}
            <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
              <input 
                type="text" 
                placeholder="Rechercher sur TMDB..." 
                value={overrideSearchQuery} 
                onChange={e => setOverrideSearchQuery(e.target.value)} 
                onKeyDown={e => { if (e.key === 'Enter') handleSearchTMDB(); }}
                className="custom-input" 
                style={{ flex: 1 }}
              />
              <button 
                type="button" 
                onClick={() => handleSearchTMDB()} 
                className="btn-primary" 
                style={{ padding: '0.5rem 1rem' }}
                disabled={overrideSearchLoading}
              >
                Rechercher
              </button>
            </div>

            {/* Search Results Area */}
            <div style={{ flex: 1, overflowY: 'auto', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', padding: '0.5rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', backgroundColor: 'var(--bg-app)', minHeight: '180px', maxHeight: '300px', marginBottom: '1rem' }}>
              {overrideSearchLoading ? (
                <div style={{ margin: 'auto', color: 'var(--text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>Recherche TMDB en cours...</div>
              ) : overrideSearchResults.length === 0 ? (
                <div style={{ margin: 'auto', color: 'var(--text-muted)', fontSize: '0.8rem', fontStyle: 'italic' }}>Aucun résultat trouvé. Ajustez la recherche.</div>
              ) : (
                overrideSearchResults.map(res => {
                  const isSelected = selectedResult && selectedResult.id === res.id;
                  return (
                    <div 
                      key={res.id} 
                      onClick={() => setSelectedResult(res)}
                      style={{ 
                        display: 'flex', 
                        gap: '0.75rem', 
                        padding: '0.5rem', 
                        borderRadius: 'var(--radius-sm)', 
                        border: isSelected ? '2px solid var(--primary-accent)' : '1px solid var(--border-color)',
                        backgroundColor: isSelected ? 'rgba(var(--primary-accent-rgb), 0.08)' : 'var(--card-bg)',
                        cursor: 'pointer',
                        transition: 'all 0.15s ease'
                      }}
                    >
                      {res.poster_path ? (
                        <img 
                          src={`https://image.tmdb.org/t/p/w92${res.poster_path}`} 
                          alt={res.title} 
                          style={{ width: '48px', height: '72px', objectFit: 'cover', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)' }}
                        />
                      ) : (
                        <div style={{ width: '48px', height: '72px', backgroundColor: 'var(--border-color)', borderRadius: 'var(--radius-sm)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.65rem', color: 'var(--text-muted)', textAlign: 'center', padding: '0.25rem' }}>Pas d'image</div>
                      )}
                      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                          <strong style={{ fontSize: '0.85rem', color: 'var(--primary-slate)' }}>{res.title}</strong>
                          {res.release_date && (
                            <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', fontWeight: 600 }}>({res.release_date.substring(0, 4)})</span>
                          )}
                        </div>
                        <p style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden', margin: 0 }}>{res.overview || "Pas de description disponible."}</p>
                      </div>
                    </div>
                  );
                })
              )}
            </div>

            {/* Optional Episode Fields for Series */}
            {overrideMediaType === 'tvshow' && selectedResult && (
              <div style={{ backgroundColor: 'var(--bg-app)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', marginBottom: '1rem', display: 'flex', gap: '1rem' }}>
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                  <label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Saison (optionnel) :</label>
                  <input type="number" placeholder="Ex: 1" value={overrideSeason} onChange={e => setOverrideSeason(e.target.value)} className="custom-input" style={{ padding: '0.35rem', fontSize: '0.8rem' }} />
                </div>
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                  <label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Épisode (optionnel) :</label>
                  <input type="number" placeholder="Ex: 5" value={overrideEpisode} onChange={e => setOverrideEpisode(e.target.value)} className="custom-input" style={{ padding: '0.35rem', fontSize: '0.8rem' }} />
                </div>
              </div>
            )}

            {overrideError && (
              <div style={{ padding: '0.5rem 0.75rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.8rem', fontWeight: 600, marginBottom: '1rem', border: '1px solid var(--status-failed-border)' }}>
                ⚠️ {overrideError}
              </div>
            )}

            {/* Actions */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: 'auto', paddingTop: '0.5rem' }}>
              <button disabled={isSubmittingOverride} onClick={() => { setIsOverrideOpen(false); setOverrideItemData(null); }} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
                Annuler
              </button>
              <button disabled={isSubmittingOverride || !selectedResult} onClick={handleForceOverride} className="btn-primary" style={{ padding: '0.5rem 1.25rem' }}>
                {isSubmittingOverride ? 'Association...' : '✓ Forcer l\'association'}
              </button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>

      {/* Accessible Create Filter Dialog (Radix UI) */}
      <Dialog.Root open={isFilterCreateOpen} onOpenChange={setIsFilterCreateOpen}>
        <Dialog.Portal>
          <Dialog.Overlay className="dialog-overlay" />
          <Dialog.Content className="dialog-content">
            <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--primary-slate)', marginBottom: '1rem' }}>
              🔍 Configurer un Nouveau Filtre de Tri
            </Dialog.Title>
            <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.5, marginBottom: '1.5rem', fontWeight: 500 }}>
              Ajoutez une expression régulière d'Inclusion ou d'Exclusion pour accepter ou rejeter automatiquement les flux de votre playlist M3U lors de l'import.
            </Dialog.Description>

            <form onSubmit={handleCreateFilter} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Nom explicite du filtre :</label>
                <input type="text" placeholder="Ex: Filtre Films Francophones" value={newFilterName} onChange={e => setNewFilterName(e.target.value)} className="custom-input" required />
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Attribut ciblé du média :</label>
                <select value={newFilterAttribute} onChange={e => setNewFilterAttribute(e.target.value)} className="custom-select">
                  <option value="group_title">Group Title (Ex: VOD-FR, SERIES-US)</option>
                  <option value="tvg_name">TVG Name (Ex: Inception.2010.mkv)</option>
                </select>
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Patterns d'Inclusion (Séparés par des virgules) :</label>
                <input type="text" placeholder="Ex: FRENCH, TRUEFRENCH, VFF" value={newFilterIncludes} onChange={e => setNewFilterIncludes(e.target.value)} className="custom-input" />
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Patterns d'Exclusion (Séparés par des virgules) :</label>
                <input type="text" placeholder="Ex: VOSTFR, SUBBED, HDLight" value={newFilterExcludes} onChange={e => setNewFilterExcludes(e.target.value)} className="custom-input" />
              </div>

              {filterError && (
                <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600, border: '1px solid var(--status-failed-border)' }}>
                  ⚠️ {filterError}
                </div>
              )}

              {/* Actions */}
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
                <button type="button" disabled={isFilterCreating} onClick={() => { setIsFilterCreateOpen(false); setFilterError(null); }} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
                  Annuler
                </button>
                <button type="submit" disabled={isFilterCreating} className="btn-primary" style={{ padding: '0.5rem 1.25rem' }}>
                  {isFilterCreating ? 'Enregistrement...' : '✓ Enregistrer le Filtre'}
                </button>
              </div>
            </form>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>

    </div>
  );
}
