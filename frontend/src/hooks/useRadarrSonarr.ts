import { useState, useEffect, useCallback } from 'react';
import { api, ApiError } from '../services/api';
import { RadarrMovieListItem, SonarrSeriesListItem, RadarrSonarrStats, MatchStatusFilter, SuiviFilter, SuiviStatus, FileFilter, FileState } from '../types';
import { useURLState, URLStateSchema } from './useURLState';

const PAGE_SIZE = 20;

const VALID_MATCH_FILTERS: MatchStatusFilter[] = ['', 'matched', 'no_match'];

const VALID_SUIVI_VALUES: SuiviStatus[] = ['monitored', 'unmonitored'];
const VALID_FILE_VALUES: FileState[] = ['downloaded', 'partial', 'none'];
const DEFAULT_SUIVI_STATUS = 'monitored';

function isValidSuiviString(raw: string): boolean {
  return raw.length > 0 && raw.split(',').every(v => (VALID_SUIVI_VALUES as string[]).includes(v));
}

// The file filter may legitimately be empty (no restriction).
function isValidFileString(raw: string): boolean {
  return raw === '' || raw.split(',').every(v => (VALID_FILE_VALUES as string[]).includes(v));
}

// The Suivi and Fichier filters are persisted to the URL as comma-joined
// strings (matching the existing filmsFilter/seriesFilter pattern) and
// converted to/from Sets at the hook boundary.
function parseSuiviFilter(raw: string): SuiviFilter {
  return new Set(raw.split(',').filter((v): v is SuiviStatus => (VALID_SUIVI_VALUES as string[]).includes(v)));
}

function serializeSuiviFilter(filter: SuiviFilter): string {
  return VALID_SUIVI_VALUES.filter(v => filter.has(v)).join(',');
}

function parseFileFilter(raw: string): FileFilter {
  return new Set(raw.split(',').filter((v): v is FileState => (VALID_FILE_VALUES as string[]).includes(v)));
}

function serializeFileFilter(filter: FileFilter): string {
  return VALID_FILE_VALUES.filter(v => filter.has(v)).join(',');
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
    default: DEFAULT_SUIVI_STATUS,
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: isValidSuiviString,
  },
  seriesStatus: {
    default: DEFAULT_SUIVI_STATUS,
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: isValidSuiviString,
  },
  filmsFile: {
    default: '',
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: isValidFileString,
  },
  seriesFile: {
    default: '',
    parse: (raw: string) => raw,
    serialize: (v: string) => v,
    isValid: isValidFileString,
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
  const filmsStatus = parseSuiviFilter(filterURLState.filmsStatus);
  const seriesStatus = parseSuiviFilter(filterURLState.seriesStatus);
  const filmsFile = parseFileFilter(filterURLState.filmsFile);
  const seriesFile = parseFileFilter(filterURLState.seriesFile);

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

  const setFilmsStatus = useCallback((value: SuiviFilter) => {
    patchFilterURLState({ filmsStatus: serializeSuiviFilter(value) });
    setFilmsPage(1);
  }, [patchFilterURLState]);

  const setSeriesStatus = useCallback((value: SuiviFilter) => {
    patchFilterURLState({ seriesStatus: serializeSuiviFilter(value) });
    setSeriesPage(1);
  }, [patchFilterURLState]);

  const setFilmsFile = useCallback((value: FileFilter) => {
    patchFilterURLState({ filmsFile: serializeFileFilter(value) });
    setFilmsPage(1);
  }, [patchFilterURLState]);

  const setSeriesFile = useCallback((value: FileFilter) => {
    patchFilterURLState({ seriesFile: serializeFileFilter(value) });
    setSeriesPage(1);
  }, [patchFilterURLState]);

  // The "Manquants" preset sets both axes in one URL patch so a single fetch fires.
  const setFilmsMissing = useCallback(() => {
    patchFilterURLState({ filmsStatus: 'monitored', filmsFile: 'none' });
    setFilmsPage(1);
  }, [patchFilterURLState]);

  const setSeriesMissing = useCallback(() => {
    patchFilterURLState({ seriesStatus: 'monitored', seriesFile: 'partial,none' });
    setSeriesPage(1);
  }, [patchFilterURLState]);

  const fetchFilms = useCallback(() => {
    setFilmsLoading(true);
    setFilmsError(null);
    api.listRadarrMovies(filmsPage, PAGE_SIZE, filmsSearch, filmsFilter, filmsStatus, filmsFile)
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
  }, [filmsPage, filmsSearch, filmsFilter, filterURLState.filmsStatus, filterURLState.filmsFile]);

  // `refresh` clears the backend's Sonarr match-status cache before recomputing,
  // matching the manual refresh button's contract - never set automatically.
  const fetchSeries = useCallback((refresh?: boolean) => {
    setSeriesLoading(true);
    setSeriesError(null);
    api.listSonarrSeries(seriesPage, PAGE_SIZE, seriesSearch, seriesFilter, refresh, seriesStatus, seriesFile)
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
  }, [seriesPage, seriesSearch, seriesFilter, filterURLState.seriesStatus, filterURLState.seriesFile]);

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
  }, [isActive, filmsPage, filmsSearch, filmsFilter, filterURLState.filmsStatus, filterURLState.filmsFile]);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(() => fetchSeries());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive, seriesPage, seriesSearch, seriesFilter, filterURLState.seriesStatus, filterURLState.seriesFile]);

  useEffect(() => {
    if (!statsActive) return;
    void Promise.resolve().then(fetchStats);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statsActive]);

  return {
    filmsItems, filmsLoading, filmsError, filmsTotal, filmsPage, setFilmsPage, filmsLimit: PAGE_SIZE, fetchFilms,
    filmsSearch, setFilmsSearch, filmsFilter, setFilmsFilter, filmsStatus, setFilmsStatus, filmsFile, setFilmsFile, setFilmsMissing,
    seriesItems, seriesLoading, seriesError, seriesTotal, seriesPage, setSeriesPage, seriesLimit: PAGE_SIZE, fetchSeries, refreshSeries,
    seriesSearch, setSeriesSearch, seriesFilter, setSeriesFilter, seriesStatus, setSeriesStatus, seriesFile, setSeriesFile, setSeriesMissing,
    stats, statsLoading, statsError, fetchStats,
  };
}
