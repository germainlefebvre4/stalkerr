import { useState, useEffect } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { api } from '../services/api';
import { PlaylistItem } from '../types';

interface ManualOverrideDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  overrideItemData: PlaylistItem | null;
  onSuccess: (message: string) => void;
}

export function ManualOverrideDialog({ 
  isOpen, 
  onOpenChange, 
  overrideItemData, 
  onSuccess 
}: ManualOverrideDialogProps) {
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

  useEffect(() => {
    if (overrideItemData) {
      // Clean raw tvg_name
      let cleaned = overrideItemData.tvg_name;
      cleaned = cleaned.replace(/^[a-zA-Z]{2,4}\s*[:\-]\s*/, '');
      cleaned = cleaned.replace(/\b(FHD|UHD|4K|1080p|720p|480p|HD|SD|MULTI|VF|VOSTFR|WEB|x264|H264|x265|HEVC|AAC|AC3)\b/ig, '');
      cleaned = cleaned.replace(/\bS\d+E\d+\b/ig, '');
      cleaned = cleaned.replace(/\bS\d+\b/ig, '');
      cleaned = cleaned.replace(/\bE\d+\b/ig, '');
      cleaned = cleaned.replace(/[\(\)\-\[\]]/g, ' ');
      cleaned = cleaned.trim().replace(/\s+/g, ' ');

      setOverrideSearchQuery(cleaned);

      const isTV = overrideItemData.content_type === 'tvshows';
      setOverrideMediaType(isTV ? 'tvshow' : 'movie');

      // Extract Season & Episode
      const seMatch = overrideItemData.tvg_name.match(/S(\d+)E(\d+)/i);
      if (seMatch) {
        setOverrideSeason(parseInt(seMatch[1], 10).toString());
        setOverrideEpisode(parseInt(seMatch[2], 10).toString());
      } else {
        const epMatch = overrideItemData.tvg_name.match(/E(\d+)/i);
        if (epMatch) {
          setOverrideSeason('1');
          setOverrideEpisode(parseInt(epMatch[1], 10).toString());
        } else {
          setOverrideSeason('');
          setOverrideEpisode('');
        }
      }

      const existingYear = isTV ? overrideItemData.tvshow?.tmdb_year : overrideItemData.movie?.tmdb_year;
      setOverrideSearchYear(existingYear ? existingYear.toString() : '');
      setSelectedResult(null);
      setOverrideError(null);
      setOverrideSearchResults([]);

      // Auto-trigger search immediately
      setOverrideSearchLoading(true);
      api.searchTMDB(cleaned, isTV ? 'tvshow' : 'movie', existingYear ? existingYear.toString() : undefined)
        .then(data => {
          setOverrideSearchResults(data);
        })
        .catch((err: any) => {
          setOverrideError(err.message);
        })
        .finally(() => {
          setOverrideSearchLoading(false);
        });
    }
  }, [overrideItemData]);

  const handleSearchTMDB = (queryOverride?: string, typeOverride?: 'movie' | 'tvshow', yearOverride?: string) => {
    const q = queryOverride !== undefined ? queryOverride : overrideSearchQuery;
    const t = typeOverride !== undefined ? typeOverride : overrideMediaType;
    const y = yearOverride !== undefined ? yearOverride : overrideYear;

    if (!q.trim()) return;
    setOverrideSearchLoading(true);
    setOverrideError(null);
    setSelectedResult(null);

    api.searchTMDB(q, t, y)
      .then(data => {
        setOverrideSearchResults(data);
      })
      .catch((err: any) => {
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

    api.forceOverride(overrideItemData.id, payload)
      .then(() => {
        onSuccess('Association TMDB mise à jour avec succès !');
        onOpenChange(false);
      })
      .catch((err: any) => {
        setOverrideError(err.message);
      })
      .finally(() => {
        setIsSubmittingOverride(false);
      });
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
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
                <input 
                  type="number" 
                  placeholder="Ex: 1" 
                  value={overrideSeason} 
                  onChange={e => setOverrideSeason(e.target.value)} 
                  className="custom-input" 
                  style={{ padding: '0.35rem', fontSize: '0.8rem' }} 
                />
              </div>
              <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Épisode (optionnel) :</label>
                <input 
                  type="number" 
                  placeholder="Ex: 5" 
                  value={overrideEpisode} 
                  onChange={e => setOverrideEpisode(e.target.value)} 
                  className="custom-input" 
                  style={{ padding: '0.35rem', fontSize: '0.8rem' }} 
                />
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
            <button 
              disabled={isSubmittingOverride} 
              onClick={() => { onOpenChange(false); }} 
              className="btn-secondary" 
              style={{ padding: '0.5rem 1rem' }}
            >
              Annuler
            </button>
            <button 
              disabled={isSubmittingOverride || !selectedResult} 
              onClick={handleForceOverride} 
              className="btn-primary" 
              style={{ padding: '0.5rem 1.25rem' }}
            >
              {isSubmittingOverride ? 'Association...' : '✓ Forcer l\'association'}
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
