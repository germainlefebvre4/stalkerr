import React from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import * as Dialog from '@radix-ui/react-dialog';
import { PlaylistItem } from '../types';
import { formatDate } from '../utils/date';
import { getPipelineStateBadgeClass } from '../utils/pipelineState';

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
            <label style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Nom du Média (VOD)</label>
            <input 
              type="text" 
              placeholder="Rechercher par titre..." 
              value={playlistSearchName} 
              onChange={e => { setPlaylistSearchName(e.target.value); setPlaylistPage(1); }} 
              className="custom-input" 
              style={{ width: '100%' }}
            />
          </div>

          {/* 2. Recherche Groupe / Catégorie */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
            <label style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Groupe / Catégorie</label>
            <input 
              type="text" 
              placeholder="Rechercher par groupe..." 
              value={playlistSearch} 
              onChange={e => { setPlaylistSearch(e.target.value); setPlaylistPage(1); }} 
              className="custom-input" 
              style={{ width: '100%' }}
            />
          </div>

          {/* 3. Filtrage d'enrichissement TMDB */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
            <label style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Enrichissement TMDB</label>
            <select 
              value={playlistTMDBFilter} 
              onChange={e => { setPlaylistTMDBFilter(e.target.value as 'all' | 'yes' | 'no'); setPlaylistPage(1); }} 
              className="custom-select" 
              style={{ width: '100%', padding: '0.4rem 1.5rem 0.4rem 1rem' }}
            >
              <option value="all">Tous</option>
              <option value="yes">Oui (Enrichis)</option>
              <option value="no">Non (Non Enrichis)</option>
            </select>
          </div>

          {/* 4. État Pipeline */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
            <label style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)' }}>État Pipeline</label>
            <select 
              value={playlistStateFilter} 
              onChange={e => { setPlaylistStateFilter(e.target.value); setPlaylistPage(1); }} 
              className="custom-select" 
              style={{ width: '100%', padding: '0.4rem 1.5rem 0.4rem 1rem' }}
            >
              <option value="all">Tous les Statuts</option>
              <option value="processed">Traité (Processed)</option>
              <option value="pending">En Attente (Pending)</option>
              <option value="downloading">En Cours (Downloading)</option>
              <option value="organizing">En cours d'organisation (Organizing)</option>
              <option value="downloaded">Téléchargé (Downloaded)</option>
              <option value="failed">Échoué (Failed)</option>
            </select>
          </div>
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
                        <span style={{ fontStyle: 'italic', color: 'var(--text-muted)' }}>Non enrichi</span>
                      )}
                    </td>
                    <td>
                      <span className={`badge ${getPipelineStateBadgeClass(item.state)}`}>
                        {item.state}
                      </span>
                    </td>
                    <td style={{ color: 'var(--text-secondary)', fontSize: '0.8rem' }}>{formatDate(item.created_at)}</td>
                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'flex', gap: '0.35rem', justifyContent: 'flex-end' }}>
                        <button
                          onClick={(e) => { e.stopPropagation(); onOpenOverride(item); }}
                          className="btn-primary"
                          style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}
                          title="Corriger ou forcer manuellement l'association TMDB"
                        >
                          {item.override_by ? 'Corriger ✏️' : 'Associer 🔍'}
                        </button>
                        <button 
                          onClick={(e) => { e.stopPropagation(); onResetPipeline(item.id, item.content_type); }} 
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

      {/* Pagination & Limit Selector */}
      {playlistTotal > 0 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
              {`Affichage de ${playlistTotal === 0 ? 0 : (playlistPage - 1) * playlistLimit + 1} à ${Math.min(playlistPage * playlistLimit, playlistTotal)} sur ${playlistTotal} entrées`}
            </span>
            
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Afficher :</span>
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
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>par page</span>
            </div>
          </div>

          {totalPages > 1 && (
            <div style={{ display: 'flex', gap: '0.35rem', alignItems: 'center' }}>
              <button 
                disabled={playlistPage === 1} 
                onClick={() => setPlaylistPage(1)} 
                className="btn-secondary" 
                style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: playlistPage === 1 ? 0.5 : 1 }}
                title="Première page"
              >
                &lt;&lt;
              </button>
              <button 
                disabled={playlistPage === 1} 
                onClick={() => setPlaylistPage(p => Math.max(1, p - 1))} 
                className="btn-secondary" 
                style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: playlistPage === 1 ? 0.5 : 1 }}
                title="Page précédente"
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
                title="Page suivante"
              >
                &gt;
              </button>
              <button
                disabled={playlistPage === totalPages}
                onClick={() => setPlaylistPage(totalPages)}
                className="btn-secondary"
                style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: playlistPage === totalPages ? 0.5 : 1 }}
                title="Dernière page"
              >
                &gt;&gt;
              </button>

              <div style={{ display: 'flex', alignItems: 'center', gap: '0.3rem', marginLeft: '0.4rem' }}>
                <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Aller à :</span>
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
                  Go
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
                📋 Détails de la Piste
              </Dialog.Title>
              <Dialog.Close className="btn-secondary" style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem', borderRadius: 'var(--radius-sm)' }}>
                Fermer ✕
              </Dialog.Close>
            </div>

            <Dialog.Description style={{ display: 'none' }}>
              Détails techniques et métadonnées d'enrichissement de la piste de playlist.
            </Dialog.Description>

            {selectedItem && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem', flex: 1, paddingBottom: '1rem' }}>
                
                {/* Section 1: Enrichissement Métadonnées (TMDB) */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <h3 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.35rem', margin: 0 }}>
                    🎬 Enrichissement Métadonnées (TMDB)
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
                              <strong>Genres :</strong> {tmdb.genres}
                            </div>
                          )}
                          {isMovie && selectedItem.movie?.duration && (
                            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                              <strong>Durée :</strong> {selectedItem.movie.duration} minutes
                            </div>
                          )}
                          {!isMovie && (selectedItem.tvshow?.season !== undefined || (selectedItem as any).season !== undefined) && (
                            <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                              <strong>Position :</strong> Saison {selectedItem.tvshow?.season ?? (selectedItem as any).season}, Épisode {selectedItem.tvshow?.episode ?? (selectedItem as any).episode}
                            </div>
                          )}
                          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontWeight: 500, marginTop: '0.25rem' }}>
                            <strong>TMDB ID :</strong> {tmdb.tmdb_id}
                          </div>
                        </div>
                      );
                    } else {
                      return (
                        <div style={{ backgroundColor: 'var(--bg-app)', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)', padding: '1rem', textAlign: 'center', color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem' }}>
                          Aucune association TMDB officielle. Cet item est classifié comme : <strong>{selectedItem.content_type}</strong>.
                        </div>
                      );
                    }
                  })()}
                </div>

                {/* Section 2: État du Pipeline d'Ingestion */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <h3 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.35rem', margin: 0 }}>
                    ⚙️ État du Pipeline
                  </h3>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem', fontSize: '0.85rem' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>Statut actuel :</span>
                      <div>
                        <span className={`badge ${getPipelineStateBadgeClass(selectedItem.state)}`} style={{ fontSize: '0.75rem' }}>
                          {selectedItem.state}
                        </span>
                      </div>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>Type de contenu :</span>
                      <span style={{ fontWeight: 600, color: 'var(--primary-slate)', textTransform: 'capitalize' }}>{selectedItem.content_type}</span>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>Date d'import :</span>
                      <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{formatDate(selectedItem.created_at)}</span>
                    </div>
                    {selectedItem.override_by && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                        <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>Forcé par :</span>
                        <span style={{ fontWeight: 600, color: 'var(--primary-accent)' }}>{selectedItem.override_by}</span>
                      </div>
                    )}
                    {selectedItem.override_at && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                        <span style={{ color: 'var(--text-muted)', fontWeight: 500 }}>Forcé le :</span>
                        <span style={{ fontWeight: 500, color: 'var(--text-secondary)' }}>{formatDate(selectedItem.override_at)}</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Section 3: Informations de Provenance M3U */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  <h3 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.35rem', margin: 0 }}>
                    📁 Provenance Playlist M3U
                  </h3>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem', fontSize: '0.85rem' }}>
                    <div>
                      <strong style={{ color: 'var(--text-secondary)' }}>🏷️ Nom d'origine :</strong>{' '}
                      <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>{selectedItem.tvg_name}</span>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--text-secondary)' }}>📁 Catégorie d'origine :</strong>{' '}
                      <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>{selectedItem.group_title}</span>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--text-secondary)' }}>🔢 Numéro de ligne M3U :</strong>{' '}
                      <span style={{ color: 'var(--primary-slate)', fontWeight: 600 }}>
                        {selectedItem.line_number > 0 ? (
                          <span style={{ backgroundColor: 'rgba(99, 102, 241, 0.08)', padding: '0.15rem 0.4rem', borderRadius: '4px', border: '1px solid rgba(99, 102, 241, 0.15)' }}>
                            Ligne {selectedItem.line_number}
                          </span>
                        ) : (
                          <span style={{ fontStyle: 'italic', color: 'var(--text-muted)' }}>Inconnue (Ancien import)</span>
                        )}
                      </span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                      <strong style={{ color: 'var(--text-secondary)' }}>🔑 Hash unique :</strong>{' '}
                      <code style={{ fontFamily: 'monospace', fontSize: '0.75rem', backgroundColor: 'var(--bg-app)', padding: '0.15rem 0.35rem', borderRadius: '4px', border: '1px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                        {selectedItem.line_hash}
                      </code>
                      <button
                        type="button"
                        onClick={() => handleCopy(selectedItem.line_hash, 'hash')}
                        style={{ fontSize: '0.75rem', color: 'var(--primary-accent)', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '0.15rem', cursor: 'pointer', border: 'none', background: 'none', padding: 0 }}
                      >
                        {copiedText === 'hash' ? '✓ Copié' : '📄 Copier'}
                      </button>
                    </div>
                  </div>
                </div>

                {/* Section 4: Ligne M3U d'origine complète (Copiable) */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  <h3 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.35rem', margin: 0 }}>
                    📄 Contenu Brut de la Ligne
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
                      {copiedText === 'content' ? '✓ Copié !' : '📋 Copier'}
                    </button>
                  </div>
                </div>

                {/* Section 5: URL Brute de streaming (Copiable) */}
                {selectedItem.line_url && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    <h3 style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.35rem', margin: 0 }}>
                      🔗 URL brute du flux
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
                        {copiedText === 'url' ? '✓ Copié !' : '📋 Copier'}
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
