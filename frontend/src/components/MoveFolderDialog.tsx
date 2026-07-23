import { useState, useEffect } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { api } from '../services/api';
import { ConfigPaths } from '../types';

interface MoveFolderDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  moveItem: { id: number; title: string; type: 'movie' | 'tvshow'; currentPath?: string } | null;
  configPaths: ConfigPaths | null;
  onSuccess: (message: string) => void;
}

export function MoveFolderDialog({ 
  isOpen, 
  onOpenChange, 
  moveItem, 
  configPaths, 
  onSuccess 
}: MoveFolderDialogProps) {
  const [targetDir, setTargetDir] = useState('');
  const [customDirInput, setCustomDirInput] = useState('');
  const [isMoving, setIsMoving] = useState(false);
  const [moveError, setMoveError] = useState<string | null>(null);
  const [moveSuccess, setMoveSuccess] = useState<string | null>(null);

  useEffect(() => {
    if (moveItem) {
      setTargetDir(moveItem.type === 'movie' ? configPaths?.movies_path || '' : configPaths?.tvshows_path || '');
      setCustomDirInput('');
      setMoveError(null);
      setMoveSuccess(null);
    }
  }, [moveItem, configPaths]);

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

    api.moveFolder(moveItem.id, moveItem.type, destDir)
      .then(() => {
        setMoveSuccess('Le dossier parent a été déplacé avec succès !');
        onSuccess('Le dossier parent a été déplacé avec succès !');
        setTimeout(() => {
          onOpenChange(false);
          setCustomDirInput('');
          setMoveSuccess(null);
        }, 2000);
      })
      .catch((err: any) => {
        setMoveError(err.message);
      })
      .finally(() => setIsMoving(false));
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
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
              <select 
                value={targetDir} 
                onChange={e => { setTargetDir(e.target.value); setCustomDirInput(''); }} 
                className="custom-select"
              >
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
                <input 
                  type="text" 
                  placeholder="Exemple: /media/enfants-movies" 
                  value={customDirInput} 
                  onChange={e => setCustomDirInput(e.target.value)} 
                  className="custom-input" 
                />
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
            <button 
              disabled={isMoving} 
              onClick={() => { onOpenChange(false); setCustomDirInput(''); }} 
              className="btn-secondary" 
              style={{ padding: '0.5rem 1rem' }}
            >
              Annuler
            </button>
            <button 
              disabled={isMoving} 
              onClick={handleMoveFolder} 
              className="btn-primary" 
              style={{ padding: '0.5rem 1.25rem' }}
            >
              {isMoving ? 'Déplacement...' : '✓ Confirmer'}
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
