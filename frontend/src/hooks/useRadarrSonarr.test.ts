import { afterEach, describe, expect, it, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useRadarrSonarr } from './useRadarrSonarr';
import { api } from '../services/api';

vi.mock('../services/api', () => ({
  api: {
    listRadarrMovies: vi.fn(),
    listSonarrSeries: vi.fn(),
    getRadarrSonarrStats: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    code: string;
    constructor(code: string, message: string) {
      super(message);
      this.code = code;
    }
  },
}));

function emptyPage() {
  return { data: [], total: 0, limit: 20, offset: 0, total_pages: 0 };
}

describe('useRadarrSonarr État status filter', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    window.history.replaceState(null, '', window.location.pathname);
  });

  it('defaults to a Monitored-only État filter when the URL has no filmsStatus/seriesStatus param', async () => {
    vi.mocked(api.listRadarrMovies).mockResolvedValue(emptyPage());
    vi.mocked(api.listSonarrSeries).mockResolvedValue(emptyPage());
    vi.mocked(api.getRadarrSonarrStats).mockResolvedValue({ radarr_monitored: null, radarr_matched: null, sonarr_monitored: null, sonarr_matched: null });

    const { result } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    expect(result.current.filmsStatus).toEqual(new Set(['monitored']));
    expect(result.current.seriesStatus).toEqual(new Set(['monitored']));
    expect(api.listRadarrMovies).toHaveBeenCalledWith(1, 20, '', '', new Set(['monitored']));
    expect(api.listSonarrSeries).toHaveBeenCalledWith(1, 20, '', '', undefined, new Set(['monitored']));
  });

  it('round-trips a multi-value filmsStatus/seriesStatus through the URL', async () => {
    window.history.replaceState(null, '', '/?filmsStatus=unmonitored,missing&seriesStatus=missing');
    vi.mocked(api.listRadarrMovies).mockResolvedValue(emptyPage());
    vi.mocked(api.listSonarrSeries).mockResolvedValue(emptyPage());
    vi.mocked(api.getRadarrSonarrStats).mockResolvedValue({ radarr_monitored: null, radarr_matched: null, sonarr_monitored: null, sonarr_matched: null });

    const { result } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    expect(result.current.filmsStatus).toEqual(new Set(['unmonitored', 'missing']));
    expect(result.current.seriesStatus).toEqual(new Set(['missing']));
  });

  it('setFilmsStatus updates the URL and resets to page 1', async () => {
    vi.mocked(api.listRadarrMovies).mockResolvedValue(emptyPage());
    vi.mocked(api.listSonarrSeries).mockResolvedValue(emptyPage());
    vi.mocked(api.getRadarrSonarrStats).mockResolvedValue({ radarr_monitored: null, radarr_matched: null, sonarr_monitored: null, sonarr_matched: null });

    const { result, rerender } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    act(() => {
      result.current.setFilmsPage(3);
    });
    rerender();
    await act(async () => {});

    act(() => {
      result.current.setFilmsStatus(new Set(['missing']));
    });
    rerender();
    await act(async () => {});

    expect(result.current.filmsPage).toBe(1);
    expect(result.current.filmsStatus).toEqual(new Set(['missing']));
    expect(window.location.search).toContain('filmsStatus=missing');
  });
});
