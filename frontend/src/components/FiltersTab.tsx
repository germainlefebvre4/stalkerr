import * as Tabs from '@radix-ui/react-tabs';
import { useTranslation } from 'react-i18next';
import { FilterConfig, SystemFilterConfig } from '../types';

interface FiltersTabProps {
  filters: FilterConfig[];
  filtersLoading: boolean;
  onDeleteFilter: (id: number) => void;
  systemFilters: SystemFilterConfig | null;
  systemFiltersLoading: boolean;
  onOpenCreate: () => void;
}

const ATTRIBUTES = ['group_title', 'tvg_name'] as const;

export function FiltersTab({
  filters,
  filtersLoading,
  onDeleteFilter,
  systemFilters,
  systemFiltersLoading,
  onOpenCreate
}: FiltersTabProps) {
  const { t } = useTranslation('filters');
  const isInitialLoading = (filtersLoading && filters.length === 0) || (systemFiltersLoading && !systemFilters);

  return (
    <Tabs.Content value="filters" className="card tab-panel">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>{t('heading')}</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('subtitle')}</p>
        </div>
        <button onClick={onOpenCreate} className="btn-primary">
          {t('configureButton')}
        </button>
      </div>

      {isInitialLoading ? (
        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>{t('loading')}</div>
      ) : (
        <div className="filter-grid">
          {ATTRIBUTES.map(attribute => {
            const origin = systemFilters ? systemFilters[attribute] : null;
            const override = filters.find(f => f.attribute === attribute);
            const originInclude = origin && origin.include_patterns.length > 0 ? origin.include_patterns.join(', ') : t('includeAll');
            const originExclude = origin && origin.exclude_patterns.length > 0 ? origin.exclude_patterns.join(', ') : t('excludeNone');

            return (
              <div key={attribute} className="filter-card">
                <h3 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--primary-slate)' }}>
                  {t(`attributeLabels.${attribute}`)}
                </h3>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.8rem', borderTop: '1px solid var(--border-color)', paddingTop: '0.75rem', marginTop: '0.5rem' }}>
                  <span className="badge badge-progress" style={{ alignSelf: 'flex-start', fontSize: '0.65rem', padding: '0.15rem 0.5rem' }}>
                    {t('originBadge')}
                  </span>
                  <div>
                    <strong style={{ color: 'var(--status-success-text)' }}>{t('include')}</strong>
                    <code style={{ background: 'var(--status-success-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                      {originInclude}
                    </code>
                  </div>
                  <div>
                    <strong style={{ color: 'var(--status-failed-text)' }}>{t('exclude')}</strong>
                    <code style={{ background: 'var(--status-failed-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                      {originExclude}
                    </code>
                  </div>
                </div>

                {override && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.8rem', borderTop: '1px solid var(--border-color)', paddingTop: '0.75rem', marginTop: '0.75rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.3rem' }}>
                        <span className="badge badge-success" style={{ alignSelf: 'flex-start', fontSize: '0.65rem', padding: '0.15rem 0.5rem' }}>
                          {t('overrideBadge')}
                        </span>
                        <strong style={{ color: 'var(--primary-slate)' }}>{override.name}</strong>
                      </div>
                      <button onClick={() => onDeleteFilter(override.id)} className="btn-danger" style={{ padding: '0.3rem 0.5rem' }} title={t('deleteTitle')}>
                        {t('delete')}
                      </button>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--status-success-text)' }}>{t('include')}</strong>
                      <code style={{ background: 'var(--status-success-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                        {override.include_patterns || t('includeAll')}
                      </code>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--status-failed-text)' }}>{t('exclude')}</strong>
                      <code style={{ background: 'var(--status-failed-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                        {override.exclude_patterns || t('excludeNone')}
                      </code>
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </Tabs.Content>
  );
}
