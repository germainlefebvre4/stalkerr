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

export default function App() {
  const [activeTab, setActiveTab] = useState('playlist');
  const [healthStatus, setHealthStatus] = useState<'healthy' | 'unhealthy' | 'checking'>('checking');
  
  // Playlist State
  const [playlist, setPlaylist] = useState<PlaylistItem[]>([]);
  const [playlistSearch, setPlaylistSearch] = useState('');
  const [playlistFilter, setPlaylistFilter] = useState<'all' | 'movies' | 'tvshows'>('all');
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

  // Check Health on Load
  useEffect(() => {
    fetch('/health')
      .then(res => res.json())
      .then(data => setHealthStatus(data.status === 'healthy' ? 'healthy' : 'unhealthy'))
      .catch(() => setHealthStatus('unhealthy'));
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
  }, [playlistPage, playlistFilter, playlistSearch]);

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
    if (activeTab === 'logs') {
      fetchLogs();
      const interval = setInterval(fetchLogs, 5000);
      return () => clearInterval(interval);
    }
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
    if (activeTab === 'downloads') {
      fetchDownloads();
      const interval = setInterval(fetchDownloads, 5000);
      return () => clearInterval(interval);
    }
  }, [activeTab]);

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
        fetchDownloads(); // Refresh list
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
    let moveMeta = { id: 0, title: 'Inconnu', type: 'movie' as const, currentPath: item.download_path };
    
    // We look in playlist for matching items
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

  const filepathBase = (path: string) => {
    const parts = path.split('/');
    return parts[parts.length - 1];
  };

  return (
    <div style={{ maxWidth: 1600, margin: '0 auto', padding: '2rem 1rem' }}>
      
      {/* Header */}
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2.5rem', backgroundColor: 'var(--bg-card)', padding: '1.25rem 2rem', borderRadius: 'var(--radius-lg)', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
        <div>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.75rem', color: 'var(--primary-slate)' }}>
            🛰️ Stalkeer Dashboard
          </h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '0.25rem' }}>IHM de pilotage, logs et réorganisation de médias</p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <span className={`badge ${healthStatus === 'healthy' ? 'badge-success' : healthStatus === 'unhealthy' ? 'badge-failed' : 'badge-pending'}`}>
            API: {healthStatus === 'healthy' ? 'En ligne' : healthStatus === 'unhealthy' ? 'Erreur de connexion' : 'Vérification...'}
          </span>
        </div>
      </header>

      {/* Tabs Navigation (Radix UI) */}
      <Tabs.Root value={activeTab} onValueChange={setActiveTab}>
        <Tabs.List style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem' }}>
          <Tabs.Trigger value="playlist" style={{ padding: '0.75rem 1.25rem', fontSize: '0.95rem', fontWeight: 600, color: activeTab === 'playlist' ? 'var(--primary-accent)' : 'var(--text-secondary)', borderBottom: activeTab === 'playlist' ? '2px solid var(--primary-accent)' : '2px solid transparent', transition: 'all 0.2s' }}>
            🗒️ Playlist M3U
          </Tabs.Trigger>
          <Tabs.Trigger value="logs" style={{ padding: '0.75rem 1.25rem', fontSize: '0.95rem', fontWeight: 600, color: activeTab === 'logs' ? 'var(--primary-accent)' : 'var(--text-secondary)', borderBottom: activeTab === 'logs' ? '2px solid var(--primary-accent)' : '2px solid transparent', transition: 'all 0.2s' }}>
            ⚙️ Traitements en cours
          </Tabs.Trigger>
          <Tabs.Trigger value="downloads" style={{ padding: '0.75rem 1.25rem', fontSize: '0.95rem', fontWeight: 600, color: activeTab === 'downloads' ? 'var(--primary-accent)' : 'var(--text-secondary)', borderBottom: activeTab === 'downloads' ? '2px solid var(--primary-accent)' : '2px solid transparent', transition: 'all 0.2s' }}>
            📥 Téléchargements
          </Tabs.Trigger>
        </Tabs.List>

        {/* Tab 1: Playlist Manager */}
        <Tabs.Content value="playlist" className="card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button onClick={() => { setPlaylistFilter('all'); setPlaylistPage(1); }} style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', fontWeight: 500, borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', backgroundColor: playlistFilter === 'all' ? 'var(--primary-accent)' : 'var(--bg-card)', color: playlistFilter === 'all' ? '#fff' : 'var(--text-main)' }}>Tous</button>
              <button onClick={() => { setPlaylistFilter('movies'); setPlaylistPage(1); }} style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', fontWeight: 500, borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', backgroundColor: playlistFilter === 'movies' ? 'var(--primary-accent)' : 'var(--bg-card)', color: playlistFilter === 'movies' ? '#fff' : 'var(--text-main)' }}>Films</button>
              <button onClick={() => { setPlaylistFilter('tvshows'); setPlaylistPage(1); }} style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', fontWeight: 500, borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', backgroundColor: playlistFilter === 'tvshows' ? 'var(--primary-accent)' : 'var(--bg-card)', color: playlistFilter === 'tvshows' ? '#fff' : 'var(--text-main)' }}>Séries</button>
            </div>
            <div style={{ display: 'flex', gap: '0.5rem', flex: 1, maxWidth: '400px' }}>
              <input type="text" placeholder="Rechercher par Groupe / VOD..." value={playlistSearch} onChange={e => { setPlaylistSearch(e.target.value); setPlaylistPage(1); }} style={{ width: '100%', padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', backgroundColor: '#fff' }} />
            </div>
          </div>

          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid var(--border-color)', paddingBottom: '0.75rem', color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                  <th style={{ padding: '0.75rem 1rem' }}>Nom du Média</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Groupe / Catégorie</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Enrichissement TMDB</th>
                  <th style={{ padding: '0.75rem 1rem' }}>État Pipeline</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Créé le</th>
                </tr>
              </thead>
              <tbody>
                {playlistLoading ? (
                  <tr>
                    <td colSpan={5} style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Chargement en cours...</td>
                  </tr>
                ) : playlist.length === 0 ? (
                  <tr>
                    <td colSpan={5} style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Aucun item trouvé</td>
                  </tr>
                ) : (
                  playlist.map(item => {
                    const isMovie = item.content_type === 'movies';
                    const tmdb = isMovie ? item.movie : item.tvshow;
                    return (
                      <tr key={item.id} style={{ borderBottom: '1px solid var(--border-color)', fontSize: '0.9rem' }}>
                        <td style={{ padding: '1rem', fontWeight: 500 }}>{item.tvg_name}</td>
                        <td style={{ padding: '1rem', color: 'var(--text-secondary)' }}>{item.group_title}</td>
                        <td style={{ padding: '1rem' }}>
                          {tmdb ? (
                            <div>
                              <strong style={{ color: 'var(--primary-accent)' }}>{tmdb.tmdb_title}</strong>{' '}
                              <span style={{ color: 'var(--text-secondary)' }}>({tmdb.tmdb_year})</span>
                            </div>
                          ) : (
                            <span style={{ fontStyle: 'italic', color: 'var(--text-secondary)' }}>Non enrichi</span>
                          )}
                        </td>
                        <td style={{ padding: '1rem' }}>
                          <span className={`badge ${item.state === 'downloaded' ? 'badge-success' : item.state === 'downloading' ? 'badge-progress' : item.state === 'failed' ? 'badge-failed' : 'badge-pending'}`}>
                            {item.state}
                          </span>
                        </td>
                        <td style={{ padding: '1rem' }}>{item.created_at}</td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {playlistTotal > 15 && (
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
              <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>Affichage de 15 sur {playlistTotal} entrées</span>
              <div style={{ display: 'flex', gap: '0.25rem' }}>
                <button disabled={playlistPage === 1} onClick={() => setPlaylistPage(p => p - 1)} style={{ padding: '0.4rem 0.8rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', opacity: playlistPage === 1 ? 0.5 : 1 }}>Précédent</button>
                <button disabled={playlistPage * 15 >= playlistTotal} onClick={() => setPlaylistPage(p => p + 1)} style={{ padding: '0.4rem 0.8rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', opacity: playlistPage * 15 >= playlistTotal ? 0.5 : 1 }}>Suivant</button>
              </div>
            </div>
          )}
        </Tabs.Content>

        {/* Tab 2: Processing Logs */}
        <Tabs.Content value="logs" className="card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 600 }}>Dernières Exécutions de Tâches</h2>
            <button onClick={fetchLogs} style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 500, backgroundColor: 'var(--bg-active)', color: 'var(--primary-accent)' }}>Actualiser ↻</button>
          </div>

          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid var(--border-color)', color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                  <th style={{ padding: '0.75rem 1rem' }}>Action</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Éléments Traités</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Statut</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Date de Lancement</th>
                  <th style={{ padding: '0.75rem 1rem' }}>Détails d'Erreur</th>
                </tr>
              </thead>
              <tbody>
                {logsLoading && logs.length === 0 ? (
                  <tr>
                    <td colSpan={5} style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Chargement...</td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td colSpan={5} style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-secondary)' }}>Aucun log enregistré</td>
                  </tr>
                ) : (
                  logs.map(log => (
                    <tr key={log.id} style={{ borderBottom: '1px solid var(--border-color)', fontSize: '0.9rem' }}>
                      <td style={{ padding: '1rem', fontWeight: 600, color: 'var(--primary-slate)' }}>{log.action}</td>
                      <td style={{ padding: '1rem' }}>{log.item_count} items</td>
                      <td style={{ padding: '1rem' }}>
                        <span className={`badge ${log.status === 'success' ? 'badge-success' : log.status === 'failed' ? 'badge-failed' : 'badge-progress'}`}>
                          {log.status === 'in_progress' ? 'En cours' : log.status === 'success' ? 'Réussi' : 'Échoué'}
                        </span>
                      </td>
                      <td style={{ padding: '1rem', color: 'var(--text-secondary)' }}>{new Date(log.started_at).toLocaleString()}</td>
                      <td style={{ padding: '1rem', maxWidth: '300px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {log.error_message ? (
                          <span style={{ color: 'var(--status-failed-text)', fontWeight: 500 }} title={log.error_message}>{log.error_message}</span>
                        ) : (
                          <span style={{ color: 'var(--text-secondary)', fontStyle: 'italic' }}>-</span>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Tabs.Content>

        {/* Tab 3: Downloads Tracker */}
        <Tabs.Content value="downloads" className="card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 600 }}>File de Téléchargement et Organisation</h2>
            <button onClick={fetchDownloads} style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 500, backgroundColor: 'var(--bg-active)', color: 'var(--primary-accent)' }}>Actualiser ↻</button>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {downloadsLoading && downloads.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-secondary)' }}>Chargement en cours...</div>
            ) : downloads.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-secondary)' }}>Aucun téléchargement enregistré</div>
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
                        <h3 style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--primary-slate)', wordBreak: 'break-all' }}>{item.download_path ? filepathBase(item.download_path) : item.url}</h3>
                        <p style={{ color: 'var(--text-secondary)', fontSize: '0.75rem', marginTop: '0.25rem', wordBreak: 'break-all' }}>{item.url}</p>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <span className={`badge ${item.status === 'completed' ? 'badge-success' : item.status === 'downloading' ? 'badge-progress' : item.status === 'failed' ? 'badge-failed' : 'badge-pending'}`}>
                          {item.status}
                        </span>
                        {isCompleted && (
                          <button onClick={() => openMoveDialog(item)} style={{ padding: '0.4rem 0.8rem', fontSize: '0.75rem', fontWeight: 600, borderRadius: 'var(--radius-sm)', border: '1px solid var(--primary-accent)', color: 'var(--primary-accent)', backgroundColor: 'var(--bg-active)', hover: { backgroundColor: '#e0f2fe' } as any }}>
                            Déplacer ⇄
                          </button>
                        )}
                      </div>
                    </div>

                    {/* Progress Bar for Downloading / Retrying */}
                    {!isCompleted && item.status !== 'failed' && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                          <span>Progression: {progress}%</span>
                          {item.file_size && (
                            <span>{(downloaded / 1024 / 1024).toFixed(1)} Mo / {(item.file_size / 1024 / 1024).toFixed(1)} Mo</span>
                          )}
                        </div>
                        {/* Radix UI Progress Bar */}
                        <Progress.Root value={progress} style={{ position: 'relative', overflow: 'hidden', background: 'var(--border-color)', borderRadius: '99999px', width: '100%', height: '10px' }}>
                          <Progress.Indicator style={{ background: 'var(--primary-accent)', width: `${progress}%`, height: '100%', transition: 'width 660ms cubic-bezier(0.65, 0, 0.35, 1)' }} />
                        </Progress.Root>
                      </div>
                    )}

                    {item.error_message && (
                      <div style={{ padding: '0.75rem', backgroundColor: 'var(--status-failed-bg)', borderRadius: 'var(--radius-sm)', color: 'var(--status-failed-text)', fontSize: '0.8rem', fontWeight: 500 }}>
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
          <Dialog.Overlay style={{ backgroundColor: 'rgba(0, 0, 0, 0.4)', position: 'fixed', inset: 0, backdropFilter: 'blur(3px)', zIndex: 100 }} />
          <Dialog.Content style={{ backgroundColor: 'var(--bg-card)', borderRadius: 'var(--radius-lg)', boxShadow: 'var(--shadow-lg)', position: 'fixed', top: '50%', left: '50%', transform: 'translate(-50%, -50%)', width: '90%', maxWidth: '500px', padding: '2rem', zIndex: 101, border: '1px solid var(--border-color)', outline: 'none' }}>
            <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--primary-slate)', marginBottom: '1rem' }}>
              ⇄ Déplacer le dossier de l'œuvre
            </Dialog.Title>
            <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.5, marginBottom: '1.5rem' }}>
              Cette action déplacera l'<strong>intégralité du dossier parent</strong> (films ou séries complètes incluant toutes les saisons) vers un nouveau répertoire de stockage et mettra à jour la base de données.
            </Dialog.Description>

            {moveItem && (
              <div style={{ backgroundColor: 'var(--bg-app)', padding: '1rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)', marginBottom: '1.5rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.85rem' }}>
                <div><strong>📁 Œuvre :</strong> {moveItem.title} <span className="badge badge-pending" style={{ padding: '0.1rem 0.4rem', fontSize: '0.65rem' }}>{moveItem.type}</span></div>
                <div><strong>📌 Chemin actuel :</strong> <code style={{ wordBreak: 'break-all', color: 'var(--text-secondary)' }}>{moveItem.currentPath || 'Inconnu'}</code></div>
              </div>
            )}

            {/* Input Selection */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 600 }}>Dossier de destination parent :</label>
                <select value={targetDir} onChange={e => { setTargetDir(e.target.value); setCustomDirInput(''); }} style={{ width: '100%', padding: '0.6rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', backgroundColor: '#fff', fontSize: '0.85rem' }}>
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
                  <label style={{ fontSize: '0.85rem', fontWeight: 600 }}>Saisir un chemin de destination personnalisé :</label>
                  <input type="text" placeholder="Exemple: /media/enfants-movies" value={customDirInput} onChange={e => setCustomDirInput(e.target.value)} style={{ width: '100%', padding: '0.6rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', backgroundColor: '#fff', fontSize: '0.85rem' }} />
                </div>
              )}
            </div>

            {moveError && (
              <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 500, marginBottom: '1.5rem' }}>
                ⚠️ {moveError}
              </div>
            )}

            {moveSuccess && (
              <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-success-bg)', color: 'var(--status-success-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 500, marginBottom: '1.5rem' }}>
                ✓ {moveSuccess}
              </div>
            )}

            {/* Actions */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
              <button disabled={isMoving} onClick={() => { setIsMoveOpen(false); setMoveItem(null); setCustomDirInput(''); }} style={{ padding: '0.5rem 1rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', color: 'var(--text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                Annuler
              </button>
              <button disabled={isMoving} onClick={handleMoveFolder} style={{ padding: '0.5rem 1.25rem', borderRadius: 'var(--radius-sm)', backgroundColor: 'var(--primary-accent)', color: '#fff', fontSize: '0.85rem', fontWeight: 600, opacity: isMoving ? 0.7 : 1 }}>
                {isMoving ? 'Déplacement...' : '✓ Confirmer le Déplacement'}
              </button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>

    </div>
  );
}
