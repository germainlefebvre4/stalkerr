import { useCallback } from 'react';
import { useURLState, URLStateSchema } from './useURLState';

const VALID_SUBTABS = ['resume', 'radarr', 'sonarr'];

const RADARR_SONARR_VIEW_URL_SCHEMA = {
  subtab: {
    default: 'radarr' as 'resume' | 'radarr' | 'sonarr',
    parse: (raw: string) => raw as 'resume' | 'radarr' | 'sonarr',
    serialize: (v: 'resume' | 'radarr' | 'sonarr') => v,
    isValid: (v: 'resume' | 'radarr' | 'sonarr') => VALID_SUBTABS.includes(v),
  },
} satisfies URLStateSchema;

// Kept as its own useURLState instance (separate from the tab's data-fetching
// hooks) so the sub-tab toggle doesn't interact with their state; independent
// useURLState instances never clobber each other's params.
export function useRadarrSonarrView() {
  const [urlState, patchURLState] = useURLState(RADARR_SONARR_VIEW_URL_SCHEMA);

  const setActiveSubTab = useCallback(
    (subtab: 'resume' | 'radarr' | 'sonarr') => patchURLState({ subtab }),
    [patchURLState]
  );

  return {
    activeSubTab: urlState.subtab,
    setActiveSubTab,
  };
}
