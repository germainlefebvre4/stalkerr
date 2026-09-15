import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAppSettings } from './useAppSettings';
import { api } from '../services/api';
import { SettingsField } from '../types';

vi.mock('../services/api', () => ({
  api: {
    getSettings: vi.fn(),
    setSetting: vi.fn(),
    clearSetting: vi.fn(),
  },
}));

const radarrUrl: SettingsField = {
  key: 'radarr.url', value: 'http://file.example.com', sensitive: false, origin: 'config', restart_required: false,
};

describe('useAppSettings', () => {
  afterEach(() => vi.restoreAllMocks());

  it('fetches settings', async () => {
    vi.mocked(api.getSettings).mockResolvedValue({ settings: [radarrUrl] });

    const { result } = renderHook(() => useAppSettings());
    await act(async () => {
      await result.current.fetchSettings();
    });

    expect(result.current.settings).toEqual([radarrUrl]);
  });

  it('setSetting stores the returned field in place, updating its origin', async () => {
    vi.mocked(api.getSettings).mockResolvedValue({ settings: [radarrUrl] });

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
