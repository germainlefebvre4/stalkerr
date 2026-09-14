import { useState, useCallback } from 'react';
import { api } from '../services/api';
import { FilterConfig, FilterOriginEntry } from '../types';

export function useFilters() {
  const [filters, setFilters] = useState<FilterConfig[]>([]);
  const [filterOrigin, setFilterOrigin] = useState<FilterOriginEntry[]>([]);
  const [filtersLoading, setFiltersLoading] = useState(false);

  const fetchFilters = useCallback(() => {
    setFiltersLoading(true);
    return Promise.all([api.getFilters(), api.getFilterOrigin()])
      .then(([filtersRes, originRes]) => {
        setFilters(filtersRes.filters || []);
        setFilterOrigin(originRes.origin || []);
      })
      .catch(() => {})
      .finally(() => setFiltersLoading(false));
  }, []);

  const deleteFilter = useCallback(async (id: number) => {
    await api.deleteFilter(id);
    fetchFilters();
  }, [fetchFilters]);

  return {
    filters,
    filterOrigin,
    filtersLoading,
    fetchFilters,
    deleteFilter,
  };
}
