import React, { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { api } from '../services/api';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';

interface CreateFilterDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: (message: string) => void;
}

export function CreateFilterDialog({ isOpen, onOpenChange, onSuccess }: CreateFilterDialogProps) {
  const { t } = useTranslation('dialogs');
  const translateApiError = useApiErrorMessage();
  const [newFilterName, setNewFilterName] = useState('');
  const [newFilterAttribute, setNewFilterAttribute] = useState('group_title');
  const [newFilterIncludes, setNewFilterIncludes] = useState('');
  const [newFilterExcludes, setNewFilterExcludes] = useState('');
  const [isFilterCreating, setIsFilterCreating] = useState(false);
  const [filterError, setFilterError] = useState<string | null>(null);

  const handleCreateFilter = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newFilterName.trim()) {
      setFilterError(t('createFilter.nameRequired'));
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
        onSuccess(t('createFilter.successMessage'));
        onOpenChange(false);
        setNewFilterName('');
        setNewFilterIncludes('');
        setNewFilterExcludes('');
      })
      .catch((err: unknown) => {
        setFilterError(translateApiError(err));
      })
      .finally(() => setIsFilterCreating(false));
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content className="dialog-content">
          <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--primary-slate)', marginBottom: '1rem' }}>
            {t('createFilter.title')}
          </Dialog.Title>
          <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.5, marginBottom: '1.5rem', fontWeight: 500 }}>
            {t('createFilter.description')}
          </Dialog.Description>

          <form onSubmit={handleCreateFilter} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('createFilter.nameLabel')}</label>
              <input
                type="text"
                placeholder={t('createFilter.namePlaceholder')}
                value={newFilterName}
                onChange={e => setNewFilterName(e.target.value)}
                className="custom-input"
                required
              />
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('createFilter.attributeLabel')}</label>
              <select
                value={newFilterAttribute}
                onChange={e => setNewFilterAttribute(e.target.value)}
                className="custom-select"
              >
                <option value="group_title">{t('createFilter.attributeGroupTitle')}</option>
                <option value="tvg_name">{t('createFilter.attributeTvgName')}</option>
              </select>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('createFilter.includeLabel')}</label>
              <input
                type="text"
                placeholder={t('createFilter.includePlaceholder')}
                value={newFilterIncludes}
                onChange={e => setNewFilterIncludes(e.target.value)}
                className="custom-input"
              />
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('createFilter.excludeLabel')}</label>
              <input
                type="text"
                placeholder={t('createFilter.excludePlaceholder')}
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
                {t('cancel')}
              </button>
              <button
                type="submit"
                disabled={isFilterCreating}
                className="btn-primary"
                style={{ padding: '0.5rem 1.25rem' }}
              >
                {isFilterCreating ? t('createFilter.saving') : t('createFilter.save')}
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
