import { useState, useEffect, FormEvent } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { M3uSource, M3uSourceInput } from '../types';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';

interface M3uSourceDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  source: M3uSource | null; // null = create mode
  existingNames: string[];
  runtimeOnlyNames: string[];
  onSubmit: (name: string, input: M3uSourceInput) => Promise<void>;
}

const emptyInput = (): M3uSourceInput => ({
  file_path: '',
  enabled: true,
  url: '',
  archive_dir: '',
  retention_count: 5,
  max_file_size_mb: 500,
  timeout_seconds: 300,
  retry_attempts: 3,
  auth_username: '',
  // auth_password intentionally omitted: undefined means "unchanged" for an
  // existing source, and "none" for a new one.
});

// Create/edit dialog for a runtime M3U source. In create mode, submitting a
// name that already identifies an existing runtime-only source warns before
// replacing it in place (see "M3U Sources Management").
export function M3uSourceDialog({ isOpen, onOpenChange, source, existingNames, runtimeOnlyNames, onSubmit }: M3uSourceDialogProps) {
  const { t } = useTranslation('settings');
  const translateApiError = useApiErrorMessage();
  const isEdit = source !== null;

  const [name, setName] = useState('');
  const [input, setInput] = useState<M3uSourceInput>(emptyInput());
  const [passwordTouched, setPasswordTouched] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isOpen) return;
    setPasswordTouched(false);
    setError(null);
    if (source) {
      setName(source.name);
      setInput({
        file_path: source.file_path,
        enabled: source.enabled,
        url: source.url,
        archive_dir: source.archive_dir,
        retention_count: source.retention_count,
        max_file_size_mb: source.max_file_size_mb,
        timeout_seconds: source.timeout_seconds,
        retry_attempts: source.retry_attempts,
        auth_username: source.auth_username,
      });
    } else {
      setName('');
      setInput(emptyInput());
    }
  }, [isOpen, source]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError(t('m3uSources.nameRequired'));
      return;
    }

    if (!isEdit && runtimeOnlyNames.includes(name)) {
      if (!confirm(t('m3uSources.confirmReplaceRuntime', { name }))) return;
    }

    setSaving(true);
    setError(null);
    try {
      await onSubmit(name, input);
      onOpenChange(false);
    } catch (err) {
      setError(translateApiError(err));
    } finally {
      setSaving(false);
    }
  };

  const nameCollisionWarning = !isEdit && name && existingNames.includes(name) && !runtimeOnlyNames.includes(name);

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content className="dialog-content">
          <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)', marginBottom: '1rem' }}>
            {isEdit ? t('m3uSources.editTitle') : t('m3uSources.createTitle')}
          </Dialog.Title>
          <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginBottom: '1.5rem' }}>
            {t('m3uSources.description')}
          </Dialog.Description>

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label htmlFor="m3u-source-name" style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('m3uSources.nameLabel')}</label>
              <input
                id="m3u-source-name"
                type="text"
                className="custom-input"
                value={name}
                onChange={e => setName(e.target.value)}
                disabled={isEdit}
                required
              />
              {nameCollisionWarning && (
                <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>
                  {t('m3uSources.nameCollidesWithOrigin')}
                </p>
              )}
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label htmlFor="m3u-source-file-path" style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('m3uSources.filePathLabel')}</label>
              <input
                id="m3u-source-file-path"
                type="text"
                className="custom-input"
                value={input.file_path}
                onChange={e => setInput(prev => ({ ...prev, file_path: e.target.value }))}
                required
              />
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('m3uSources.urlLabel')}</label>
              <input
                type="text"
                className="custom-input"
                value={input.url}
                onChange={e => setInput(prev => ({ ...prev, url: e.target.value }))}
              />
            </div>

            <div style={{ display: 'flex', gap: '1rem' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', flex: 1 }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('m3uSources.authUsernameLabel')}</label>
                <input
                  type="text"
                  className="custom-input"
                  value={input.auth_username}
                  onChange={e => setInput(prev => ({ ...prev, auth_username: e.target.value }))}
                />
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', flex: 1 }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('m3uSources.authPasswordLabel')}</label>
                <input
                  type="password"
                  className="custom-input"
                  value={input.auth_password ?? ''}
                  placeholder={isEdit ? (source?.has_auth_password ? t('sensitive.set') : t('sensitive.notSet')) : undefined}
                  onChange={e => {
                    setPasswordTouched(true);
                    setInput(prev => ({ ...prev, auth_password: e.target.value }));
                  }}
                />
                {isEdit && passwordTouched && (
                  <p className="settings-field-hint">{t('m3uSources.passwordWillChange')}</p>
                )}
              </div>
            </div>

            <div style={{ display: 'flex', gap: '1rem' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', flex: 1 }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('m3uSources.retentionCountLabel')}</label>
                <input
                  type="number"
                  className="custom-input"
                  value={input.retention_count}
                  onChange={e => setInput(prev => ({ ...prev, retention_count: Number(e.target.value) }))}
                />
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', flex: 1 }}>
                <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('m3uSources.retryAttemptsLabel')}</label>
                <input
                  type="number"
                  className="custom-input"
                  value={input.retry_attempts}
                  onChange={e => setInput(prev => ({ ...prev, retry_attempts: Number(e.target.value) }))}
                />
              </div>
            </div>

            <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>
              <input
                type="checkbox"
                checked={input.enabled}
                onChange={e => setInput(prev => ({ ...prev, enabled: e.target.checked }))}
              />
              {t('m3uSources.enabledLabel')}
            </label>

            {error && (
              <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600 }}>
                ⚠️ {error}
              </div>
            )}

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
              <button type="button" disabled={saving} onClick={() => onOpenChange(false)} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
                {t('cancel')}
              </button>
              <button type="submit" disabled={saving} className="btn-primary" style={{ padding: '0.5rem 1.25rem' }}>
                {saving ? t('saving') : t('save')}
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
