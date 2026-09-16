import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { api } from '../services/api';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';
import { FilterDryRunSummaryResponse, FilterDryRunSearchResponse } from '../types';

export type FilterDryRunStatus = 'idle' | 'testing' | 'summary' | 'no_archive' | 'error';

interface FilterDryRunPanelProps {
  status: FilterDryRunStatus;
  summary: FilterDryRunSummaryResponse | null;
  errorMessage: string | null;
  // The exact params the last test was run with - reused to re-issue the
  // content-search request without needing the parent to own that state.
  sourceName: string;
  attribute: string;
  includePatterns: string;
  excludePatterns: string;
}

// Renders the result of a filter dry-run test (idle/testing/summary/error/
// no-archive) and, once a summary is shown, an optional content-search field
// that re-queries the dry-run endpoint scoped to the tested attribute. Shared
// between CreateFilterDialog (in-progress patterns) and FiltersSection cards
// (already-saved patterns). See frontend-filters-management's "Dry-Run
// Filter Testing" and "Dry-Run Content Search" requirements.
export function FilterDryRunPanel({ status, summary, errorMessage, sourceName, attribute, includePatterns, excludePatterns }: FilterDryRunPanelProps) {
  const { t } = useTranslation('filters');
  const translateApiError = useApiErrorMessage();

  const [searchTerm, setSearchTerm] = useState('');
  const [searchResult, setSearchResult] = useState<FilterDryRunSearchResponse | null>(null);
  const [searching, setSearching] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);

  // A fresh test (new source/attribute/patterns) invalidates any previous
  // search result. Adjusted during render rather than a useEffect+setState
  // pair (see CreateFilterDialog's dryRunInputsKey for the same convention).
  const dryRunResultKey = `${sourceName}|${attribute}|${includePatterns}|${excludePatterns}|${status}`;
  const [lastDryRunResultKey, setLastDryRunResultKey] = useState(dryRunResultKey);
  if (dryRunResultKey !== lastDryRunResultKey) {
    setLastDryRunResultKey(dryRunResultKey);
    setSearchTerm('');
    setSearchResult(null);
    setSearchError(null);
  }

  const handleSearchChange = (term: string) => {
    setSearchTerm(term);
    if (!term.trim()) {
      setSearchResult(null);
      setSearchError(null);
      return;
    }
    setSearching(true);
    setSearchError(null);
    api.dryRunFilter({
      source_name: sourceName,
      attribute,
      include_patterns: includePatterns,
      exclude_patterns: excludePatterns,
      search: term,
    })
      .then(res => setSearchResult(res as FilterDryRunSearchResponse))
      .catch((err: unknown) => setSearchError(translateApiError(err)))
      .finally(() => setSearching(false));
  };

  if (status === 'idle') return null;

  if (status === 'testing') {
    return <p className="settings-field-hint">{t('dryRun.testing')}</p>;
  }

  if (status === 'error') {
    return <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>{errorMessage}</p>;
  }

  if (status === 'no_archive') {
    return <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>{t('dryRun.noArchive')}</p>;
  }

  if (!summary) return null;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem', fontSize: '0.8rem' }}>
      <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
        <span className="badge badge-neutral">{t('dryRun.totalLines', { count: summary.total_lines })}</span>
        <span className="badge badge-success">{t('dryRun.matchedCount', { count: summary.matched_count })}</span>
        <span className="badge badge-failed">{t('dryRun.excludedCount', { count: summary.excluded_count })}</span>
      </div>

      {(summary.top_matched.length > 0 || summary.top_excluded.length > 0) && (
        <div className="dry-run-columns">
          {summary.top_matched.length > 0 && (
            <div>
              <strong style={{ color: 'var(--status-success-text)' }}>{t('dryRun.topMatched')}</strong>
              <ul style={{ margin: '0.25rem 0 0', paddingLeft: '1.1rem' }}>
                {summary.top_matched.map(v => (
                  <li key={v.value}>{v.value} ({v.count})</li>
                ))}
              </ul>
            </div>
          )}

          {summary.top_excluded.length > 0 && (
            <div>
              <strong style={{ color: 'var(--status-failed-text)' }}>{t('dryRun.topExcluded')}</strong>
              <ul style={{ margin: '0.25rem 0 0', paddingLeft: '1.1rem' }}>
                {summary.top_excluded.map(v => (
                  <li key={v.value}>{v.value} ({v.count})</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', marginTop: '0.25rem' }}>
        <label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)' }}>{t('dryRun.searchLabel')}</label>
        <input
          type="text"
          className="custom-input"
          placeholder={t('dryRun.searchPlaceholder')}
          value={searchTerm}
          onChange={e => handleSearchChange(e.target.value)}
        />
      </div>

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
                        <td>{attribute === 'tvg_name' ? line.tvg_name : line.group_title}</td>
                        <td>
                          <span className={`badge ${line.matched ? 'badge-success' : 'badge-failed'}`}>
                            {line.matched ? t('dryRun.wouldMatch') : t('dryRun.wouldBeExcluded')}
                          </span>
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
