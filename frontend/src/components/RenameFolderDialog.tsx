import { useState, useEffect } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { api } from '../services/api';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';

interface RenameFolderDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  renameItem: { id: number; title: string; folderName: string } | null;
  onSuccess: (id: number, newPath: string, message: string) => void;
}

export function RenameFolderDialog({
  isOpen,
  onOpenChange,
  renameItem,
  onSuccess
}: RenameFolderDialogProps) {
  const { t } = useTranslation('dialogs');
  const translateApiError = useApiErrorMessage();
  const [newName, setNewName] = useState('');
  const [destinationParentDir, setDestinationParentDir] = useState('');
  const [isRenaming, setIsRenaming] = useState(false);
  const [renameError, setRenameError] = useState<string | null>(null);
  const [renameSuccess, setRenameSuccess] = useState<string | null>(null);

  useEffect(() => {
    if (!renameItem) return;
    void Promise.resolve().then(() => {
      setNewName(renameItem.folderName);
      setDestinationParentDir('');
      setRenameError(null);
      setRenameSuccess(null);
    });
  }, [renameItem]);

  const handleRenameFolder = () => {
    if (!renameItem) return;
    if (!newName.trim()) {
      setRenameError(t('renameFolder.nameRequired'));
      return;
    }

    setIsRenaming(true);
    setRenameError(null);
    setRenameSuccess(null);

    api.renameDownload(renameItem.id, {
      new_name: newName.trim(),
      ...(destinationParentDir ? { destination_parent_dir: destinationParentDir } : {}),
    })
      .then((res) => {
        const successMessage = t('renameFolder.successMessage');
        setRenameSuccess(successMessage);
        onSuccess(renameItem.id, res.new_path, successMessage);
        setTimeout(() => {
          onOpenChange(false);
          setRenameSuccess(null);
        }, 2000);
      })
      .catch((err: unknown) => {
        setRenameError(translateApiError(err));
      })
      .finally(() => setIsRenaming(false));
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content className="dialog-content">
          <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--primary-slate)', marginBottom: '1rem' }}>
            {t('renameFolder.title')}
          </Dialog.Title>
          <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.5, marginBottom: '1.5rem', fontWeight: 500 }}>
            {t('renameFolder.description')}
          </Dialog.Description>

          {renameItem && (
            <div style={{ backgroundColor: 'var(--bg-app)', padding: '1rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', marginBottom: '1.5rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.85rem' }}>
              <div><strong>{t('renameFolder.workLabel')}</strong> {renameItem.title}</div>
            </div>
          )}

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginBottom: '1.5rem' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('renameFolder.nameLabel')}</label>
              <input
                type="text"
                value={newName}
                onChange={e => setNewName(e.target.value)}
                className="custom-input"
              />
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('renameFolder.destinationLabel')}</label>
              <input
                type="text"
                placeholder={t('renameFolder.destinationPlaceholder')}
                value={destinationParentDir}
                onChange={e => setDestinationParentDir(e.target.value)}
                className="custom-input"
              />
            </div>
          </div>

          {renameError && (
            <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600, marginBottom: '1.5rem', border: '1px solid var(--status-failed-border)' }}>
              ⚠️ {renameError}
            </div>
          )}

          {renameSuccess && (
            <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-success-bg)', color: 'var(--status-success-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600, marginBottom: '1.5rem', border: '1px solid var(--status-success-border)' }}>
              ✓ {renameSuccess}
            </div>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
            <button
              disabled={isRenaming}
              onClick={() => onOpenChange(false)}
              className="btn-secondary"
              style={{ padding: '0.5rem 1rem' }}
            >
              {t('cancel')}
            </button>
            <button
              disabled={isRenaming}
              onClick={handleRenameFolder}
              className="btn-primary"
              style={{ padding: '0.5rem 1.25rem' }}
            >
              {isRenaming ? t('renameFolder.renaming') : t('renameFolder.confirm')}
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
