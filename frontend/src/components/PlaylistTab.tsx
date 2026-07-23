import React from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import { PlaylistItem } from '../types';

interface PlaylistTabProps {
  playlist: PlaylistItem[];
  playlistSearch: string;
  setPlaylistSearch: (search: string) => void;
  playlistFilter: 'all' | 'movies' | 'tvshows';
  setPlaylistFilter: (filter: 'all' | 'movies' | 'tvshows') => void;
  playlistStateFilter: string;
  setPlaylistStateFilter: (state: string) => void;
  playlistTotal: number;
  playlistPage: number;
  setPlaylistPage: React.Dispatch<React.SetStateAction<number>>;
  playlistLoading: boolean;
  onOpenOverride: (item: PlaylistItem) => void;
  onResetPipeline: (id: number, contentType: string) => void;
}

export function PlaylistTab({
  playlist,
  playlistSearch,
  setPlaylistSearch,
  playlistFilter,
  setPlaylistFilter,
  playlistStateFilter,
  setPlaylistStateFilter,
  playlistTotal,
  playlistPage,
  setPlaylistPage,
  playlistLoading,
  onOpenOverride,
  onResetPipeline
}: PlaylistTabProps) {
  return (
    <Tabs.Content value="playlist" className="card" style={{ padding: '2rem' }}>
      <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
        
        {/* Filter Content Type Buttons */}
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
          <button 
            onClick={() => { setPlaylistFilter('all'); setPlaylistPage(1); }} 
            className={playlistFilter === 'all' ? 'btn-primary' : 'btn-secondary'} 
            style={{ padding: '0.45rem 1rem' }}
          >
            Tous
          </button>
          <button 
            onClick={() => { setPlaylistFilter('movies'); setPlaylistPage(1); }} 
            className={playlistFilter === 'movies' ? 'btn-primary' : 'btn-secondary'} 
            style={{ padding: '0.45rem 1rem' }}
          >
            Films
          </button>
          <button 
            onClick={() => { setPlaylistFilter('tvshows'); setPlaylistPage(1); }} 
            className={playlistFilter === 'tvshows' ? 'btn-primary' : 'btn-secondary'} 
            style={{ padding: '0.45rem 1rem' }}
          >
            Séries
          </button>
          
          {/* Vertical Divider */}
          <div style={{ width: '1px', background: 'var(--border-color)', margin: '0 0.5rem' }} />
          
          {/* Filter State Selection */}
          <select 
            value={playlistStateFilter} 
            onChange={e => { setPlaylistStateFilter(e.target.value); setPlaylistPage(1); }} 
            className="custom-select" 
            style={{ width: 'auto', padding: '0.4rem 1.5rem 0.4rem 1rem', height: '100%' }}
          >
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
          <input 
            type="text" 
            placeholder="Rechercher par Groupe / VOD..." 
            value={playlistSearch} 
            onChange={e => { setPlaylistSearch(e.target.value); setPlaylistPage(1); }} 
            className="custom-input" 
          />
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
                          onClick={() => onOpenOverride(item)}
                          className="btn-primary"
                          style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}
                          title="Corriger ou forcer manuellement l'association TMDB"
                        >
                          {item.override_by ? 'Corriger ✏️' : 'Associer 🔍'}
                        </button>
                        <button 
                          onClick={() => onResetPipeline(item.id, item.content_type)} 
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
            <button 
              disabled={playlistPage === 1} 
              onClick={() => setPlaylistPage(p => p - 1)} 
              className="btn-secondary" 
              style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem', opacity: playlistPage === 1 ? 0.5 : 1 }}
            >
              Précédent
            </button>
            <button 
              disabled={playlistPage * 15 >= playlistTotal} 
              onClick={() => setPlaylistPage(p => p + 1)} 
              className="btn-secondary" 
              style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem', opacity: playlistPage * 15 >= playlistTotal ? 0.5 : 1 }}
            >
              Suivant
            </button>
          </div>
        </div>
      )}
    </Tabs.Content>
  );
}
