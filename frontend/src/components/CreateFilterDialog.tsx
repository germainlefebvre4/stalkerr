import React, { useCallback, useEffect, useRef, useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { api } from '../services/api';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';
import { useM3uSources } from '../hooks/useM3uSources';
import { FilterConfig, FilterOriginEntry, FilterTestTarget } from '../types';
import { DialogReplaceWarning } from './DialogReplaceWarning';
import { FilterTestDrawer } from './FilterTestDrawer';

interface CreateFilterDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: (message: string) => void;
  filters: FilterConfig[];
  filterOrigin: FilterOriginEntry[];
}

export function CreateFilterDialog({ isOpen, onOpenChange, onSuccess, filters, filterOrigin }: CreateFilterDialogProps) {
  const { t } = useTranslation('dialogs');
  const translateApiError = useApiErrorMessage();
  const [newFilterName, setNewFilterName] = useState('');
  const [newFilterAttribute, setNewFilterAttribute] = useState('group_title');
  const [newFilterIncludes, setNewFilterIncludes] = useState('');
  const [newFilterExcludes, setNewFilterExcludes] = useState('');
  const [isFilterCreating, setIsFilterCreating] = useState(false);
  const [filterError, setFilterError] = useState<string | null>(null);

  const { sources, fetchSources } = useM3uSources();
  const [testTarget, setTestTarget] = useState<FilterTestTarget | null>(null);

  // Radix mounts FilterTestDrawer's Dialog.Content in its own portal, outside
  // this dialog's own Content subtree, so dismissing the drawer looks like an
  // "outside interaction" to this dialog's dismissable layer and would
  // otherwise close it too. See RunItemsDialog.tsx for the same pattern.
  const justClosedDrawerRef = useRef(false);

  useEffect(() => {
    if (testTarget !== null) return;
    const id = setTimeout(() => { justClosedDrawerRef.current = false; }, 0);
    return () => clearTimeout(id);
  }, [testTarget]);

  const handleDialogOpenChange = useCallback((open: boolean) => {
    if (!open && (testTarget || justClosedDrawerRef.current)) {
      justClosedDrawerRef.current = false;
      return;
    }
    onOpenChange(open);
  }, [onOpenChange, testTarget]);

  // Reset the form once the dialog finishes closing - adjusted during render
  // (rather than a useEffect+setState pair) per this codebase's convention
  // for "reset local state when a prop transitions" (see SettingsFieldRow).
  const [wasOpen, setWasOpen] = useState(isOpen);
  if (isOpen !== wasOpen) {
    setWasOpen(isOpen);
    if (!isOpen) {
      setNewFilterName('');
      setNewFilterAttribute('group_title');
      setNewFilterIncludes('');
      setNewFilterExcludes('');
      setFilterError(null);
      setTestTarget(null);
    }
  }

  useEffect(() => {
    if (isOpen) {
      fetchSources();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen]);

  // Editing the in-progress values invalidates the current test target/
  // result - adjusted during render (rather than a useEffect+setState pair),
  // same convention as `wasOpen` above - see the "editing patterns resets
  // any previous test result" scenario in frontend-filters-management.
  const testInputsKey = `${newFilterAttribute}|${newFilterIncludes}|${newFilterExcludes}`;
  const [lastTestInputsKey, setLastTestInputsKey] = useState(testInputsKey);
  if (testInputsKey !== lastTestInputsKey) {
    setLastTestInputsKey(testInputsKey);
    setTestTarget(null);
  }

  const handleTestFilter = () => {
    const attributeLabel = newFilterAttribute === 'tvg_name' ? t('filters:tvgNameLabel') : t('filters:groupTitleLabel');
    setTestTarget({
      mode: 'single',
      attribute: newFilterAttribute as 'group_title' | 'tvg_name',
      includePatterns: newFilterIncludes,
      excludePatterns: newFilterExcludes,
      label: t('filters:dryRun.titleInProgress', { attribute: attributeLabel }),
    });
  };

  const existingOverride = filters.find(f => f.attribute === newFilterAttribute);
  const originForAttribute = filterOrigin.find(o => o.attribute === newFilterAttribute);

  const handleLoadCurrentConfig = () => {
    if (existingOverride) {
      setNewFilterIncludes(existingOverride.include_patterns || '');
      setNewFilterExcludes(existingOverride.exclude_patterns || '');
    } else {
      setNewFilterIncludes((originForAttribute?.include_patterns || []).join(', '));
      setNewFilterExcludes((originForAttribute?.exclude_patterns || []).join(', '));
    }
  };

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
      })
      .catch((err: unknown) => {
        setFilterError(translateApiError(err));
      })
      .finally(() => setIsFilterCreating(false));
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={handleDialogOpenChange}>
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
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '0.5rem' }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('createFilter.attributeLabel')}</label>
                <button type="button" onClick={handleLoadCurrentConfig} className="btn-secondary" style={{ padding: '0.25rem 0.6rem', fontSize: '0.75rem' }}>
                  {t('createFilter.loadCurrentConfig')}
                </button>
              </div>
              <select
                value={newFilterAttribute}
                onChange={e => setNewFilterAttribute(e.target.value)}
                className="custom-select"
              >
                <option value="group_title">{t('createFilter.attributeGroupTitle')}</option>
                <option value="tvg_name">{t('createFilter.attributeTvgName')}</option>
              </select>
            </div>

            {existingOverride && (
              <DialogReplaceWarning
                message={t('createFilter.replaceWarning', { name: existingOverride.name })}
              />
            )}

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

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', borderTop: '1px solid var(--border-color)', paddingTop: '0.75rem' }}>
              <button
                type="button"
                onClick={handleTestFilter}
                className="btn-secondary"
                style={{ alignSelf: 'flex-start' }}
              >
                {t('createFilter.testButton')}
              </button>
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

      <FilterTestDrawer
        target={testTarget}
        onOpenChange={(open) => { if (!open) { justClosedDrawerRef.current = true; setTestTarget(null); } }}
        sources={sources}
        modal={false}
        withOverlay={false}
      />
    </Dialog.Root>
  );
}
