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

  const fetchPlaylist = useCallback(() => {
    setPlaylistLoading(true);
    api.getPlaylist(
      playlistPage,
      15,
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
  }, [playlistPage, playlistFilter, playlistStateFilter, playlistSearch]);

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
    playlistLoading,
    fetchPlaylist,
  };
}
