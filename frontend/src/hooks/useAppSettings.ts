import { useState, useCallback } from 'react';
import { api } from '../services/api';
import { SettingsField, BootstrapField } from '../types';

// Fetches and mutates the app-settings-backed fields (Radarr, Sonarr, TMDB,
// Jellyfin, Notifications, Downloads, logging) plus the read-only bootstrap
// configuration. See the app-settings spec.
export function useAppSettings() {
  const [settings, setSettings] = useState<SettingsField[]>([]);
  const [bootstrap, setBootstrap] = useState<BootstrapField[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchSettings = useCallback(() => {
    setLoading(true);
    return Promise.all([api.getSettings(), api.getBootstrapSettings()])
      .then(([settingsRes, bootstrapRes]) => {
        setSettings(settingsRes.settings || []);
        setBootstrap(bootstrapRes.bootstrap || []);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const setSetting = useCallback(async (key: string, value: unknown) => {
    const updated = await api.setSetting(key, value);
    setSettings(prev => prev.map(f => (f.key === key ? updated : f)));
    return updated;
  }, []);

  const clearSetting = useCallback(async (key: string) => {
    const updated = await api.clearSetting(key);
    setSettings(prev => prev.map(f => (f.key === key ? updated : f)));
    return updated;
  }, []);

  const getField = useCallback(
    (key: string) => settings.find(f => f.key === key),
    [settings]
  );

  return { settings, bootstrap, loading, fetchSettings, setSetting, clearSetting, getField };
}
