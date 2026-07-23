import * as Tabs from '@radix-ui/react-tabs';
import * as Progress from '@radix-ui/react-progress';
import { DownloadEnriched } from '../types';

interface DownloadsTabProps {
  downloads: DownloadEnriched[];
  downloadsLoading: boolean;
  statusFilter: string;
  setStatusFilter: (status: string) => void;
  typeFilter: string;
  setTypeFilter: (type: string) => void;
  problemFilter: string;
  setProblemFilter: (problem: string) => void;
  onFetchDownloads: () => void;
  onOpenMoveDialog: (item: DownloadEnriched) => void;
}

export function DownloadsTab({
  downloads,
  downloadsLoading,
  statusFilter,
  setStatusFilter,
  typeFilter,
  setTypeFilter,
  problemFilter,
  setProblemFilter,
  onFetchDownloads,
  onOpenMoveDialog
}: DownloadsTabProps) {
  
  const filepathBase = (path: string) => {
    const parts = path.split('/');
    return parts[parts.length - 1];
  };

  return (
    <Tabs.Content value="downloads" className="card" style={{ padding: '2rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>File de Téléchargement et Organisation</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>Suivi dynamique des téléchargements de fichiers vidéo et réorganisations</p>
        </div>
        <button onClick={onFetchDownloads} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>Actualiser ↻</button>
      </div>

      {/* Filter Bar */}
      <div className="filters" style={{ display: 'flex', gap: '0.75rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
        <select 
          value={statusFilter} 
          onChange={e => setStatusFilter(e.target.value)}
          style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: '#fff', fontWeight: 600, color: 'var(--text-secondary)' }}
        >
          <option value="">Statut : Tous</option>
          <option value="completed">Complétés</option>
          <option value="downloading">En cours</option>
          <option value="failed">Échoués</option>
        </select>
        
        <select 
          value={typeFilter} 
          onChange={e => setTypeFilter(e.target.value)}
          style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: '#fff', fontWeight: 600, color: 'var(--text-secondary)' }}
        >
          <option value="">Type : Tous</option>
          <option value="movies">Films</option>
          <option value="tvshows">Séries</option>
        </select>
        
        <select 
          value={problemFilter} 
          onChange={e => setProblemFilter(e.target.value)}
          style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: '#fff', fontWeight: 600, color: 'var(--text-secondary)' }}
        >
          <option value="">Problèmes : Aucun</option>
          <option value="missing_year">Année manquante</option>
          <option value="year_mismatch">Année incorrecte</option>
          <option value="unknown_format">Format inconnu</option>
          <option value="low_quality">Basse qualité</option>
        </select>
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

            // Metadata extraction
            const title = item.content?.title || (item.download_path ? filepathBase(item.download_path) : item.url);
            const year = item.content?.year ? `(${item.content.year})` : '';
            const typeIcon = item.content?.type === 'movies' ? '🎬' : item.content?.type === 'tvshows' ? '📺' : '🔗';
            
            // Status mapping
            let statusLabel: string = item.status;
            let statusBadgeClass = 'badge-pending';
            if (item.status === 'completed') {
              statusLabel = '✅ Complété';
              statusBadgeClass = 'badge-success';
            } else if (item.status === 'downloading') {
              statusLabel = '⏳ En cours';
              statusBadgeClass = 'badge-progress';
            } else if (item.status === 'failed') {
              statusLabel = item.retry_count > 0 ? `❌ Échec (${item.retry_count}×)` : '❌ Échec';
              statusBadgeClass = 'badge-failed';
            } else if (item.status === 'pending') {
              statusLabel = '⏳ En attente';
              statusBadgeClass = 'badge-pending';
            } else if (item.status === 'retrying') {
              statusLabel = '🔄 Réessai';
              statusBadgeClass = 'badge-pending';
            }

            // Problem detection flags
            const hasYearIssue = item.file_info && !item.file_info.has_year_in_path;
            const hasYearMismatch = item.file_info?.year_mismatch;
            const hasFormatIssue = item.file_info && !item.file_info.is_valid_format;
            const resolutionLower = item.file_info?.detected_resolution?.toLowerCase() || '';
            const isLowQuality = resolutionLower === '480p' || resolutionLower === '360p';

            return (
              <div key={item.id} className="download-card" style={{ border: '1px solid var(--border-color)', borderRadius: 'var(--radius-md)', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', backgroundColor: '#fff', boxShadow: 'var(--shadow-sm)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '0.5rem' }}>
                  <div style={{ flex: 1, minWidth: '250px' }}>
                    <h3 style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--primary-slate)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <span>{typeIcon}</span>
                      <span>{title} {year}</span>
                    </h3>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '0.75rem', marginTop: '0.25rem', wordBreak: 'break-all', fontWeight: 500 }}>{item.url}</p>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <span className={`badge ${statusBadgeClass}`} style={{ fontSize: '0.8rem', padding: '0.25rem 0.6rem', fontWeight: 600 }}>
                      {statusLabel}
                    </span>
                    {isCompleted && (
                      <button onClick={() => onOpenMoveDialog(item)} className="btn-primary" style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}>
                        Déplacer ⇄
                      </button>
                    )}
                  </div>
                </div>

                {/* File Path Section */}
                {item.file_info && (
                  <div className="file-info" style={{ backgroundColor: 'var(--bg-app)', padding: '0.5rem 0.75rem', borderRadius: 'var(--radius-sm)', fontFamily: "'Courier New', monospace", fontSize: '0.85rem', lineHeight: '1.4', border: '1px solid var(--border-color)' }}>
                    <div style={{ fontWeight: 600, color: 'var(--text-secondary)' }}>📂 Dossier : {item.file_info.folder_name}/</div>
                    {item.file_info.file_name && (
                      <div style={{ color: 'var(--text-muted)', paddingLeft: '1rem' }}>└─ {item.file_info.file_name}</div>
                    )}
                  </div>
                )}

                {/* Technical specs & Validation row */}
                {item.file_info && (
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem', fontSize: '0.8rem' }}>
                    {/* Technical values inline row */}
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', alignItems: 'center', color: 'var(--primary-slate)', fontWeight: 500 }}>
                      <span>📹 Format : {item.file_info.extension.toUpperCase() || 'Inconnu'}</span>
                      <span>•</span>
                      <span>{item.file_info.detected_resolution || 'Résolution inconnue'}</span>
                      <span>•</span>
                      {item.file_size && (
                        <span>{(item.file_size / 1024 / 1024).toFixed(1)} Mo</span>
                      )}
                      {item.content?.duration && (
                        <>
                          <span>•</span>
                          <span>🕒 {item.content.duration} min</span>
                        </>
                      )}
                    </div>

                    {/* Validation chips row */}
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.35rem', alignItems: 'center' }}>
                      {/* Low quality check */}
                      {isLowQuality && (
                        <span className="badge badge-pending" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          ⚠️ Basse qualité ({item.file_info.detected_resolution})
                        </span>
                      )}
                      
                      {/* Year validity */}
                      {hasYearIssue ? (
                        <span className="badge badge-failed" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          ⚠️ Année manquante
                        </span>
                      ) : hasYearMismatch ? (
                        <span className="badge badge-failed" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }} title="Année dans le path différente de TMDB">
                          ⚠️ Année incorrecte
                        </span>
                      ) : (
                        <span className="badge badge-success" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          ✅ Année OK
                        </span>
                      )}

                      {/* Format validity */}
                      {hasFormatIssue ? (
                        <span className="badge badge-failed" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          ⚠️ Format inconnu
                        </span>
                      ) : (
                        <span className="badge badge-success" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                          ✅ Format OK
                        </span>
                      )}
                    </div>
                  </div>
                )}

                {/* Genres Section */}
                {item.content?.genres && (
                  <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '0.35rem', fontWeight: 500 }}>
                    <span>🎭 Genres :</span>
                    <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>{item.content.genres}</span>
                  </div>
                )}

                {/* Progress Bar for Downloading / Retrying */}
                {!isCompleted && item.status !== 'failed' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>
                      <span>Progression : {progress}%</span>
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
                    🔴 {item.error_message}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </Tabs.Content>
  );
}
