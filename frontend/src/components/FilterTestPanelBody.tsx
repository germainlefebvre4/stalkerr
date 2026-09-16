import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import { api } from '../services/api';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';
import {
  FilterTestTarget,
  M3uSource,
  FilterDryRunRequest,
  FilterDryRunSummaryResponse,
  FilterDryRunCombinedSummaryResponse,
  FilterDryRunSearchResponse,
  FilterDryRunValueCount,
} from '../types';

type DryRunStatus = 'idle' | 'testing' | 'summary' | 'no_archive' | 'error';

interface FilterTestPanelBodyProps {
  target: FilterTestTarget;
  sources: M3uSource[];
}

// Builds the dry-run request for target, optionally adding a content search
// scoped to searchAttribute (required for combined-mode search).
function buildRequest(target: FilterTestTarget, sourceName: string, search?: string, searchAttribute?: string): FilterDryRunRequest {
  const base: FilterDryRunRequest = target.mode === 'combined'
    ? {
      source_name: sourceName,
      attributes: ['group_title', 'tvg_name'],
      group_title_include_patterns: target.groupTitleInclude,
      group_title_exclude_patterns: target.groupTitleExclude,
      tvg_name_include_patterns: target.tvgNameInclude,
      tvg_name_exclude_patterns: target.tvgNameExclude,
    }
    : {
      source_name: sourceName,
      attributes: [target.attribute],
      ...(target.attribute === 'tvg_name'
        ? { tvg_name_include_patterns: target.includePatterns, tvg_name_exclude_patterns: target.excludePatterns }
        : { group_title_include_patterns: target.includePatterns, group_title_exclude_patterns: target.excludePatterns }),
    };

  if (!search) return base;
  return { ...base, search, search_attribute: searchAttribute };
}

function verdictLabel(verdict: string | undefined, t: TFunction): string {
  switch (verdict) {
    case 'kept': return t('dryRun.verdictKept');
    case 'excluded_by_group_title': return t('dryRun.verdictExcludedByGroupTitle');
    case 'excluded_by_tvg_name': return t('dryRun.verdictExcludedByTvgName');
    case 'excluded_by_both': return t('dryRun.verdictExcludedByBoth');
    default: return '';
  }
}

