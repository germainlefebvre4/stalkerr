import { useState, useCallback } from 'react';
import { api } from '../services/api';
import { M3uSource, M3uSourceInput } from '../types';

// Fetches and mutates the effective M3U source list (origin config.yml
// entries merged with runtime overrides/additions). See the
// m3u-source-overrides spec.
export function useM3uSources() {
  const [sources, setSources] = useState<M3uSource[]>([]);
  const [originNames, setOriginNames] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);

  const fetchSources = useCallback(() => {
    setLoading(true);
    return Promise.all([api.getM3uSources(), api.getM3uSourcesOrigin()])
      .then(([effective, origin]) => {
        setSources(effective.sources || []);
        setOriginNames(new Set((origin.sources || []).map(s => s.name)));
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const createSource = useCallback(async (name: string, input: M3uSourceInput) => {
    await api.createM3uSource(name, input);
    await fetchSources();
  }, [fetchSources]);

  const updateSource = useCallback(async (name: string, input: M3uSourceInput) => {
    await api.updateM3uSource(name, input);
    await fetchSources();
  }, [fetchSources]);

  const deleteSource = useCallback(async (name: string) => {
    await api.deleteM3uSource(name);
    await fetchSources();
  }, [fetchSources]);

  return { sources, originNames, loading, fetchSources, createSource, updateSource, deleteSource };
}
