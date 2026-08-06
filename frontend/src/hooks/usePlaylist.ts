import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '../services/api';
import { PlaylistItem } from '../types';
import { useURLState, URLStateSchema } from './useURLState';

const VALID_LIMITS = [10, 50, 100];
const DEFAULT_LIMIT = 10;
const VALID_TMDB_FILTERS = ['all', 'yes', 'no'];
const VALID_CONTENT_FILTERS = ['all', 'movies', 'tvshows'];
const VALID_STATE_FILTERS = ['all', 'processed', 'pending', 'downloading', 'organizing', 'downloaded', 'failed'];

const PLAYLIST_URL_SCHEMA = {
  page: {
    default: 1,
    parse: (raw: string) => parseInt(raw, 10),
    serialize: String,
    isValid: (v: number) => !isNaN(v) && v > 0,
  },
  limit: {
    default: DEFAULT_LIMIT,
    parse: (raw: string) => parseInt(raw, 10),
    serialize: String,
    isValid: (v: number) => VALID_LIMITS.includes(v),
  },
  search: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
  searchName: { default: '', parse: (raw: string) => raw, serialize: (v: string) => v },
  tmdb: {
    default: 'all' as 'all' | 'yes' | 'no',
    parse: (raw: string) => raw as 'all' | 'yes' | 'no',
    serialize: (v: 'all' | 'yes' | 'no') => v,
    isValid: (v: 'all' | 'yes' | 'no') => VALID_TMDB_FILTERS.includes(v),
  },
  type: {
    default: 'all' as 'all' | 'movies' | 'tvshows',
    parse: (raw: string) => raw as 'all' | 'movies' | 'tvshows',
    serialize: (v: 'all' | 'movies' | 'tvshows') => v,
    isValid: (v: 'all' | 'movies' | 'tvshows') => VALID_CONTENT_FILTERS.includes(v),
  },
  state: {
    default: 'all',
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: (v: string) => VALID_STATE_FILTERS.includes(v),
  },
} satisfies URLStateSchema;

export function usePlaylist() {
  const [urlState, patchURLState] = useURLState(PLAYLIST_URL_SCHEMA);

  const [playlist, setPlaylist] = useState<PlaylistItem[]>([]);
  const [playlistTotal, setPlaylistTotal] = useState(0);
  const [playlistLoading, setPlaylistLoading] = useState(false);

  // If the URL carried no valid `limit`, fall back to the last limit the user
  // picked (persisted in localStorage) instead of the hardcoded default, and
  // reflect that choice back into the URL. Runs once, on mount only.
  const didApplyStoredLimit = useRef(false);
  useEffect(() => {
    if (didApplyStoredLimit.current) return;
    didApplyStoredLimit.current = true;

    const rawLimit = new URLSearchParams(window.location.search).get('limit');
    const hasValidURLLimit = rawLimit !== null && VALID_LIMITS.includes(parseInt(rawLimit, 10));
    if (hasValidURLLimit) return;

    const stored = localStorage.getItem('stalkeer_playlist_limit');
    const parsedStored = stored ? parseInt(stored, 10) : NaN;
    if (!isNaN(parsedStored) && VALID_LIMITS.includes(parsedStored) && parsedStored !== DEFAULT_LIMIT) {
      patchURLState({ limit: parsedStored });
    }
  }, [patchURLState]);

  const setPlaylistSearch = useCallback((value: string) => patchURLState({ search: value }), [patchURLState]);
  const setPlaylistSearchName = useCallback((value: string) => patchURLState({ searchName: value }), [patchURLState]);
  const setPlaylistTMDBFilter = useCallback(
    (value: 'all' | 'yes' | 'no') => patchURLState({ tmdb: value }),
    [patchURLState]
  );
  const setPlaylistFilter = useCallback(
    (value: 'all' | 'movies' | 'tvshows') => patchURLState({ type: value }),
    [patchURLState]
  );
  const setPlaylistStateFilter = useCallback((value: string) => patchURLState({ state: value }), [patchURLState]);

  const setPlaylistPage = useCallback(
    (valueOrUpdater: number | ((prevPage: number) => number)) => {
      const next = typeof valueOrUpdater === 'function' ? valueOrUpdater(urlState.page) : valueOrUpdater;
      patchURLState({ page: next });
    },
    [patchURLState, urlState.page]
  );

  const setPlaylistLimit = useCallback(
    (limit: number) => {
      localStorage.setItem('stalkeer_playlist_limit', limit.toString());
      patchURLState({ limit, page: 1 });
    },
    [patchURLState]
  );

  // Reset to page 1 when any filter actually changes value, but not on the
  // initial mount (which would otherwise clobber a page number restored from
  // the URL). Compares against the previous values rather than a boolean
  // "first render" ref, since React StrictMode double-invokes this effect on
  // mount in development and a boolean ref would trip the reset on the
  // second, synthetic invocation.
  const prevFiltersRef = useRef({
    search: urlState.search,
    searchName: urlState.searchName,
    tmdb: urlState.tmdb,
    type: urlState.type,
    state: urlState.state,
  });
  useEffect(() => {
    const prev = prevFiltersRef.current;
    const changed = prev.search !== urlState.search
      || prev.searchName !== urlState.searchName
      || prev.tmdb !== urlState.tmdb
      || prev.type !== urlState.type
      || prev.state !== urlState.state;
    prevFiltersRef.current = {
      search: urlState.search,
      searchName: urlState.searchName,
      tmdb: urlState.tmdb,
      type: urlState.type,
      state: urlState.state,
    };
    if (changed) {
      patchURLState({ page: 1 });
    }
  }, [urlState.search, urlState.searchName, urlState.tmdb, urlState.type, urlState.state, patchURLState]);

  const fetchPlaylist = useCallback(() => {
    setPlaylistLoading(true);
    api.getPlaylist(
      urlState.page,
      urlState.limit,
      urlState.type,
      urlState.state,
      urlState.search,
      urlState.searchName,
      urlState.tmdb
    )
      .then(data => {
        setPlaylist(data.data || []);
        setPlaylistTotal(data.total || 0);
      })
      .catch(() => {})
      .finally(() => setPlaylistLoading(false));
  }, [urlState.page, urlState.limit, urlState.type, urlState.state, urlState.search, urlState.searchName, urlState.tmdb]);

  useEffect(() => {
    fetchPlaylist();
  }, [fetchPlaylist]);

  return {
    playlist,
    playlistSearch: urlState.search,
    setPlaylistSearch,
    playlistSearchName: urlState.searchName,
    setPlaylistSearchName,
    playlistTMDBFilter: urlState.tmdb,
    setPlaylistTMDBFilter,
    playlistFilter: urlState.type,
    setPlaylistFilter,
    playlistStateFilter: urlState.state,
    setPlaylistStateFilter,
    playlistTotal,
    playlistPage: urlState.page,
    setPlaylistPage,
    playlistLimit: urlState.limit,
    setPlaylistLimit,
    playlistLoading,
    fetchPlaylist,
  };
}
