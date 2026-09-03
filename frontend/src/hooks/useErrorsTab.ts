import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { DownloadEnriched } from '../types';
import { useURLState, URLStateSchema } from './useURLState';

const VALID_LIMITS = [20, 50, 100];
const DEFAULT_LIMIT = 20;

export const COMBINED_PROBLEM_FILTER = 'missing_year,year_mismatch,unknown_format';
const VALID_REASONS = ['', 'unknown_format', 'missing_year', 'year_mismatch'];

const ERRORS_URL_SCHEMA = {
  errReason: {
    default: '',
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: (v: string) => VALID_REASONS.includes(v),
  },
  errPage: {
    default: 1,
    parse: (raw: string) => parseInt(raw, 10),
    serialize: String,
    isValid: (v: number) => !isNaN(v) && v > 0,
  },
  errLimit: {
    default: DEFAULT_LIMIT,
    parse: (raw: string) => parseInt(raw, 10),
    serialize: String,
    isValid: (v: number) => VALID_LIMITS.includes(v),
  },
} satisfies URLStateSchema;

// Empty reason ("Tous") maps to the combined OR-filter; any other reason is
// already the single `problem` value to send as-is.
function problemForReason(reason: string): string {
  return reason === '' ? COMBINED_PROBLEM_FILTER : reason;
}

export function useErrorsTab(isActive: boolean) {
  const [errors, setErrors] = useState<DownloadEnriched[]>([]);
  const [errorsLoading, setErrorsLoading] = useState(false);
  const [errorsTotal, setErrorsTotal] = useState(0);
  const [urlState, patchURLState] = useURLState(ERRORS_URL_SCHEMA);

  const reasonFilter = urlState.errReason;
  const errorsPage = urlState.errPage;
  const errorsLimit = urlState.errLimit;

  // Changing the reason resets pagination to page 1.
  const setReasonFilter = useCallback(
    (value: string) => patchURLState({ errReason: value, errPage: 1 }),
    [patchURLState]
  );

  const setErrorsPage = useCallback(
    (valueOrUpdater: number | ((prevPage: number) => number)) => {
      const next = typeof valueOrUpdater === 'function' ? valueOrUpdater(urlState.errPage) : valueOrUpdater;
      patchURLState({ errPage: next });
    },
    [patchURLState, urlState.errPage]
  );

  const setErrorsLimit = useCallback(
    (limit: number) => patchURLState({ errLimit: limit, errPage: 1 }),
    [patchURLState]
  );

  const fetchErrors = useCallback(() => {
    setErrorsLoading(true);
    const offset = (errorsPage - 1) * errorsLimit;
    api.getDownloads(errorsLimit, '', '', problemForReason(reasonFilter), offset)
      .then(data => {
        setErrors(data.data || []);
        setErrorsTotal(data.total || 0);
      })
      .catch(() => {})
      .finally(() => setErrorsLoading(false));
  }, [reasonFilter, errorsPage, errorsLimit]);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchErrors);
  }, [isActive, fetchErrors]);

  return {
    errors,
    errorsLoading,
    errorsTotal,
    reasonFilter,
    setReasonFilter,
    errorsPage,
    setErrorsPage,
    errorsLimit,
    setErrorsLimit,
    fetchErrors,
  };
}
