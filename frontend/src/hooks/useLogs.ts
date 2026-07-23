import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { ProcessingLog } from '../types';

export function useLogs(isActive: boolean) {
  const [logs, setLogs] = useState<ProcessingLog[]>([]);
  const [logsLoading, setLogsLoading] = useState(false);

  const fetchLogs = useCallback(() => {
    setLogsLoading(true);
    api.getLogs(20)
      .then(data => setLogs(data.data || []))
      .catch(() => {})
      .finally(() => setLogsLoading(false));
  }, []);

  useEffect(() => {
    if (!isActive) return;
    fetchLogs();
    const interval = setInterval(fetchLogs, 5000);
    return () => clearInterval(interval);
  }, [isActive, fetchLogs]);

  return {
    logs,
    logsLoading,
    fetchLogs,
  };
}
