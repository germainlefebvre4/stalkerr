import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { PlaylistItem } from '../types';

export function usePlaylist() {
  const [playlist, setPlaylist] = useState<PlaylistItem[]>([]);
  const [playlistSearch, setPlaylistSearch] = useState('');
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

  const fetchPlaylist = useCallback(() => {
    setPlaylistLoading(true);
    api.getPlaylist(
      playlistPage,
      playlistLimit,
      playlistFilter,
      playlistStateFilter,
      playlistSearch
    )
      .then(data => {
        setPlaylist(data.data || []);
        setPlaylistTotal(data.total || 0);
      })
      .catch(() => {})
      .finally(() => setPlaylistLoading(false));
  }, [playlistPage, playlistLimit, playlistFilter, playlistStateFilter, playlistSearch]);

  useEffect(() => {
    fetchPlaylist();
  }, [fetchPlaylist]);

  return {
    playlist,
    playlistSearch,
    setPlaylistSearch,
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
