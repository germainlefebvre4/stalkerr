import { useState, useEffect, useCallback } from 'react';
import { api, ApiError } from '../services/api';
import { RadarrMovieListItem, SonarrSeriesListItem, RadarrSonarrStats, MatchStatusFilter, EtatFilter, EtatStatus } from '../types';
import { useURLState, URLStateSchema } from './useURLState';

const PAGE_SIZE = 20;

const VALID_MATCH_FILTERS: MatchStatusFilter[] = ['', 'matched', 'no_match'];

const VALID_ETAT_VALUES: EtatStatus[] = ['monitored', 'unmonitored', 'missing'];
const DEFAULT_ETAT_STATUS = 'monitored';

function isValidEtatStatusString(raw: string): boolean {
  return raw.length > 0 && raw.split(',').every(v => (VALID_ETAT_VALUES as string[]).includes(v));
}

// The État filter is persisted to the URL as a comma-joined string (matching
// the existing filmsFilter/seriesFilter pattern) and converted to/from a
// Set<EtatStatus> at the hook boundary.
function parseEtatFilter(raw: string): EtatFilter {
  return new Set(raw.split(',').filter((v): v is EtatStatus => (VALID_ETAT_VALUES as string[]).includes(v)));
}

function serializeEtatFilter(filter: EtatFilter): string {
  return VALID_ETAT_VALUES.filter(v => filter.has(v)).join(',');
}

// Kept as its own useURLState instance so the match-status filters don't
// interact with the sub-tab toggle's URL state (see useRadarrSonarrView).
const RADARR_SONARR_FILTER_URL_SCHEMA = {
  filmsFilter: {
    default: '' as MatchStatusFilter,
    parse: (raw: string) => raw as MatchStatusFilter,
    serialize: (v: MatchStatusFilter) => v,
    isValid: (v: MatchStatusFilter) => VALID_MATCH_FILTERS.includes(v),
  },
  seriesFilter: {
    default: '' as MatchStatusFilter,
    parse: (raw: string) => raw as MatchStatusFilter,
    serialize: (v: MatchStatusFilter) => v,
    isValid: (v: MatchStatusFilter) => VALID_MATCH_FILTERS.includes(v),
  },
  filmsStatus: {
    default: DEFAULT_ETAT_STATUS,
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: isValidEtatStatusString,
  },
  seriesStatus: {
    default: DEFAULT_ETAT_STATUS,
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: isValidEtatStatusString,
  },
} satisfies URLStateSchema;