// The top-20 matched/excluded values for one attribute, as the existing
// two-column grid (see improve-filter-dryrun-results-layout).
function TopValuesGrid({ matched, excluded, t }: { matched: FilterDryRunValueCount[]; excluded: FilterDryRunValueCount[]; t: TFunction }) {
  if (matched.length === 0 && excluded.length === 0) return null;
  return (
    <div className="dry-run-columns">
      {matched.length > 0 && (
        <div>
          <strong style={{ color: 'var(--status-success-text)' }}>{t('dryRun.topMatched')}</strong>
          <ul style={{ margin: '0.25rem 0 0', paddingLeft: '1.1rem' }}>
            {matched.map(v => (
              <li key={v.value}>{v.value} ({v.count})</li>
            ))}
          </ul>
        </div>
      )}
      {excluded.length > 0 && (
        <div>
          <strong style={{ color: 'var(--status-failed-text)' }}>{t('dryRun.topExcluded')}</strong>
          <ul style={{ margin: '0.25rem 0 0', paddingLeft: '1.1rem' }}>
            {excluded.map(v => (
              <li key={v.value}>{v.value} ({v.count})</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

function SingleSummaryView({ summary, t }: { summary: FilterDryRunSummaryResponse; t: TFunction }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem', fontSize: '0.8rem' }}>
      <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
        <span className="badge badge-neutral">{t('dryRun.totalLines', { count: summary.total_lines })}</span>
        <span className="badge badge-success">{t('dryRun.matchedCount', { count: summary.matched_count })}</span>
        <span className="badge badge-failed">{t('dryRun.excludedCount', { count: summary.excluded_count })}</span>
      </div>
      <TopValuesGrid matched={summary.top_matched} excluded={summary.top_excluded} t={t} />
    </div>
  );
}

function CombinedSummaryView({ summary, t }: { summary: FilterDryRunCombinedSummaryResponse; t: TFunction }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem', fontSize: '0.8rem' }}>
      <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
        <span className="badge badge-neutral">{t('dryRun.totalLines', { count: summary.total_lines })}</span>
        <span className="badge badge-success">{t('dryRun.combinedKept', { count: summary.kept_count })}</span>
        <span className="badge badge-failed">{t('dryRun.combinedExcludedGroupTitleOnly', { count: summary.excluded_by_group_title_only })}</span>
        <span className="badge badge-failed">{t('dryRun.combinedExcludedTvgNameOnly', { count: summary.excluded_by_tvg_name_only })}</span>
        <span className="badge badge-failed">{t('dryRun.combinedExcludedBoth', { count: summary.excluded_by_both })}</span>
      </div>
      <div>
        <h4 style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--primary-slate)', marginBottom: '0.4rem' }}>{t('groupTitleLabel')}</h4>
        <TopValuesGrid matched={summary.group_title_top_matched} excluded={summary.group_title_top_excluded} t={t} />
      </div>
      <div>
        <h4 style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--primary-slate)', marginBottom: '0.4rem' }}>{t('tvgNameLabel')}</h4>
        <TopValuesGrid matched={summary.tvg_name_top_matched} excluded={summary.tvg_name_top_excluded} t={t} />
      </div>
    </div>
  );
}

// The M3U source selector, the "Tester" trigger, and the mode-dependent
// result display (summary, top-value columns, search field, results table).
// Pure content - no Dialog.Root - so it can be embedded by FilterTestDrawer
// with different modal/overlay settings. See design.md's component split.
export function FilterTestPanelBody({ target, sources }: FilterTestPanelBodyProps) {
  const { t } = useTranslation('filters');
  const translateApiError = useApiErrorMessage();

  const [testSourceName, setTestSourceName] = useState('');
  const [status, setStatus] = useState<DryRunStatus>('idle');
  const [summary, setSummary] = useState<FilterDryRunSummaryResponse | FilterDryRunCombinedSummaryResponse | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [searchAttribute, setSearchAttribute] = useState<'group_title' | 'tvg_name'>(target.mode === 'single' ? target.attribute : 'group_title');
  const [searchTerm, setSearchTerm] = useState('');
  const [searchResult, setSearchResult] = useState<FilterDryRunSearchResponse | null>(null);
  const [searching, setSearching] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);

  const targetKey = JSON.stringify(target);

  // A new target (different card, different in-progress patterns, or
  // "Tester l'ensemble") invalidates the previous test result and search.
  useEffect(() => {
    setStatus('idle');
    setSummary(null);
    setErrorMessage(null);
    setSearchAttribute(target.mode === 'single' ? target.attribute : 'group_title');
    setSearchTerm('');
    setSearchResult(null);
    setSearchError(null);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [targetKey]);

  const runSearch = (term: string, attribute: 'group_title' | 'tvg_name') => {
    if (!term.trim()) {
      setSearchResult(null);
      setSearchError(null);
      return;
    }
    setSearching(true);
    setSearchError(null);
    api.dryRunFilter(buildRequest(target, testSourceName, term, attribute))
      .then(res => setSearchResult(res as FilterDryRunSearchResponse))
      .catch((err: unknown) => setSearchError(translateApiError(err)))
      .finally(() => setSearching(false));
  };

  const handleTest = () => {
    if (!testSourceName) return;
    setStatus('testing');
    setErrorMessage(null);

    api.dryRunFilter(buildRequest(target, testSourceName))
      .then(res => {
        if (res.no_archive) {
          setStatus('no_archive');
          return;
        }
        setSummary(res as FilterDryRunSummaryResponse | FilterDryRunCombinedSummaryResponse);
        setStatus('summary');
      })
      .catch((err: unknown) => {
        setErrorMessage(translateApiError(err));
        setStatus('error');
      });
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
        <label style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('dryRun.sourceLabel')}</label>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <select
            value={testSourceName}
            onChange={e => setTestSourceName(e.target.value)}
            className="custom-select"
            style={{ flex: 1 }}
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
          >
            {t('dryRun.testButton')}
          </button>
        </div>
      </div>

      {status === 'testing' && <p className="settings-field-hint">{t('dryRun.testing')}</p>}
      {status === 'error' && <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>{errorMessage}</p>}
      {status === 'no_archive' && <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>{t('dryRun.noArchive')}</p>}

      {status === 'summary' && summary && (
        target.mode === 'combined'
          ? <CombinedSummaryView summary={summary as FilterDryRunCombinedSummaryResponse} t={t} />
          : <SingleSummaryView summary={summary as FilterDryRunSummaryResponse} t={t} />
      )}

      {status === 'summary' && summary && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', marginTop: '0.25rem' }}>
          <label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('dryRun.searchLabel')}</label>

          {target.mode === 'combined' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.3rem' }}>
              <label style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{t('dryRun.searchAttributeLabel')}</label>
              <select
                value={searchAttribute}
                onChange={e => {
                  const attribute = e.target.value as 'group_title' | 'tvg_name';
                  setSearchAttribute(attribute);
                  runSearch(searchTerm, attribute);
                }}
                className="custom-select"
                style={{ fontSize: '0.8rem' }}
              >
                <option value="group_title">{t('groupTitleLabel')}</option>
                <option value="tvg_name">{t('tvgNameLabel')}</option>
              </select>
            </div>
          )}

          <input
            type="text"
            className="custom-input"
            placeholder={t('dryRun.searchPlaceholder')}
            value={searchTerm}
            onChange={e => { setSearchTerm(e.target.value); runSearch(e.target.value, searchAttribute); }}
          />
        </div>
      )}

      {searching && <p className="settings-field-hint">{t('dryRun.testing')}</p>}

      {searchError && (
        <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>{searchError}</p>
      )}

      {!searching && !searchError && searchResult && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.3rem' }}>
          {searchResult.results.length === 0 ? (
            <span className="settings-field-hint">{t('dryRun.searchNoResults')}</span>
          ) : (
            <>
              <div className="dry-run-search-table-container">
                <table className="dry-run-search-table">
                  <tbody>
                    {searchResult.results.map((line, idx) => (
                      <tr key={`${line.group_title}-${line.tvg_name}-${idx}`}>
                        <td>{searchAttribute === 'tvg_name' ? line.tvg_name : line.group_title}</td>
                        <td>
                          {target.mode === 'combined' ? (
                            <span className={`badge ${line.verdict === 'kept' ? 'badge-success' : 'badge-failed'}`}>
                              {verdictLabel(line.verdict, t)}
                            </span>
                          ) : (
                            <span className={`badge ${line.matched ? 'badge-success' : 'badge-failed'}`}>
                              {line.matched ? t('dryRun.wouldMatch') : t('dryRun.wouldBeExcluded')}
                            </span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {searchResult.truncated && <span className="settings-field-hint">{t('dryRun.truncated')}</span>}
            </>
          )}
        </div>
      )}
    </div>
  );
}
