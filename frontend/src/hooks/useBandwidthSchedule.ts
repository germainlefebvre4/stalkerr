import { useState, useCallback } from 'react';
import { api } from '../services/api';
import { EffectivePolicy, ScheduleWindow, ScheduleWindowInput } from '../types';

// Fetches and mutates the weekly bandwidth schedule windows, and the
// currently effective download policy preview. See the
// download-bandwidth-schedule and adaptive-download-throttling specs.
export function useBandwidthSchedule() {
  const [windows, setWindows] = useState<ScheduleWindow[]>([]);
  const [effectivePolicy, setEffectivePolicy] = useState<EffectivePolicy | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchWindows = useCallback(() => {
    setLoading(true);
    return api.getScheduleWindows()
      .then(result => setWindows(result.windows || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const fetchEffectivePolicy = useCallback(() => {
    return api.getEffectivePolicy()
      .then(setEffectivePolicy)
      .catch(() => {});
  }, []);

  const createWindow = useCallback(async (input: ScheduleWindowInput) => {
    await api.createScheduleWindow(input);
    await fetchWindows();
  }, [fetchWindows]);

  const updateWindow = useCallback(async (id: number, input: ScheduleWindowInput) => {
    await api.updateScheduleWindow(id, input);
    await fetchWindows();
  }, [fetchWindows]);

  const deleteWindow = useCallback(async (id: number) => {
    await api.deleteScheduleWindow(id);
    await fetchWindows();
  }, [fetchWindows]);

  return {
    windows, loading, fetchWindows, createWindow, updateWindow, deleteWindow,
    effectivePolicy, fetchEffectivePolicy,
  };
}
