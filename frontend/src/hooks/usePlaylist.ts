import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { PlaylistItem } from '../types';

export function usePlaylist() {
  const [playlist, setPlaylist] = useState<PlaylistItem[]>([]);
  const [playlistSearch, setPlaylistSearch] = useState('');
  const [playlistSearchName, setPlaylistSearchName] = useState('');
  const [playlistTMDBFilter, setPlaylistTMDBFilter] = useState<'all' | 'yes' | 'no'>('all');
  const [playlistFilter, setPlaylistFilter] = useState<'all' | 'movies' | 'tvshows'>('all');
  const [playlistStateFilter, setPlaylistStateFilter] = useState<string>('all');
  const [playlistTotal, setPlaylistTotal] = useState(0);
  const [playlistPage, setPlaylistPage] = useState(1);
  const [playlistLoading, setPlaylistLoading] = useState(false);
  const [playlistLimit, setPlaylistLimitState] = useState<number>(() => {
    const stored = localStorage.getItem('stalkeer_playlist_limit');
    if (stored) {
      const parsed = parseInt(stored, 10);
      if (!isNaN(parsed) && [10, 50, 100].includes(parsed)) {
        return parsed;
      }
    }
    return 50;
  });

  const setPlaylistLimit = useCallback((limit: number) => {
    setPlaylistLimitState(limit);
    localStorage.setItem('stalkeer_playlist_limit', limit.toString());
    setPlaylistPage(1);
  }, []);

  // Reset to page 1 when any filter is modified
  useEffect(() => {
    setPlaylistPage(1);
  }, [playlistSearch, playlistSearchName, playlistTMDBFilter, playlistFilter, playlistStateFilter]);

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
