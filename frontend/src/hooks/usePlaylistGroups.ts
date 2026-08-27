import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '../services/api';
import { MediaGroupItem } from '../types';

const GROUPS_LIMIT = 10;

/**
 * Fetches the grouped ("Films & Séries") playlist view, sharing the same
 * content-type/state/tmdb/search filters as the Items view (usePlaylist) but
 * owning its own pagination, since a group's position in the list is unrelated
 * to the Items view's page.
 */
export function usePlaylistGroups(
  active: boolean,
  contentType: 'all' | 'movies' | 'tvshows',
  stateFilter: string,
  search: string,
  searchName: string,
  tmdbFilter: 'all' | 'yes' | 'no'
) {
  const [groups, setGroups] = useState<MediaGroupItem[]>([]);
  const [groupsTotal, setGroupsTotal] = useState(0);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsPage, setGroupsPage] = useState(1);

  // Reset to page 1 when a filter changes, but not on the initial mount.
  const didMount = useRef(false);
  useEffect(() => {
    if (!didMount.current) {
      didMount.current = true;
      return;
    }
    setGroupsPage(1);
  }, [contentType, stateFilter, search, searchName, tmdbFilter]);

  const fetchGroups = useCallback(() => {
    if (!active) return;
    setGroupsLoading(true);
    api.getGroupedPlaylist(groupsPage, GROUPS_LIMIT, contentType, stateFilter, search, searchName, tmdbFilter)
      .then(data => {
        setGroups(data.data || []);
        setGroupsTotal(data.total || 0);
      })
      .catch(() => {})
      .finally(() => setGroupsLoading(false));
  }, [active, groupsPage, contentType, stateFilter, search, searchName, tmdbFilter]);

  useEffect(() => {
    void Promise.resolve().then(fetchGroups);
  }, [fetchGroups]);

  return {
    groups,
    groupsTotal,
    groupsLoading,
    groupsPage,
    setGroupsPage,
    groupsLimit: GROUPS_LIMIT,
    fetchGroups,
  };
}
