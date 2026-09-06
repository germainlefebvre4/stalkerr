import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '../services/api';
import { DownloadEnriched, ConfigPaths } from '../types';
import { useURLState, URLStateSchema } from './useURLState';

const VALID_LIMITS = [20, 50, 100];
const DEFAULT_LIMIT = 20;

const DOWNLOADS_URL_SCHEMA = {
  dlStatus: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
  dlType: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
  dlProblem: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
  dlPage: {
    default: 1,
    parse: (raw: string) => parseInt(raw, 10),
    serialize: String,
    isValid: (v: number) => !isNaN(v) && v > 0,
  },
  dlLimit: {
    default: DEFAULT_LIMIT,
    parse: (raw: string) => parseInt(raw, 10),
    serialize: String,
    isValid: (v: number) => VALID_LIMITS.includes(v),
  },
} satisfies URLStateSchema;

export function useDownloads(isActive: boolean) {
  const [downloads, setDownloads] = useState<DownloadEnriched[]>([]);
  const [downloadsLoading, setDownloadsLoading] = useState(false);
  const [downloadsTotal, setDownloadsTotal] = useState(0);
  const [urlState, patchURLState] = useURLState(DOWNLOADS_URL_SCHEMA);
  const [configPaths, setConfigPaths] = useState<ConfigPaths | null>(null);

  const statusFilter = urlState.dlStatus;
  const typeFilter = urlState.dlType;
  const problemFilter = urlState.dlProblem;
  const downloadsPage = urlState.dlPage;
  const downloadsLimit = urlState.dlLimit;

  const setStatusFilter = useCallback((value: string) => patchURLState({ dlStatus: value }), [patchURLState]);
  const setTypeFilter = useCallback((value: string) => patchURLState({ dlType: value }), [patchURLState]);
  const setProblemFilter = useCallback((value: string) => patchURLState({ dlProblem: value }), [patchURLState]);

  const setDownloadsPage = useCallback(
    (valueOrUpdater: number | ((prevPage: number) => number)) => {
      const next = typeof valueOrUpdater === 'function' ? valueOrUpdater(urlState.dlPage) : valueOrUpdater;
      patchURLState({ dlPage: next });
    },
    [patchURLState, urlState.dlPage]
  );

  const setDownloadsLimit = useCallback(
    (limit: number) => patchURLState({ dlLimit: limit, dlPage: 1 }),
    [patchURLState]
  );

  const fetchDownloads = useCallback(() => {
    setDownloadsLoading(true);
    const offset = (downloadsPage - 1) * downloadsLimit;
    api.getDownloads(downloadsLimit, statusFilter, typeFilter, problemFilter, offset)
      .then(data => {
        setDownloads(data.data || []);
        setDownloadsTotal(data.total || 0);
      })
      .catch(() => {})
      .finally(() => setDownloadsLoading(false));
  }, [statusFilter, typeFilter, problemFilter, downloadsPage, downloadsLimit]);

  useEffect(() => {
    api.getConfigPaths()
      .then(data => setConfigPaths(data))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchDownloads);
  }, [isActive, fetchDownloads]);

  // Reset to page 1 when a filter actually changes value, but not on the
  // initial mount (which would otherwise clobber a page number restored from
  // the URL). Mirrors usePlaylist's equivalent effect.
  const prevFiltersRef = useRef({ status: statusFilter, type: typeFilter, problem: problemFilter });
  useEffect(() => {
    const prev = prevFiltersRef.current;
    const changed = prev.status !== statusFilter || prev.type !== typeFilter || prev.problem !== problemFilter;
    prevFiltersRef.current = { status: statusFilter, type: typeFilter, problem: problemFilter };
    if (changed) {
      patchURLState({ dlPage: 1 });
    }
  }, [statusFilter, typeFilter, problemFilter, patchURLState]);

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

  const updateDownloadStatus = useCallback((id: number, newStatus: DownloadEnriched['status']) => {
    setDownloads(prev => prev.map(item => (item.id === id ? { ...item, status: newStatus } : item)));
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
    downloadsPage,
    setDownloadsPage,
    downloadsLimit,
    setDownloadsLimit,
    downloadsTotal,
    configPaths,
    fetchDownloads,
    updateDownloadPath,
    updateDownloadStatus,
  };
}
