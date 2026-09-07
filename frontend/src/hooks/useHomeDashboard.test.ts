import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useHomeDashboard } from './useHomeDashboard';
import { api } from '../services/api';

vi.mock('../services/api', () => ({
  api: {
    getLogs: vi.fn(),
    getDownloads: vi.fn(),
  },
}));

function emptyPage() {
  return { data: [], total: 0, limit: 1, offset: 0, total_pages: 0 };
}

describe('useHomeDashboard', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('does not fetch anything while inactive', async () => {
    vi.mocked(api.getLogs).mockResolvedValue(emptyPage());
    vi.mocked(api.getDownloads).mockResolvedValue(emptyPage());

    renderHook(() => useHomeDashboard(false));
    await act(async () => {});

    expect(api.getLogs).not.toHaveBeenCalled();
    expect(api.getDownloads).not.toHaveBeenCalled();
  });

  it('fetches the latest processing log, downloads total, and errors total once active', async () => {
    vi.mocked(api.getLogs).mockResolvedValue(emptyPage());
    vi.mocked(api.getDownloads).mockResolvedValue(emptyPage());

    renderHook(() => useHomeDashboard(true));
    await act(async () => {});

    expect(api.getLogs).toHaveBeenCalledWith(1);
    expect(api.getDownloads).toHaveBeenCalledWith(1);
    expect(api.getDownloads).toHaveBeenCalledWith(1, undefined, undefined, 'missing_year,year_mismatch,unknown_format');
  });

  it('exposes the latest log and the downloads/errors totals from the responses', async () => {
    const log = { id: 1, action: 'process_m3u', item_count: 5, status: 'success' as const, started_at: '2026-01-01T00:00:00Z' };
    vi.mocked(api.getLogs).mockResolvedValue({ data: [log], total: 1, limit: 1, offset: 0, total_pages: 1 });
    vi.mocked(api.getDownloads)
      .mockResolvedValueOnce({ data: [], total: 42, limit: 1, offset: 0, total_pages: 42 })
      .mockResolvedValueOnce({ data: [], total: 3, limit: 1, offset: 0, total_pages: 3 });

    const { result } = renderHook(() => useHomeDashboard(true));
    await act(async () => {});

    expect(result.current.latestLog).toEqual(log);
    expect(result.current.downloadsTotal).toBe(42);
    expect(result.current.errorsTotal).toBe(3);
  });
});
