import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { StatsResponse } from '../types';

export function useHealthAndStats() {
  const [healthStatus, setHealthStatus] = useState<'healthy' | 'unhealthy' | 'checking'>('checking');
  const [stats, setStats] = useState<StatsResponse | null>(null);

  const fetchHealth = useCallback(() => {
    api.getHealth()
      .then(data => setHealthStatus(data.status === 'healthy' ? 'healthy' : 'unhealthy'))
      .catch(() => setHealthStatus('unhealthy'));
  }, []);

  const fetchStats = useCallback(() => {
    api.getStats()
      .then(data => setStats(data))
      .catch(() => {});
  }, []);

  useEffect(() => {
    fetchHealth();
    fetchStats();
    const interval = setInterval(() => {
      fetchHealth();
      fetchStats();
    }, 10000);
    return () => clearInterval(interval);
  }, [fetchHealth, fetchStats]);

  const getDownloadSuccessRatio = useCallback(() => {
    if (!stats || !stats.by_state) return '100%';
    const downloaded = stats.by_state.downloaded || 0;
    const failed = stats.by_state.failed || 0;
    const total = downloaded + failed;
    if (total === 0) return '100%';
    return `${Math.round((downloaded / total) * 100)}%`;
  }, [stats]);

  return {
    healthStatus,
    stats,
    fetchStats,
    getDownloadSuccessRatio
  };
}
