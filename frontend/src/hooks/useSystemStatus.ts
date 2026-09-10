import { useState, useCallback } from 'react';
import { api, ApiError } from '../services/api';
import { SystemStatusResponse } from '../types';

// No effect fetches on mount and there is no polling interval: the caller
// (the header icon's dialog) is solely responsible for invoking fetchStatus,
// on open and again on the dialog's manual refresh action.
export function useSystemStatus() {
  const [status, setStatus] = useState<SystemStatusResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = useCallback(() => {
    setLoading(true);
    setError(null);
    api.getSystemStatus()
      .then(data => setStatus(data))
      .catch(err => {
        setStatus(null);
        setError(err instanceof ApiError ? err.code : 'generic');
      })
      .finally(() => setLoading(false));
  }, []);

  return { status, loading, error, fetchStatus };
}
