import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { ProcessingLog } from '../types';
import { COMBINED_PROBLEM_FILTER } from './useErrorsTab';

// The Home tab's own data layer: the latest processing run, plus the
// downloads and errors *counts* (not their item lists - `PaginatedResponse.Total`
// is populated regardless of `limit`, so a `limit=1` request is enough).
// Gated on tab activation like every other non-always-on tab's hook
// (useDownloads, useErrorsTab): fetches only while the Home tab is active.
export function useHomeDashboard(isActive: boolean) {
  const [latestLog, setLatestLog] = useState<ProcessingLog | null>(null);
  const [latestLogLoading, setLatestLogLoading] = useState(false);
  const [downloadsTotal, setDownloadsTotal] = useState(0);
  const [errorsTotal, setErrorsTotal] = useState(0);

  const fetchHomeDashboard = useCallback(() => {
    setLatestLogLoading(true);
    api.getLogs(1)
      .then(data => setLatestLog(data.data?.[0] ?? null))
      .catch(() => setLatestLog(null))
      .finally(() => setLatestLogLoading(false));

    api.getDownloads(1)
      .then(data => setDownloadsTotal(data.total || 0))
      .catch(() => {});

    api.getDownloads(1, undefined, undefined, COMBINED_PROBLEM_FILTER)
      .then(data => setErrorsTotal(data.total || 0))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchHomeDashboard);
  }, [isActive, fetchHomeDashboard]);

  return {
    latestLog,
    latestLogLoading,
    downloadsTotal,
    errorsTotal,
    fetchHomeDashboard,
  };
}
