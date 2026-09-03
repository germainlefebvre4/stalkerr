import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useErrorsTab } from './useErrorsTab';
import { api } from '../services/api';

vi.mock('../services/api', () => ({
  api: {
    getDownloads: vi.fn(),
  },
}));

function emptyPage() {
  return { data: [], total: 0, limit: 20, offset: 0, total_pages: 0 };
}

describe('useErrorsTab', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    window.history.replaceState(null, '', window.location.pathname);
  });

  it('fetches the combined problem filter by default', async () => {
    vi.mocked(api.getDownloads).mockResolvedValue(emptyPage());

    renderHook(() => useErrorsTab(true));

    await act(async () => {});

    expect(api.getDownloads).toHaveBeenCalledWith(20, '', '', 'missing_year,year_mismatch,unknown_format', 0);
  });

  it('re-fetches with the single problem value and resets to page 1 when a reason is selected on a later page', async () => {
    vi.mocked(api.getDownloads).mockResolvedValue(emptyPage());

    const { result, rerender } = renderHook(() => useErrorsTab(true));

    await act(async () => {});

    act(() => {
      result.current.setErrorsPage(3);
    });
    rerender();
    await act(async () => {});

    vi.mocked(api.getDownloads).mockClear();

    act(() => {
      result.current.setReasonFilter('unknown_format');
    });
    rerender();
    await act(async () => {});

    expect(result.current.errorsPage).toBe(1);
    expect(api.getDownloads).toHaveBeenCalledWith(20, '', '', 'unknown_format', 0);
  });

  it('restores the combined filter when returning to "Tous"', async () => {
    vi.mocked(api.getDownloads).mockResolvedValue(emptyPage());

    const { result, rerender } = renderHook(() => useErrorsTab(true));
    await act(async () => {});

    act(() => {
      result.current.setReasonFilter('missing_year');
    });
    rerender();
    await act(async () => {});

    vi.mocked(api.getDownloads).mockClear();

    act(() => {
      result.current.setReasonFilter('');
    });
    rerender();
    await act(async () => {});

    expect(api.getDownloads).toHaveBeenCalledWith(20, '', '', 'missing_year,year_mismatch,unknown_format', 0);
  });
});
