import { useCallback, useEffect, useRef } from 'react';
import { useURLState, URLStateSchema } from './useURLState';

const VALID_VIEWS = ['items', 'grouped'];
const DEFAULT_VIEW = 'items';
const PLAYLIST_VIEW_STORAGE_KEY = 'stalkeer_playlist_view';

const PLAYLIST_VIEW_URL_SCHEMA = {
  view: {
    default: DEFAULT_VIEW as 'items' | 'grouped',
    parse: (raw: string) => raw as 'items' | 'grouped',
    serialize: (v: 'items' | 'grouped') => v,
    isValid: (v: 'items' | 'grouped') => VALID_VIEWS.includes(v),
  },
} satisfies URLStateSchema;

// Kept as its own useURLState instance (separate from usePlaylist's schema) so
// the sub-tab toggle doesn't interact with the Items view's page-reset-on-filter-
// change effect; independent useURLState instances never clobber each other's params.
export function usePlaylistView() {
  const [urlState, patchURLState] = useURLState(PLAYLIST_VIEW_URL_SCHEMA);

  // If the URL carried no valid `view`, fall back to the last view the user
  // picked (persisted in localStorage) instead of the hardcoded default,
  // mirroring usePlaylist's stored-limit fallback. Runs once, on mount only.
  const didApplyStoredView = useRef(false);
  useEffect(() => {
    if (didApplyStoredView.current) return;
    didApplyStoredView.current = true;

    const rawView = new URLSearchParams(window.location.search).get('view');
    const hasValidURLView = rawView !== null && VALID_VIEWS.includes(rawView);
    if (hasValidURLView) return;

    const stored = localStorage.getItem(PLAYLIST_VIEW_STORAGE_KEY);
    if (stored && VALID_VIEWS.includes(stored) && stored !== DEFAULT_VIEW) {
      patchURLState({ view: stored as 'items' | 'grouped' });
    }
  }, [patchURLState]);

  const setPlaylistView = useCallback(
    (view: 'items' | 'grouped') => {
      localStorage.setItem(PLAYLIST_VIEW_STORAGE_KEY, view);
      patchURLState({ view });
    },
    [patchURLState]
  );

  return {
    playlistView: urlState.view,
    setPlaylistView,
  };
}
