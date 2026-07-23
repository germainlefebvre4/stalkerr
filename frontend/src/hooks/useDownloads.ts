import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { DownloadEnriched, ConfigPaths } from '../types';

export function useDownloads(isActive: boolean) {
  const [downloads, setDownloads] = useState<DownloadEnriched[]>([]);
  const [downloadsLoading, setDownloadsLoading] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [problemFilter, setProblemFilter] = useState<string>('');
  const [configPaths, setConfigPaths] = useState<ConfigPaths | null>(null);

  const fetchDownloads = useCallback(() => {
    setDownloadsLoading(true);
    api.getDownloads(20, statusFilter, typeFilter, problemFilter)
      .then(data => setDownloads(data.data || []))
      .catch(() => {})
      .finally(() => setDownloadsLoading(false));
  }, [statusFilter, typeFilter, problemFilter]);

  useEffect(() => {
    api.getConfigPaths()
      .then(data => setConfigPaths(data))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!isActive) return;
    fetchDownloads();
    const interval = setInterval(fetchDownloads, 5000);
    return () => clearInterval(interval);
  }, [isActive, fetchDownloads]);

  return {
    downloads,
    downloadsLoading,
    statusFilter,
    setStatusFilter,
    typeFilter,
    setTypeFilter,
    problemFilter,
    setProblemFilter,
    configPaths,
    fetchDownloads,
  };
}
