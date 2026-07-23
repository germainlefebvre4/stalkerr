import * as Tabs from '@radix-ui/react-tabs';
import { FilterConfig } from '../types';

interface FiltersTabProps {
  filters: FilterConfig[];
  filtersLoading: boolean;
  onDeleteFilter: (id: number) => void;
  onOpenCreate: () => void;
}

export function FiltersTab({
  filters,
  filtersLoading,
  onDeleteFilter,
  onOpenCreate
}: FiltersTabProps) {
  return (
    <Tabs.Content value="filters" className="card" style={{ padding: '2rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>Expressions régulières d'Inclusion & d'Exclusion</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>Gérez les patterns de filtrage appliqués aux flux M3U lors de l'enrichissement</p>
        </div>
        <button onClick={onOpenCreate} className="btn-primary">
          + Configurer un Filtre
        </button>
      </div>

      {filtersLoading && filters.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>Chargement des filtres...</div>
      ) : filters.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '4rem', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)' }}>
          <p style={{ color: 'var(--text-secondary)', fontWeight: 600 }}>Aucun filtre de tri configuré</p>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.8rem', marginTop: '0.25rem' }}>Tous les flux M3U importés seront acceptés par défaut</p>
        </div>
      ) : (
        <div className="filter-grid">
          {filters.map(filter => (
            <div key={filter.id} className="filter-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <h3 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{filter.name}</h3>
                  <span className="badge badge-progress" style={{ marginTop: '0.4rem', fontSize: '0.65rem', padding: '0.15rem 0.5rem' }}>
                    Attribut: {filter.attribute}
                  </span>
                </div>
                <button onClick={() => onDeleteFilter(filter.id)} className="btn-danger" style={{ padding: '0.3rem 0.5rem' }} title="Supprimer ce filtre">
                  Supprimer 🗑️
                </button>
              </div>
              
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.8rem', borderTop: '1px solid var(--border-color)', paddingTop: '0.75rem' }}>
                <div>
                  <strong style={{ color: 'var(--status-success-text)' }}>🟢 Inclure : </strong>
                  <code style={{ background: 'var(--status-success-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                    {filter.include_patterns || '* (Tout accepter)'}
                  </code>
                </div>
                <div>
                  <strong style={{ color: 'var(--status-failed-text)' }}>🔴 Exclure : </strong>
                  <code style={{ background: 'var(--status-failed-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                    {filter.exclude_patterns || '- (Aucun)'}
                  </code>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </Tabs.Content>
  );
}
