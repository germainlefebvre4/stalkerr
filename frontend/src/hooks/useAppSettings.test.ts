import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAppSettings } from './useAppSettings';
import { api } from '../services/api';
import { SettingsField, BootstrapField } from '../types';

vi.mock('../services/api', () => ({
  api: {
    getSettings: vi.fn(),
    getBootstrapSettings: vi.fn(),
    setSetting: vi.fn(),
    clearSetting: vi.fn(),
  },
}));

const radarrUrl: SettingsField = {
  key: 'radarr.url', value: 'http://file.example.com', sensitive: false, origin: 'config', restart_required: false,
};
const dbHost: BootstrapField = { key: 'database.host', value: 'db', sensitive: false, origin: 'config' };

describe('useAppSettings', () => {
  afterEach(() => vi.restoreAllMocks());

  it('fetches both settings and bootstrap configuration', async () => {
    vi.mocked(api.getSettings).mockResolvedValue({ settings: [radarrUrl] });
    vi.mocked(api.getBootstrapSettings).mockResolvedValue({ bootstrap: [dbHost] });

    const { result } = renderHook(() => useAppSettings());
    await act(async () => {
      await result.current.fetchSettings();
    });

    expect(result.current.settings).toEqual([radarrUrl]);
    expect(result.current.bootstrap).toEqual([dbHost]);
  });

  it('setSetting stores the returned field in place, updating its origin', async () => {
    vi.mocked(api.getSettings).mockResolvedValue({ settings: [radarrUrl] });
    vi.mocked(api.getBootstrapSettings).mockResolvedValue({ bootstrap: [] });

    const { result } = renderHook(() => useAppSettings());
    await act(async () => {
      await result.current.fetchSettings();
    });

    const overridden: SettingsField = { ...radarrUrl, value: 'http://override.example.com', origin: 'interface' };
    vi.mocked(api.setSetting).mockResolvedValue(overridden);

    await act(async () => {
      await result.current.setSetting('radarr.url', 'http://override.example.com');
    });

    expect(api.setSetting).toHaveBeenCalledWith('radarr.url', 'http://override.example.com');
    expect(result.current.getField('radarr.url')).toEqual(overridden);
  });

  it('clearSetting stores the returned reverted field in place', async () => {
    const overridden: SettingsField = { ...radarrUrl, value: 'http://override.example.com', origin: 'interface' };
    vi.mocked(api.getSettings).mockResolvedValue({ settings: [overridden] });
    vi.mocked(api.getBootstrapSettings).mockResolvedValue({ bootstrap: [] });

    const { result } = renderHook(() => useAppSettings());
    await act(async () => {
      await result.current.fetchSettings();
    });

    vi.mocked(api.clearSetting).mockResolvedValue(radarrUrl);

    await act(async () => {
      await result.current.clearSetting('radarr.url');
    });

    expect(api.clearSetting).toHaveBeenCalledWith('radarr.url');
    expect(result.current.getField('radarr.url')).toEqual(radarrUrl);
  });
});
