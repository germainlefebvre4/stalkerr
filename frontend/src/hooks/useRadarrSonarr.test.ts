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

const EMPTY_STATS = { radarr_monitored: null, radarr_matched: null, sonarr_monitored: null, sonarr_matched: null };

function mockEmptyApi() {
  vi.mocked(api.listRadarrMovies).mockResolvedValue(emptyPage());
  vi.mocked(api.listSonarrSeries).mockResolvedValue(emptyPage());
  vi.mocked(api.getRadarrSonarrStats).mockResolvedValue(EMPTY_STATS);
}

describe('useRadarrSonarr Suivi and Fichier filters', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    window.history.replaceState(null, '', window.location.pathname);
  });

  it('defaults to Monitored-only Suivi and an empty Fichier filter when the URL has no params', async () => {
    mockEmptyApi();

    const { result } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    expect(result.current.filmsStatus).toEqual(new Set(['monitored']));
    expect(result.current.seriesStatus).toEqual(new Set(['monitored']));
    expect(result.current.filmsFile).toEqual(new Set());
    expect(result.current.seriesFile).toEqual(new Set());
    expect(api.listRadarrMovies).toHaveBeenCalledWith(1, 20, '', '', new Set(['monitored']), new Set());
    expect(api.listSonarrSeries).toHaveBeenCalledWith(1, 20, '', '', undefined, new Set(['monitored']), new Set());
  });

  it('round-trips multi-value status and file params through the URL', async () => {
    window.history.replaceState(null, '', '/?filmsStatus=monitored,unmonitored&filmsFile=none&seriesFile=partial,none');
    mockEmptyApi();

    const { result } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    expect(result.current.filmsStatus).toEqual(new Set(['monitored', 'unmonitored']));
    expect(result.current.filmsFile).toEqual(new Set(['none']));
    expect(result.current.seriesFile).toEqual(new Set(['partial', 'none']));
  });

  it('falls back to the default when the obsolete "missing" value is in the URL', async () => {
    window.history.replaceState(null, '', '/?filmsStatus=missing&seriesStatus=monitored,missing&filmsFile=bogus');
    mockEmptyApi();

    const { result } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    expect(result.current.filmsStatus).toEqual(new Set(['monitored']));
    expect(result.current.seriesStatus).toEqual(new Set(['monitored']));
    expect(result.current.filmsFile).toEqual(new Set());
  });

  it('setFilmsStatus and setFilmsFile update the URL and reset to page 1', async () => {
    mockEmptyApi();

    const { result, rerender } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    act(() => {
      result.current.setFilmsPage(3);
    });
    rerender();
    await act(async () => {});

    act(() => {
      result.current.setFilmsStatus(new Set(['unmonitored']));
    });
    rerender();
    await act(async () => {});
    expect(result.current.filmsPage).toBe(1);
    expect(window.location.search).toContain('filmsStatus=unmonitored');

    act(() => {
      result.current.setFilmsPage(2);
    });
    rerender();
    act(() => {
      result.current.setFilmsFile(new Set(['downloaded']));
    });
    rerender();
    await act(async () => {});
    expect(result.current.filmsPage).toBe(1);
    expect(window.location.search).toContain('filmsFile=downloaded');
  });

  it('the Manquants presets set both axes at once and fetch with them', async () => {
    mockEmptyApi();

    const { result, rerender } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    act(() => {
      result.current.setSeriesMissing();
      result.current.setFilmsMissing();
    });
    rerender();
    await act(async () => {});

    expect(result.current.seriesStatus).toEqual(new Set(['monitored']));
    expect(result.current.seriesFile).toEqual(new Set(['partial', 'none']));
    expect(result.current.filmsFile).toEqual(new Set(['none']));
    expect(api.listSonarrSeries).toHaveBeenLastCalledWith(1, 20, '', '', undefined, new Set(['monitored']), new Set(['partial', 'none']));
    expect(api.listRadarrMovies).toHaveBeenLastCalledWith(1, 20, '', '', new Set(['monitored']), new Set(['none']));
  });

  it('reset restores the section defaults in one fetch, leaves the other section and the search untouched', async () => {
    mockEmptyApi();

    const { result, rerender } = renderHook(() => useRadarrSonarr(true));
    await act(async () => {});

    act(() => {
      result.current.setFilmsSearch('matrix');
      result.current.setFilmsFilter('matched');
      result.current.setFilmsStatus(new Set(['unmonitored']));
      result.current.setFilmsFile(new Set(['downloaded']));
      result.current.setSeriesFile(new Set(['none']));
      result.current.setFilmsPage(3);
    });
    rerender();
    await act(async () => {});

    vi.mocked(api.listRadarrMovies).mockClear();
    act(() => {
      result.current.resetFilmsFilters();
    });
    rerender();
    await act(async () => {});

    expect(result.current.filmsFilter).toBe('');
    expect(result.current.filmsStatus).toEqual(new Set(['monitored']));
    expect(result.current.filmsFile).toEqual(new Set());
    expect(result.current.filmsPage).toBe(1);
    expect(result.current.filmsSearch).toBe('matrix');
    expect(result.current.seriesFile).toEqual(new Set(['none']));
    expect(api.listRadarrMovies).toHaveBeenCalledTimes(1);
    expect(api.listRadarrMovies).toHaveBeenLastCalledWith(1, 20, 'matrix', '', new Set(['monitored']), new Set());

    act(() => {
      result.current.resetSeriesFilters();
    });
    rerender();
    await act(async () => {});
    expect(result.current.seriesFile).toEqual(new Set());
    expect(result.current.seriesPage).toBe(1);
  });
});
