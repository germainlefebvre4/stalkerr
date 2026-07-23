import React, { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { api } from '../services/api';

interface CreateFilterDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: (message: string) => void;
}

export function CreateFilterDialog({ isOpen, onOpenChange, onSuccess }: CreateFilterDialogProps) {
  const [newFilterName, setNewFilterName] = useState('');
  const [newFilterAttribute, setNewFilterAttribute] = useState('group_title');
  const [newFilterIncludes, setNewFilterIncludes] = useState('');
  const [newFilterExcludes, setNewFilterExcludes] = useState('');
  const [isFilterCreating, setIsFilterCreating] = useState(false);
  const [filterError, setFilterError] = useState<string | null>(null);

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

    api.createFilter(payload)
      .then(() => {
        onSuccess('Nouveau filtre de tri configuré avec succès !');
        onOpenChange(false);
        setNewFilterName('');
        setNewFilterIncludes('');
        setNewFilterExcludes('');
      })
      .catch((err: any) => {
        setFilterError(err.message);
      })
      .finally(() => setIsFilterCreating(false));
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
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
              <input 
                type="text" 
                placeholder="Ex: Filtre Films Francophones" 
                value={newFilterName} 
                onChange={e => setNewFilterName(e.target.value)} 
                className="custom-input" 
                required 
              />
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Attribut ciblé du média :</label>
              <select 
                value={newFilterAttribute} 
                onChange={e => setNewFilterAttribute(e.target.value)} 
                className="custom-select"
              >
                <option value="group_title">Group Title (Ex: VOD-FR, SERIES-US)</option>
                <option value="tvg_name">TVG Name (Ex: Inception.2010.mkv)</option>
              </select>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Patterns d'Inclusion (Séparés par des virgules) :</label>
              <input 
                type="text" 
                placeholder="Ex: FRENCH, TRUEFRENCH, VFF" 
                value={newFilterIncludes} 
                onChange={e => setNewFilterIncludes(e.target.value)} 
                className="custom-input" 
              />
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>Patterns d'Exclusion (Séparés par des virgules) :</label>
              <input 
                type="text" 
                placeholder="Ex: VOSTFR, SUBBED, HDLight" 
                value={newFilterExcludes} 
                onChange={e => setNewFilterExcludes(e.target.value)} 
                className="custom-input" 
              />
            </div>

            {filterError && (
              <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600, border: '1px solid var(--status-failed-border)' }}>
                ⚠️ {filterError}
              </div>
            )}

            {/* Actions */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
              <button 
                type="button" 
                disabled={isFilterCreating} 
                onClick={() => { onOpenChange(false); setFilterError(null); }} 
                className="btn-secondary" 
                style={{ padding: '0.5rem 1rem' }}
              >
                Annuler
              </button>
              <button 
                type="submit" 
                disabled={isFilterCreating} 
                className="btn-primary" 
                style={{ padding: '0.5rem 1.25rem' }}
              >
                {isFilterCreating ? 'Enregistrement...' : '✓ Enregistrer le Filtre'}
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
