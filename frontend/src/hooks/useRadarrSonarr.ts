import { useState, useEffect, useCallback } from 'react';
import { api, ApiError } from '../services/api';
import { RadarrMovieListItem, SonarrSeriesListItem } from '../types';

const PAGE_SIZE = 20;

// Each section (Films/Séries) owns fully independent loading/error/pagination
// state and only ever fetches on an explicit trigger: the initial activation of
// the tab, a pagination change, or the section's own refresh button - never on a
// timer, and never because the other section's request failed or succeeded.
export function useRadarrSonarr(isActive: boolean) {
  const [filmsItems, setFilmsItems] = useState<RadarrMovieListItem[]>([]);
  const [filmsLoading, setFilmsLoading] = useState(false);
  const [filmsError, setFilmsError] = useState<string | null>(null);
  const [filmsTotal, setFilmsTotal] = useState(0);
  const [filmsPage, setFilmsPage] = useState(1);

  const [seriesItems, setSeriesItems] = useState<SonarrSeriesListItem[]>([]);
  const [seriesLoading, setSeriesLoading] = useState(false);
  const [seriesError, setSeriesError] = useState<string | null>(null);
  const [seriesTotal, setSeriesTotal] = useState(0);
  const [seriesPage, setSeriesPage] = useState(1);

  const fetchFilms = useCallback(() => {
    setFilmsLoading(true);
    setFilmsError(null);
    api.listRadarrMovies(filmsPage, PAGE_SIZE)
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
  }, [filmsPage]);

  const fetchSeries = useCallback(() => {
    setSeriesLoading(true);
    setSeriesError(null);
    api.listSonarrSeries(seriesPage, PAGE_SIZE)
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
  }, [seriesPage]);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchFilms);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive, filmsPage]);

  useEffect(() => {
    if (!isActive) return;
    void Promise.resolve().then(fetchSeries);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive, seriesPage]);

  return {
    filmsItems, filmsLoading, filmsError, filmsTotal, filmsPage, setFilmsPage, filmsLimit: PAGE_SIZE, fetchFilms,
    seriesItems, seriesLoading, seriesError, seriesTotal, seriesPage, setSeriesPage, seriesLimit: PAGE_SIZE, fetchSeries,
  };
}
