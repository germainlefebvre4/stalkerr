import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDownloads } from './useDownloads';
import { api } from '../services/api';

vi.mock('../services/api', () => ({
  api: {
    getDownloads: vi.fn(),
    getConfigPaths: vi.fn(),
  },
}));

describe('useDownloads background refresh', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it('does not re-fetch on a timer after the initial load when the user takes no action', async () => {
    vi.mocked(api.getDownloads).mockResolvedValue({ data: [], total: 0, limit: 20, offset: 0, total_pages: 0 });
    vi.mocked(api.getConfigPaths).mockResolvedValue({ movies_path: '', tvshows_path: '' });

    vi.useFakeTimers();
    renderHook(() => useDownloads(true));

    await act(async () => {
      await vi.runOnlyPendingTimersAsync();
    });
    expect(api.getDownloads).toHaveBeenCalledTimes(1);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(10000);
    });
    expect(api.getDownloads).toHaveBeenCalledTimes(1);
  });
});
