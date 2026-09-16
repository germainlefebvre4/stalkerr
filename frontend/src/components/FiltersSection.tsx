import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { FilterConfig, FilterOriginEntry, M3uSource, FilterTestTarget } from '../types';
import { FilterTestDrawer } from './FilterTestDrawer';

interface FiltersSectionProps {
  isExpanded: boolean;
  filters: FilterConfig[];
  filterOrigin: FilterOriginEntry[];
  filtersLoading: boolean;
  onDeleteFilter: (id: number) => void;
  onOpenCreate: () => void;
  sources: M3uSource[];
  searchQuery?: string;
}

const ATTRIBUTES = ['group_title', 'tvg_name'] as const;

function joinPatterns(patterns: string[]): string {
  return patterns.join(', ');
}

// The "Filtres" section within the Configuration page's "Contenu" tab: one
// block per attribute (Group Title, TVG Name) showing the origin config.yml
// patterns alongside the active runtime override, if any. See the Filters
// List View requirement in frontend-filters-management.
export function FiltersSection({
  isExpanded,
  filters,
  filterOrigin,
  filtersLoading,
  onDeleteFilter,
  onOpenCreate,
  sources,
  searchQuery,
}: FiltersSectionProps) {
  const { t } = useTranslation('filters');
  const [testTarget, setTestTarget] = useState<FilterTestTarget | null>(null);

  if (!isExpanded) return null;

  const attributeLabel = (attribute: string) => (attribute === 'group_title' ? t('groupTitleLabel') : t('tvgNameLabel'));

  const query = (searchQuery ?? '').trim().toLowerCase();
  const groups = ATTRIBUTES
    .map(attribute => ({
      attribute,
      origin: filterOrigin.find(o => o.attribute === attribute),
      override: filters.find(f => f.attribute === attribute),
    }))
    .filter(group => {
      if (!query) return true;
      const label = attributeLabel(group.attribute).toLowerCase();
      return label.includes(query) || (group.override?.name.toLowerCase().includes(query) ?? false);
    });

  const testSingle = (attribute: 'group_title' | 'tvg_name', includePatterns: string, excludePatterns: string, isOverride: boolean) => {
    setTestTarget({
      mode: 'single',
      attribute,
      includePatterns,
      excludePatterns,
      label: t(isOverride ? 'dryRun.titleOverride' : 'dryRun.titleOrigin', { attribute: attributeLabel(attribute) }),
    });
  };

  // Resolves each attribute's effective patterns - the active override if
  // one exists, otherwise the origin config.yml patterns - the same
  // resolution already used per card above.
  const effectivePatterns = (attribute: 'group_title' | 'tvg_name') => {
    const override = filters.find(f => f.attribute === attribute);
    if (override) {
      return { include: override.include_patterns || '', exclude: override.exclude_patterns || '' };
    }
    const origin = filterOrigin.find(o => o.attribute === attribute);
    return {
      include: origin ? joinPatterns(origin.include_patterns) : '',
      exclude: origin ? joinPatterns(origin.exclude_patterns) : '',
    };
  };

  const testAll = () => {
    const groupTitle = effectivePatterns('group_title');
    const tvgName = effectivePatterns('tvg_name');
    setTestTarget({
      mode: 'combined',
      groupTitleInclude: groupTitle.include,
      groupTitleExclude: groupTitle.exclude,
      tvgNameInclude: tvgName.include,
      tvgNameExclude: tvgName.exclude,
      label: t('dryRun.titleCombined'),
    });
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h4 style={{ fontSize: '0.95rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{t('heading')}</h4>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('subtitle')}</p>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button onClick={testAll} className="btn-secondary">
            {t('dryRun.testAllButton')}
          </button>
          <button onClick={onOpenCreate} className="btn-primary">
            {t('configureButton')}
          </button>
        </div>
      </div>

      {filtersLoading && filterOrigin.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>{t('loading')}</div>
      ) : groups.length === 0 ? null : (
        <div className="filter-grid">
          {groups.map(({ attribute, origin, override }) => (
            <div key={attribute} className="filter-card">
              <h3 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{attributeLabel(attribute)}</h3>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.8rem' }}>
                <span className="badge badge-neutral" style={{ alignSelf: 'flex-start', fontSize: '0.65rem' }}>
                  {t('originLabel')}
                </span>
                <div>
                  <strong style={{ color: 'var(--status-success-text)' }}>{t('include')}</strong>
                  <code style={{ background: 'var(--status-success-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                    {origin && origin.include_patterns.length > 0 ? joinPatterns(origin.include_patterns) : t('includeAll')}
                  </code>
                </div>
                <div>
                  <strong style={{ color: 'var(--status-failed-text)' }}>{t('exclude')}</strong>
                  <code style={{ background: 'var(--status-failed-bg)', padding: '0.15rem 0.4rem', borderRadius: '4px', wordBreak: 'break-all' }}>
                    {origin && origin.exclude_patterns.length > 0 ? joinPatterns(origin.exclude_patterns) : t('excludeNone')}
                  </code>
                </div>
                <button
                  type="button"
                  onClick={() => testSingle(attribute, origin ? joinPatterns(origin.include_patterns) : '', origin ? joinPatterns(origin.exclude_patterns) : '', false)}
                  className="btn-secondary"
                  style={{ alignSelf: 'flex-start', padding: '0.3rem 0.6rem', fontSize: '0.75rem' }}
                >
                  {t('dryRun.testButton')}
                </button>
              </div>

              {override && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.8rem', borderTop: '1px solid var(--border-color)', paddingTop: '0.75rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <span className="badge badge-progress" style={{ fontSize: '0.65rem' }}>{t('overrideLabel')}</span>
                    <button onClick={() => onDeleteFilter(override.id)} className="btn-danger" style={{ padding: '0.3rem 0.5rem' }} title={t('deleteTitle')}>
                      {t('delete')}
                    </button>
                  </div>
                  <div style={{ fontWeight: 700, color: 'var(--primary-slate)' }}>{override.name}</div>
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
                  <button
                    type="button"
                    onClick={() => testSingle(attribute, override.include_patterns || '', override.exclude_patterns || '', true)}
                    className="btn-secondary"
                    style={{ alignSelf: 'flex-start', padding: '0.3rem 0.6rem', fontSize: '0.75rem' }}
                  >
                    {t('dryRun.testButton')}
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      <FilterTestDrawer
        target={testTarget}
        onOpenChange={(open) => { if (!open) setTestTarget(null); }}
        sources={sources}
      />
    </div>
  );
}
