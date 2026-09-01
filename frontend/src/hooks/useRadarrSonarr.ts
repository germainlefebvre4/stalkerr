import { useState, useEffect, useCallback } from 'react';
import { api, ApiError } from '../services/api';
import { RadarrMovieListItem, SonarrSeriesListItem, RadarrSonarrStats } from '../types';

const PAGE_SIZE = 20;

// Each section (Films/Séries) owns fully independent loading/error/pagination
// state and only ever fetches on an explicit trigger: the initial activation of
// the tab, a pagination change, a search term change, or the section's own
// refresh button - never on a timer, and never because the other section's
// request failed or succeeded. The Résumé stats fetch follows the same
// activation-triggered discipline.
export function useRadarrSonarr(isActive: boolean) {
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

  const fetchFilms = useCallback(() => {
    setFilmsLoading(true);
    setFilmsError(null);
    api.listRadarrMovies(filmsPage, PAGE_SIZE, filmsSearch)
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
  }, [filmsPage, filmsSearch]);

  const fetchSeries = useCallback(() => {
    setSeriesLoading(true);
    setSeriesError(null);
    api.listSonarrSeries(seriesPage, PAGE_SIZE, seriesSearch)
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
  }, [seriesPage, seriesSearch]);

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

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchFilms);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive, filmsPage, filmsSearch]);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchSeries);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive, seriesPage, seriesSearch]);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchStats);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive]);

  return {
    filmsItems, filmsLoading, filmsError, filmsTotal, filmsPage, setFilmsPage, filmsLimit: PAGE_SIZE, fetchFilms,
    filmsSearch, setFilmsSearch,
    seriesItems, seriesLoading, seriesError, seriesTotal, seriesPage, setSeriesPage, seriesLimit: PAGE_SIZE, fetchSeries,
    seriesSearch, setSeriesSearch,
    stats, statsLoading, statsError, fetchStats,
  };
}
