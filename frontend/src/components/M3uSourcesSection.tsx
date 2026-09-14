import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { M3uSource, M3uSourceInput } from '../types';
import { M3uSourceDialog } from './M3uSourceDialog';

interface M3uSourcesSectionProps {
  isExpanded: boolean;
  sources: M3uSource[];
  originNames: Set<string>;
  loading: boolean;
  onCreate: (name: string, input: M3uSourceInput) => Promise<void>;
  onUpdate: (name: string, input: M3uSourceInput) => Promise<void>;
  onDelete: (name: string) => Promise<void>;
  searchQuery?: string;
}

export function M3uSourcesSection({
  isExpanded, sources, originNames, loading, onCreate, onUpdate, onDelete, searchQuery,
}: M3uSourcesSectionProps) {
  const { t } = useTranslation('settings');
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingSource, setEditingSource] = useState<M3uSource | null>(null);

  if (!isExpanded) return null;

  const runtimeOnlyNames = sources.filter(s => s.is_runtime && !originNames.has(s.name)).map(s => s.name);
  const existingNames = sources.map(s => s.name);

  const query = (searchQuery ?? '').trim().toLowerCase();
  const visibleSources = query ? sources.filter(s => s.name.toLowerCase().includes(query)) : sources;

  const openCreate = () => {
    setEditingSource(null);
    setIsDialogOpen(true);
  };

  const openEdit = (source: M3uSource) => {
    setEditingSource(source);
    setIsDialogOpen(true);
  };

  const handleDelete = (name: string) => {
    if (!confirm(t('m3uSources.confirmDelete', { name }))) return;
    onDelete(name);
  };

  const originForName = (name: string) => (originNames.has(name) ? 'config' : undefined);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h4 style={{ fontSize: '0.95rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{t('m3uSources.heading')}</h4>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', marginTop: '0.25rem' }}>{t('m3uSources.subtitle')}</p>
        </div>
        <button onClick={openCreate} className="btn-primary">{t('m3uSources.createButton')}</button>
      </div>

      {loading && sources.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>{t('loading')}</div>
      ) : sources.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)' }}>
          <p style={{ color: 'var(--text-secondary)', fontWeight: 600 }}>{t('m3uSources.emptyTitle')}</p>
        </div>
      ) : visibleSources.length === 0 ? null : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {visibleSources.map(source => (
            <div key={source.name} className="filter-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <h3 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{source.name}</h3>
                  <span
                    className={`badge ${source.is_runtime ? 'badge-progress' : 'badge-muted'}`}
                    style={{ marginTop: '0.4rem', fontSize: '0.65rem' }}
                  >
                    {source.is_runtime ? t('origin.interface') : t('origin.config')}
                  </span>
                  {source.is_runtime && originForName(source.name) === 'config' && (
                    <span className="badge badge-progress" style={{ marginTop: '0.4rem', marginLeft: '0.35rem', fontSize: '0.65rem' }}>
                      {t('m3uSources.overridesOrigin')}
                    </span>
                  )}
                </div>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button onClick={() => openEdit(source)} className="btn-secondary" style={{ padding: '0.3rem 0.6rem' }}>
                    {t('m3uSources.edit')}
                  </button>
                  {source.is_runtime && (
                    <button onClick={() => handleDelete(source.name)} className="btn-danger" style={{ padding: '0.3rem 0.6rem' }}>
                      {t('m3uSources.delete')}
                    </button>
                  )}
                </div>
              </div>

              <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', marginTop: '0.5rem', display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                <div>{t('m3uSources.filePathLabel')}: <code>{source.file_path || '—'}</code></div>
                <div>{t('m3uSources.urlLabel')}: <code>{source.url || '—'}</code></div>
                <div>
                  {t('m3uSources.authPasswordLabel')}: {source.has_auth_password ? t('sensitive.set') : t('sensitive.notSet')}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <M3uSourceDialog
        isOpen={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        source={editingSource}
        existingNames={existingNames}
        runtimeOnlyNames={runtimeOnlyNames}
        onSubmit={async (name, input) => {
          if (editingSource) {
            await onUpdate(name, input);
          } else {
            await onCreate(name, input);
          }
        }}
      />
    </div>
  );
}
