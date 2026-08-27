import { useCallback } from 'react';
import { useURLState, URLStateSchema } from './useURLState';

const VALID_VIEWS = ['items', 'grouped'];

const PLAYLIST_VIEW_URL_SCHEMA = {
  view: {
    default: 'items' as 'items' | 'grouped',
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

  const setPlaylistView = useCallback(
    (view: 'items' | 'grouped') => patchURLState({ view }),
    [patchURLState]
  );

  return {
    playlistView: urlState.view,
    setPlaylistView,
  };
}
