import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useM3uSources } from './useM3uSources';
import { api } from '../services/api';
import { M3uSource, M3uSourceInput } from '../types';

vi.mock('../services/api', () => ({
  api: {
    getM3uSources: vi.fn(),
    getM3uSourcesOrigin: vi.fn(),
    createM3uSource: vi.fn(),
    updateM3uSource: vi.fn(),
    deleteM3uSource: vi.fn(),
  },
}));

const originSource: M3uSource = {
  name: 'a', file_path: '/tmp/a.m3u', enabled: true, url: '', archive_dir: '',
  retention_count: 5, max_file_size_mb: 500, timeout_seconds: 300, retry_attempts: 3,
  auth_username: '', has_auth_password: false, is_runtime: false,
};

const input: M3uSourceInput = {
  file_path: '/tmp/c.m3u', enabled: true, url: '', archive_dir: '',
  retention_count: 5, max_file_size_mb: 500, timeout_seconds: 300, retry_attempts: 3, auth_username: '',
};

describe('useM3uSources', () => {
  afterEach(() => vi.restoreAllMocks());

  it('fetches the effective list and the set of origin names', async () => {
    vi.mocked(api.getM3uSources).mockResolvedValue({ sources: [originSource] });
    vi.mocked(api.getM3uSourcesOrigin).mockResolvedValue({ sources: [originSource] });

    const { result } = renderHook(() => useM3uSources());
    await act(async () => {
      await result.current.fetchSources();
    });

    expect(result.current.sources).toEqual([originSource]);
    expect(result.current.originNames.has('a')).toBe(true);
  });

  it('createSource calls the API then refreshes the list', async () => {
    vi.mocked(api.getM3uSources).mockResolvedValue({ sources: [] });
    vi.mocked(api.getM3uSourcesOrigin).mockResolvedValue({ sources: [] });
    vi.mocked(api.createM3uSource).mockResolvedValue({ ...originSource, name: 'c', is_runtime: true });

    const { result } = renderHook(() => useM3uSources());
    await act(async () => {
      await result.current.createSource('c', input);
    });

    expect(api.createM3uSource).toHaveBeenCalledWith('c', input);
    expect(api.getM3uSources).toHaveBeenCalled();
  });

  it('updateSource and deleteSource call their respective API methods then refresh', async () => {
    vi.mocked(api.getM3uSources).mockResolvedValue({ sources: [] });
    vi.mocked(api.getM3uSourcesOrigin).mockResolvedValue({ sources: [] });
    vi.mocked(api.updateM3uSource).mockResolvedValue({ ...originSource, is_runtime: true });
    vi.mocked(api.deleteM3uSource).mockResolvedValue({});

    const { result } = renderHook(() => useM3uSources());
    await act(async () => {
      await result.current.updateSource('a', input);
    });
    expect(api.updateM3uSource).toHaveBeenCalledWith('a', input);

    await act(async () => {
      await result.current.deleteSource('a');
    });
    expect(api.deleteM3uSource).toHaveBeenCalledWith('a');
  });
});
