import { useState, useCallback } from 'react';
import { api } from '../services/api';
import { FilterConfig, SystemFilterConfig } from '../types';

export function useFilters() {
  const [filters, setFilters] = useState<FilterConfig[]>([]);
  const [filtersLoading, setFiltersLoading] = useState(false);
  const [systemFilters, setSystemFilters] = useState<SystemFilterConfig | null>(null);
  const [systemFiltersLoading, setSystemFiltersLoading] = useState(false);

  const fetchFilters = useCallback(() => {
    setFiltersLoading(true);
    api.getFilters()
      .then(data => {
        setFilters(data.filters || []);
      })
      .catch(() => {})
      .finally(() => setFiltersLoading(false));
  }, []);

  const fetchSystemFilters = useCallback(() => {
    setSystemFiltersLoading(true);
    api.getSystemFilters()
      .then(data => {
        setSystemFilters(data);
      })
      .catch(() => {})
      .finally(() => setSystemFiltersLoading(false));
  }, []);

  const deleteFilter = useCallback(async (id: number) => {
    await api.deleteFilter(id);
    fetchFilters();
  }, [fetchFilters]);

  return {
    filters,
    filtersLoading,
    fetchFilters,
    deleteFilter,
    systemFilters,
    systemFiltersLoading,
    fetchSystemFilters,
  };
}
