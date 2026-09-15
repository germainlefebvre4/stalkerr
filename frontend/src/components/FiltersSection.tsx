import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { FilterConfig, FilterOriginEntry, M3uSource, FilterDryRunSummaryResponse } from '../types';
import { api } from '../services/api';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';
import { FilterDryRunPanel, FilterDryRunStatus } from './FilterDryRunPanel';

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

// A read-only dry-run "Tester" for one filter card block (origin or active
// override): its own source selection and its own result panel, testing
// exactly that block's existing patterns without opening the create dialog
// or affecting any saved filter. See frontend-filters-management's "Dry-Run
// Filter Testing" - "Testing an already-saved filter from its card".
function FilterCardTester({ attribute, includePatterns, excludePatterns, sources }: {
  attribute: string;
  includePatterns: string;
  excludePatterns: string;
  sources: M3uSource[];
}) {
  const { t } = useTranslation('filters');
  const translateApiError = useApiErrorMessage();
  const [testSourceName, setTestSourceName] = useState('');
  const [status, setStatus] = useState<FilterDryRunStatus>('idle');
  const [summary, setSummary] = useState<FilterDryRunSummaryResponse | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleTest = () => {
    if (!testSourceName) return;
    setStatus('testing');
    setErrorMessage(null);

    api.dryRunFilter({
      source_name: testSourceName,
      attribute,
      include_patterns: includePatterns,
      exclude_patterns: excludePatterns,
    })
      .then(res => {
        if (res.no_archive) {
          setStatus('no_archive');
          return;
        }
        setSummary(res as FilterDryRunSummaryResponse);
        setStatus('summary');
      })
      .catch((err: unknown) => {
        setErrorMessage(translateApiError(err));
        setStatus('error');
      });
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
      <div style={{ display: 'flex', gap: '0.5rem' }}>
        <select
          value={testSourceName}
          onChange={e => { setTestSourceName(e.target.value); setStatus('idle'); setSummary(null); setErrorMessage(null); }}
          className="custom-select"
          style={{ flex: 1, fontSize: '0.8rem' }}
        >
          <option value="">{t('dryRun.sourcePlaceholder')}</option>
          {sources.map(source => (
            <option key={source.name} value={source.name}>{source.name}</option>
          ))}
        </select>
        <button
          type="button"
          onClick={handleTest}
          disabled={!testSourceName || status === 'testing'}
          className="btn-secondary"
          style={{ padding: '0.3rem 0.6rem', fontSize: '0.75rem' }}
        >
          {t('dryRun.testButton')}
        </button>
      </div>
      <FilterDryRunPanel
        status={status}
        summary={summary}
        errorMessage={errorMessage}
        sourceName={testSourceName}
        attribute={attribute}
        includePatterns={includePatterns}
        excludePatterns={excludePatterns}
      />
    </div>
  );
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

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h4 style={{ fontSize: '0.95rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{t('heading')}</h4>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('subtitle')}</p>
        </div>
        <button onClick={onOpenCreate} className="btn-primary">
          {t('configureButton')}
        </button>
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
                <FilterCardTester
                  attribute={attribute}
                  includePatterns={origin ? joinPatterns(origin.include_patterns) : ''}
                  excludePatterns={origin ? joinPatterns(origin.exclude_patterns) : ''}
                  sources={sources}
                />
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
                  <FilterCardTester
                    attribute={attribute}
                    includePatterns={override.include_patterns || ''}
                    excludePatterns={override.exclude_patterns || ''}
                    sources={sources}
                  />
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