// Each section (Films/Séries) owns fully independent loading/error/pagination
// state and only ever fetches on an explicit trigger: the initial activation of
// the tab, a pagination change, a search term change, or the section's own
// refresh button - never on a timer, and never because the other section's
// request failed or succeeded. The Résumé stats fetch follows the same
// activation-triggered discipline, but is also requested on Home-tab
// activation (`statsActive`, distinct from `isActive`) since the Home tab
// shows the same Radarr/Sonarr summary without mounting the Films/Séries lists.
export function useRadarrSonarr(isActive: boolean, statsActive: boolean = isActive) {
  const [filmsItems, setFilmsItems] = useState<RadarrMovieListItem[]>([]);
  const [filmsLoading, setFilmsLoading] = useState(false);
  const [filmsError, setFilmsError] = useState<string | null>(null);
  const [filmsTotal, setFilmsTotal] = useState(0);
  const [filmsPage, setFilmsPage] = useState(1);
  const [filmsSearch, setFilmsSearchState] = useState('');

  const [seriesItems, setSeriesItems] = useState<SonarrSeriesListItem[]>([]);
  const [seriesLoading, setSeriesLoading] = useState(false);
  const [seriesError, setSeriesError] = useState<string | null>(null);
  const [seriesTotal, setSeriesTotal] = useState(0);
  const [seriesPage, setSeriesPage] = useState(1);
  const [seriesSearch, setSeriesSearchState] = useState('');

  const [filterURLState, patchFilterURLState] = useURLState(RADARR_SONARR_FILTER_URL_SCHEMA);
  const filmsFilter = filterURLState.filmsFilter;
  const seriesFilter = filterURLState.seriesFilter;
  const filmsStatus = parseEtatFilter(filterURLState.filmsStatus);
  const seriesStatus = parseEtatFilter(filterURLState.seriesStatus);

  const [stats, setStats] = useState<RadarrSonarrStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);
  const [statsError, setStatsError] = useState<string | null>(null);

  const setFilmsSearch = useCallback((value: string) => {
    setFilmsSearchState(value);
    setFilmsPage(1);
  }, []);

  const setSeriesSearch = useCallback((value: string) => {
    setSeriesSearchState(value);
    setSeriesPage(1);
  }, []);

  const setFilmsFilter = useCallback((value: MatchStatusFilter) => {
    patchFilterURLState({ filmsFilter: value });
    setFilmsPage(1);
  }, [patchFilterURLState]);

  const setSeriesFilter = useCallback((value: MatchStatusFilter) => {
    patchFilterURLState({ seriesFilter: value });
    setSeriesPage(1);
  }, [patchFilterURLState]);

  const setFilmsStatus = useCallback((value: EtatFilter) => {
    patchFilterURLState({ filmsStatus: serializeEtatFilter(value) });
    setFilmsPage(1);
  }, [patchFilterURLState]);

  const setSeriesStatus = useCallback((value: EtatFilter) => {
    patchFilterURLState({ seriesStatus: serializeEtatFilter(value) });
    setSeriesPage(1);
  }, [patchFilterURLState]);

  const fetchFilms = useCallback(() => {
    setFilmsLoading(true);
    setFilmsError(null);
    api.listRadarrMovies(filmsPage, PAGE_SIZE, filmsSearch, filmsFilter, filmsStatus)
      .then(data => {
        setFilmsItems(data.data || []);
        setFilmsTotal(data.total);
      })
      .catch(err => {
        setFilmsItems([]);
        setFilmsTotal(0);
        setFilmsError(err instanceof ApiError ? err.code : 'generic');
      })
      .finally(() => setFilmsLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filmsPage, filmsSearch, filmsFilter, filterURLState.filmsStatus]);

  // `refresh` clears the backend's Sonarr match-status cache before recomputing,
  // matching the manual refresh button's contract - never set automatically.
  const fetchSeries = useCallback((refresh?: boolean) => {
    setSeriesLoading(true);
    setSeriesError(null);
    api.listSonarrSeries(seriesPage, PAGE_SIZE, seriesSearch, seriesFilter, refresh, seriesStatus)
      .then(data => {
        setSeriesItems(data.data || []);
        setSeriesTotal(data.total);
      })
      .catch(err => {
        setSeriesItems([]);
        setSeriesTotal(0);
        setSeriesError(err instanceof ApiError ? err.code : 'generic');
      })
      .finally(() => setSeriesLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [seriesPage, seriesSearch, seriesFilter, filterURLState.seriesStatus]);

  const fetchStats = useCallback(() => {
    setStatsLoading(true);
    setStatsError(null);
    api.getRadarrSonarrStats()
      .then(data => setStats(data))
      .catch(err => {
        setStats(null);
        setStatsError(err instanceof ApiError ? err.code : 'generic');
      })
      .finally(() => setStatsLoading(false));
  }, []);

  // refreshSeries is the Séries section's manual refresh action: unlike the
  // automatic fetch effect below, it clears the backend's Sonarr match-status
  // cache so a subsequent filtered request recomputes rather than reusing a
  // stale cached value.
  const refreshSeries = useCallback(() => fetchSeries(true), [fetchSeries]);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchFilms);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive, filmsPage, filmsSearch, filmsFilter, filterURLState.filmsStatus]);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(() => fetchSeries());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive, seriesPage, seriesSearch, seriesFilter, filterURLState.seriesStatus]);

  useEffect(() => {
    if (!statsActive) return;
    void Promise.resolve().then(fetchStats);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statsActive]);

  return {
    filmsItems, filmsLoading, filmsError, filmsTotal, filmsPage, setFilmsPage, filmsLimit: PAGE_SIZE, fetchFilms,
    filmsSearch, setFilmsSearch, filmsFilter, setFilmsFilter, filmsStatus, setFilmsStatus,
    seriesItems, seriesLoading, seriesError, seriesTotal, seriesPage, setSeriesPage, seriesLimit: PAGE_SIZE, fetchSeries, refreshSeries,
    seriesSearch, setSeriesSearch, seriesFilter, setSeriesFilter, seriesStatus, setSeriesStatus,
    stats, statsLoading, statsError, fetchStats,
  };
}
