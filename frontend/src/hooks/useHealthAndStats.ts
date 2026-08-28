import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { StatsResponse } from '../types';

export function useHealthAndStats() {
  const [stats, setStats] = useState<StatsResponse | null>(null);


  const fetchStats = useCallback(() => {
    api.getStats()
      .then(data => setStats(data))
      .catch(() => {});
  }, []);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(() => {
      fetchStats();
    }, 10000);
    return () => clearInterval(interval);
  }, [fetchStats]);

  const getDownloadSuccessRatio = useCallback(() => {
    if (!stats || !stats.by_state) return '100%';
    const downloaded = stats.by_state.downloaded || 0;
    const failed = stats.by_state.failed || 0;
    const total = downloaded + failed;
    if (total === 0) return '100%';
    return `${Math.round((downloaded / total) * 100)}%`;
  }, [stats]);

  return {
    stats,
    fetchStats,
    getDownloadSuccessRatio
  };
}
