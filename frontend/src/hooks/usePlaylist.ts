import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '../services/api';
import { PlaylistItem } from '../types';

const VALID_LIMITS = [10, 50, 100];
const DEFAULT_LIMIT = 50;
const VALID_TMDB_FILTERS = ['all', 'yes', 'no'];
const VALID_CONTENT_FILTERS = ['all', 'movies', 'tvshows'];
const VALID_STATE_FILTERS = ['all', 'processed', 'pending', 'downloading', 'organizing', 'downloaded', 'failed'];

function readInitialStateFromURL() {
  const params = new URLSearchParams(window.location.search);

  const page = parseInt(params.get('page') || '', 10);
  const limit = parseInt(params.get('limit') || '', 10);
  const tmdb = params.get('tmdb');
  const type = params.get('type');
  const state = params.get('state');

  return {
    page: !isNaN(page) && page > 0 ? page : 1,
    limit: !isNaN(limit) && VALID_LIMITS.includes(limit) ? limit : null,
    search: params.get('search') || '',
    searchName: params.get('searchName') || '',
    tmdb: (tmdb && VALID_TMDB_FILTERS.includes(tmdb) ? tmdb : 'all') as 'all' | 'yes' | 'no',
    type: (type && VALID_CONTENT_FILTERS.includes(type) ? type : 'all') as 'all' | 'movies' | 'tvshows',
    state: state && VALID_STATE_FILTERS.includes(state) ? state : 'all',
  };
}

export function usePlaylist() {
  const initialURLState = useRef(readInitialStateFromURL()).current;

  const [playlist, setPlaylist] = useState<PlaylistItem[]>([]);
  const [playlistSearch, setPlaylistSearch] = useState(initialURLState.search);
  const [playlistSearchName, setPlaylistSearchName] = useState(initialURLState.searchName);
  const [playlistTMDBFilter, setPlaylistTMDBFilter] = useState<'all' | 'yes' | 'no'>(initialURLState.tmdb);
  const [playlistFilter, setPlaylistFilter] = useState<'all' | 'movies' | 'tvshows'>(initialURLState.type);
  const [playlistStateFilter, setPlaylistStateFilter] = useState<string>(initialURLState.state);
  const [playlistTotal, setPlaylistTotal] = useState(0);
  const [playlistPage, setPlaylistPage] = useState(initialURLState.page);
  const [playlistLoading, setPlaylistLoading] = useState(false);
  const [playlistLimit, setPlaylistLimitState] = useState<number>(() => {
    if (initialURLState.limit) {
      return initialURLState.limit;
    }
    const stored = localStorage.getItem('stalkeer_playlist_limit');
    if (stored) {
      const parsed = parseInt(stored, 10);
      if (!isNaN(parsed) && VALID_LIMITS.includes(parsed)) {
        return parsed;
      }
    }
    return DEFAULT_LIMIT;
  });

  const setPlaylistLimit = useCallback((limit: number) => {
    setPlaylistLimitState(limit);
    localStorage.setItem('stalkeer_playlist_limit', limit.toString());
    setPlaylistPage(1);
  }, []);

  // Reset to page 1 when any filter actually changes value, but not on the
  // initial mount (which would otherwise clobber a page number restored from
  // the URL). Compares against the previous values rather than a boolean
  // "first render" ref, since React StrictMode double-invokes this effect on
  // mount in development and a boolean ref would trip the reset on the
  // second, synthetic invocation.
  const prevFiltersRef = useRef({
    search: initialURLState.search,
    searchName: initialURLState.searchName,
    tmdb: initialURLState.tmdb,
    type: initialURLState.type,
    state: initialURLState.state,
  });
  useEffect(() => {
    const prev = prevFiltersRef.current;
    const changed = prev.search !== playlistSearch
      || prev.searchName !== playlistSearchName
      || prev.tmdb !== playlistTMDBFilter
      || prev.type !== playlistFilter
      || prev.state !== playlistStateFilter;
    prevFiltersRef.current = {
      search: playlistSearch,
      searchName: playlistSearchName,
      tmdb: playlistTMDBFilter,
      type: playlistFilter,
      state: playlistStateFilter,
    };
    if (changed) {
      setPlaylistPage(1);
    }
  }, [playlistSearch, playlistSearchName, playlistTMDBFilter, playlistFilter, playlistStateFilter]);

  // Keep the URL query string in sync with the current page, limit, and filters
  // so a refresh restores the exact same view. Uses replaceState (not pushState)
  // to avoid flooding browser history with every keystroke/page change.
  useEffect(() => {
    const params = new URLSearchParams();
    if (playlistPage !== 1) params.set('page', String(playlistPage));
    if (playlistLimit !== DEFAULT_LIMIT) params.set('limit', String(playlistLimit));
    if (playlistSearch) params.set('search', playlistSearch);
    if (playlistSearchName) params.set('searchName', playlistSearchName);
    if (playlistTMDBFilter !== 'all') params.set('tmdb', playlistTMDBFilter);
    if (playlistFilter !== 'all') params.set('type', playlistFilter);
    if (playlistStateFilter !== 'all') params.set('state', playlistStateFilter);

    const query = params.toString();
    const newUrl = `${window.location.pathname}${query ? `?${query}` : ''}${window.location.hash}`;
    window.history.replaceState(null, '', newUrl);
  }, [playlistPage, playlistLimit, playlistSearch, playlistSearchName, playlistTMDBFilter, playlistFilter, playlistStateFilter]);

  const fetchPlaylist = useCallback(() => {
    setPlaylistLoading(true);
    api.getPlaylist(
      playlistPage,
      playlistLimit,
      playlistFilter,
      playlistStateFilter,
      playlistSearch,
      playlistSearchName,
      playlistTMDBFilter
    )
      .then(data => {
        setPlaylist(data.data || []);
        setPlaylistTotal(data.total || 0);
      })
      .catch(() => {})
      .finally(() => setPlaylistLoading(false));
  }, [playlistPage, playlistLimit, playlistFilter, playlistStateFilter, playlistSearch, playlistSearchName, playlistTMDBFilter]);

  useEffect(() => {
    fetchPlaylist();
  }, [fetchPlaylist]);

  return {
    playlist,
    playlistSearch,
    setPlaylistSearch,
    playlistSearchName,
    setPlaylistSearchName,
    playlistTMDBFilter,
    setPlaylistTMDBFilter,
    playlistFilter,
    setPlaylistFilter,
    playlistStateFilter,
    setPlaylistStateFilter,
    playlistTotal,
    playlistPage,
    setPlaylistPage,
    playlistLimit,
    setPlaylistLimit,
    playlistLoading,
    fetchPlaylist,
  };
}
