import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '../services/api';
import { DownloadEnriched, ConfigPaths } from '../types';
import { useURLState, URLStateSchema } from './useURLState';

const VALID_LIMITS = [20, 50, 100];
const DEFAULT_LIMIT = 20;
const SEARCH_DEBOUNCE_MS = 300;

// Status filter value meaning "no status constraint" (Tous). Serialized in the
// URL so a refresh doesn't fall back to the "active" default.
export const STATUS_ALL = 'all';
export const STATUS_ACTIVE = 'active';

// Playlist-style `searchName` links are honored once as an alias of `dlSearch`.
function readSearchAlias(): string {
  try {
    const params = new URLSearchParams(window.location.search);
    return params.get('dlSearch') === null ? params.get('searchName') ?? '' : '';
  } catch {
    return '';
  }
}

const DOWNLOADS_URL_SCHEMA = {
  dlStatus: { default: STATUS_ACTIVE, parse: (raw: string) => raw, serialize: (v: string) => v },
  dlSearch: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
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
  const [cancelledHidden, setCancelledHidden] = useState(0);
  const [configPaths, setConfigPaths] = useState<ConfigPaths | null>(null);

  const statusFilter = urlState.dlStatus;
  const typeFilter = urlState.dlType;
  const problemFilter = urlState.dlProblem;
  const searchFilter = urlState.dlSearch;
  const downloadsPage = urlState.dlPage;
  const downloadsLimit = urlState.dlLimit;

  // The input value updates immediately; the URL/search filter follows after a debounce.
  const [searchInput, setSearchInput] = useState(() => urlState.dlSearch || readSearchAlias());

  // Promote a `searchName` alias to `dlSearch` once, so the first fetch is filtered.
  useEffect(() => {
    const alias = readSearchAlias();
    if (alias) patchURLState({ dlSearch: alias });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (searchInput === searchFilter) return;
    const timer = setTimeout(() => patchURLState({ dlSearch: searchInput, dlPage: 1 }), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchInput, searchFilter, patchURLState]);

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
    const apiStatus = statusFilter === STATUS_ALL ? '' : statusFilter;
    api.getDownloads(downloadsLimit, apiStatus, typeFilter, problemFilter, offset, searchFilter)
      .then(data => {
        setDownloads(data.data || []);
        setDownloadsTotal(data.total || 0);
        setCancelledHidden(data.cancelled_hidden || 0);
      })
      .catch(() => {})
      .finally(() => setDownloadsLoading(false));
  }, [statusFilter, typeFilter, problemFilter, searchFilter, downloadsPage, downloadsLimit]);

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
  const prevFiltersRef = useRef({ status: statusFilter, type: typeFilter, problem: problemFilter, search: searchFilter });
  useEffect(() => {
    const prev = prevFiltersRef.current;
    const changed = prev.status !== statusFilter || prev.type !== typeFilter || prev.problem !== problemFilter || prev.search !== searchFilter;
    prevFiltersRef.current = { status: statusFilter, type: typeFilter, problem: problemFilter, search: searchFilter };
    if (changed) {
      patchURLState({ dlPage: 1 });
    }
  }, [statusFilter, typeFilter, problemFilter, searchFilter, patchURLState]);

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
    searchInput,
    setSearchInput,
    cancelledHidden,
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
