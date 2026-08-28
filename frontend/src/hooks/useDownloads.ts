import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { DownloadEnriched, ConfigPaths } from '../types';
import { useURLState, URLStateSchema } from './useURLState';

const DOWNLOADS_URL_SCHEMA = {
  dlStatus: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
  dlType: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
  dlProblem: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
} satisfies URLStateSchema;

export function useDownloads(isActive: boolean) {
  const [downloads, setDownloads] = useState<DownloadEnriched[]>([]);
  const [downloadsLoading, setDownloadsLoading] = useState(false);
  const [urlState, patchURLState] = useURLState(DOWNLOADS_URL_SCHEMA);
  const [configPaths, setConfigPaths] = useState<ConfigPaths | null>(null);

  const statusFilter = urlState.dlStatus;
  const typeFilter = urlState.dlType;
  const problemFilter = urlState.dlProblem;
  const setStatusFilter = useCallback((value: string) => patchURLState({ dlStatus: value }), [patchURLState]);
  const setTypeFilter = useCallback((value: string) => patchURLState({ dlType: value }), [patchURLState]);
  const setProblemFilter = useCallback((value: string) => patchURLState({ dlProblem: value }), [patchURLState]);

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
    void Promise.resolve().then(fetchDownloads);
    const interval = setInterval(fetchDownloads, 5000);
    return () => clearInterval(interval);
  }, [isActive, fetchDownloads]);

  const updateDownloadPath = useCallback((id: number, newPath: string) => {
    setDownloads(prev => prev.map(item => {
      if (item.id !== id) return item;
      const parts = newPath.split('/');
      const fileName = parts[parts.length - 1] || '';
      const folderName = parts.length > 1 ? parts[parts.length - 2] : '';
      return {
        ...item,
        download_path: newPath,
        file_info: item.file_info
          ? { ...item.file_info, folder_name: folderName, file_name: fileName }
          : item.file_info,
      };
    }));
  }, []);

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
    updateDownloadPath,
  };
}
