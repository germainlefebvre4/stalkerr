import { useState, useEffect } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { api } from '../services/api';
import { ConfigPaths } from '../types';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';

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
  const { t } = useTranslation('dialogs');
  const translateApiError = useApiErrorMessage();
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
      setMoveError(t('moveFolder.destinationRequired'));
      setIsMoving(false);
      return;
    }

    api.moveFolder(moveItem.id, moveItem.type, destDir)
      .then(() => {
        const successMessage = t('moveFolder.successMessage');
        setMoveSuccess(successMessage);
        onSuccess(successMessage);
        setTimeout(() => {
          onOpenChange(false);
          setCustomDirInput('');
          setMoveSuccess(null);
        }, 2000);
      })
      .catch((err: unknown) => {
        setMoveError(translateApiError(err));
      })
      .finally(() => setIsMoving(false));
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content className="dialog-content">
          <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--primary-slate)', marginBottom: '1rem' }}>
            {t('moveFolder.title')}
          </Dialog.Title>
          <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.5, marginBottom: '1.5rem', fontWeight: 500 }}>
            {t('moveFolder.description')}
          </Dialog.Description>

          {moveItem && (
            <div style={{ backgroundColor: 'var(--bg-app)', padding: '1rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', marginBottom: '1.5rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.85rem' }}>
              <div><strong>{t('moveFolder.workLabel')}</strong> {moveItem.title} <span className="badge badge-pending" style={{ padding: '0.1rem 0.4rem', fontSize: '0.65rem' }}>{moveItem.type}</span></div>
              <div><strong>{t('moveFolder.currentPathLabel')}</strong> <code style={{ wordBreak: 'break-all', color: 'var(--text-secondary)' }}>{moveItem.currentPath || t('moveFolder.unknownPath')}</code></div>
            </div>
          )}

          {/* Input Selection */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginBottom: '1.5rem' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('moveFolder.destinationLabel')}</label>
              <select
                value={targetDir}
                onChange={e => { setTargetDir(e.target.value); setCustomDirInput(''); }}
                className="custom-select"
              >
                <option value="">{t('moveFolder.selectRootPlaceholder')}</option>
                {configPaths && (
                  <>
                    <option value={configPaths.movies_path}>{t('moveFolder.moviesFolderOption', { path: configPaths.movies_path })}</option>
                    <option value={configPaths.tvshows_path}>{t('moveFolder.tvshowsFolderOption', { path: configPaths.tvshows_path })}</option>
                  </>
                )}
                <option value="custom">{t('moveFolder.customFolderOption')}</option>
              </select>
            </div>

            {(targetDir === 'custom' || !configPaths) && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('moveFolder.customPathLabel')}</label>
                <input
                  type="text"
                  placeholder={t('moveFolder.customPathPlaceholder')}
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
              {t('cancel')}
            </button>
            <button
              disabled={isMoving}
              onClick={handleMoveFolder}
              className="btn-primary"
              style={{ padding: '0.5rem 1.25rem' }}
            >
              {isMoving ? t('moveFolder.moving') : t('moveFolder.confirm')}
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
