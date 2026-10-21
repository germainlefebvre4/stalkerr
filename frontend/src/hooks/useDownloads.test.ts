import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
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

describe('useDownloads status default and search', () => {
  const page = { data: [], total: 0, limit: 20, offset: 0, total_pages: 0, cancelled_hidden: 4 };

  beforeEach(() => {
    window.history.replaceState(null, '', '/');
    vi.mocked(api.getDownloads).mockReset().mockResolvedValue(page);
    vi.mocked(api.getConfigPaths).mockResolvedValue({ movies_path: '', tvshows_path: '' });
  });
  afterEach(() => {
    vi.useRealTimers();
    window.history.replaceState(null, '', '/');
  });

  it('defaults to status active and exposes cancelledHidden', async () => {
    const { result } = renderHook(() => useDownloads(true));
    await waitFor(() => expect(result.current.cancelledHidden).toBe(4));
    expect(result.current.statusFilter).toBe('active');
    expect(api.getDownloads).toHaveBeenLastCalledWith(20, 'active', '', '', 0, '');
  });

  it('maps dlStatus=all to no status param', async () => {
    window.history.replaceState(null, '', '/?dlStatus=all');
    renderHook(() => useDownloads(true));
    await waitFor(() => expect(api.getDownloads).toHaveBeenCalled());
    expect(api.getDownloads).toHaveBeenLastCalledWith(20, '', '', '', 0, '');
  });

  it('debounces search, persists dlSearch and resets the page', async () => {
    window.history.replaceState(null, '', '/?dlPage=3');
    const { result } = renderHook(() => useDownloads(true));
    await waitFor(() => expect(api.getDownloads).toHaveBeenCalled());
    vi.mocked(api.getDownloads).mockClear();
    vi.useFakeTimers();

    act(() => result.current.setSearchInput('peace'));
    act(() => result.current.setSearchInput('peacemaker'));
    await act(async () => { await vi.advanceTimersByTimeAsync(299); });
    expect(api.getDownloads).not.toHaveBeenCalled();
    await act(async () => { await vi.advanceTimersByTimeAsync(10); });

    expect(window.location.search).toContain('dlSearch=peacemaker');
    expect(result.current.downloadsPage).toBe(1);
    expect(api.getDownloads).toHaveBeenLastCalledWith(20, 'active', '', '', 0, 'peacemaker');
  });

  it('reads searchName as an alias of dlSearch', async () => {
    window.history.replaceState(null, '', '/?tab=downloads&searchName=peacemaker');
    const { result } = renderHook(() => useDownloads(true));
    await waitFor(() => expect(api.getDownloads).toHaveBeenLastCalledWith(20, 'active', '', '', 0, 'peacemaker'));
    expect(result.current.searchInput).toBe('peacemaker');
    expect(window.location.search).toContain('searchName=peacemaker');
  });
});
