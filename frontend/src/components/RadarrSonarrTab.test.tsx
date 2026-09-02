import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, cleanup, act, waitFor } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { RadarrSonarrTab } from './RadarrSonarrTab';
import { RadarrMovieListItem, SonarrSeriesListItem, RadarrSonarrStats, SonarrSeriesEpisodesResponse, MatchStatusFilter } from '../types';
import { api } from '../services/api';
import { useIsMobile } from '../hooks/useMediaQuery';

vi.mock('../services/api', () => ({
  api: {
    getRadarrMovieMatches: vi.fn(),
    getSonarrSeriesEpisodes: vi.fn(),
    getItem: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    code: string;
    constructor(code: string, message: string) {
      super(message);
      this.code = code;
    }
  },
}));

vi.mock('../hooks/useMediaQuery', () => ({
  useIsMobile: vi.fn(() => false),
}));

afterEach(() => {
  cleanup();
  window.history.replaceState(null, '', '/');
  vi.mocked(useIsMobile).mockReturnValue(false);
  vi.clearAllMocks();
});

interface Overrides {
  filmsItems?: RadarrMovieListItem[];
  filmsSearch?: string;
  setFilmsSearch?: (value: string) => void;
  filmsFilter?: MatchStatusFilter;
  setFilmsFilter?: (value: MatchStatusFilter) => void;
  seriesItems?: SonarrSeriesListItem[];
  seriesSearch?: string;
  setSeriesSearch?: (value: string) => void;
  seriesFilter?: MatchStatusFilter;
  setSeriesFilter?: (value: MatchStatusFilter) => void;
  refreshSeries?: () => void;
  stats?: RadarrSonarrStats | null;
  statsLoading?: boolean;
  statsError?: string | null;
}

function renderTab(overrides: Overrides = {}) {
  return render(
    <I18nextProvider i18n={i18n}>
      <Tabs.Root value="radarr-sonarr">
        <RadarrSonarrTab
          filmsItems={overrides.filmsItems ?? []}
          filmsLoading={false}
          filmsError={null}
          filmsTotal={overrides.filmsItems?.length ?? 0}
          filmsPage={1}
          setFilmsPage={() => {}}
          filmsLimit={20}
          fetchFilms={() => {}}
          filmsSearch={overrides.filmsSearch ?? ''}
          setFilmsSearch={overrides.setFilmsSearch ?? (() => {})}
          filmsFilter={overrides.filmsFilter ?? ''}
          setFilmsFilter={overrides.setFilmsFilter ?? (() => {})}
          seriesItems={overrides.seriesItems ?? []}
          seriesLoading={false}
          seriesError={null}
          seriesTotal={overrides.seriesItems?.length ?? 0}
          seriesPage={1}
          setSeriesPage={() => {}}
          seriesLimit={20}
          fetchSeries={() => {}}
          refreshSeries={overrides.refreshSeries ?? (() => {})}
          seriesSearch={overrides.seriesSearch ?? ''}
          setSeriesSearch={overrides.setSeriesSearch ?? (() => {})}
          seriesFilter={overrides.seriesFilter ?? ''}
          setSeriesFilter={overrides.setSeriesFilter ?? (() => {})}
          stats={overrides.stats ?? null}
          statsLoading={overrides.statsLoading ?? false}
          statsError={overrides.statsError ?? null}
          fetchStats={() => {}}
        />
      </Tabs.Root>
    </I18nextProvider>
  );
}

describe('RadarrSonarrTab sub-tabs', () => {
  it('defaults to the Summary sub-tab and shows the monitoring summary', () => {
    const stats: RadarrSonarrStats = { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 5 };
    renderTab({ stats });

    expect(screen.getByText('10')).toBeInTheDocument();
    expect(screen.getByText('7')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('switching to the Sonarr sub-tab shows Séries content and hides Films', () => {
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, has_file: true, matched: true, occurrence_count: 0 };
    const series: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, matched_count: 2, monitored_count: 4, occurrence_count: 0 };
    renderTab({ filmsItems: [movie], seriesItems: [series] });

    fireEvent.mouseDown(screen.getByText('Sonarr'), { button: 0 });

    expect(screen.getByText('Example Series')).toBeInTheDocument();
    expect(screen.queryByText('Example Movie')).not.toBeInTheDocument();
  });

  it('switching to the Radarr sub-tab shows Films content', () => {
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, has_file: true, matched: true, occurrence_count: 0 };
    renderTab({ filmsItems: [movie] });

    fireEvent.mouseDown(screen.getByText('Radarr'), { button: 0 });

    expect(screen.getByText('Example Movie')).toBeInTheDocument();
  });
});

describe('RadarrSonarrTab sub-tab persistence', () => {
  it('persists the selected sub-tab in the URL query string', () => {
    renderTab();

    fireEvent.mouseDown(screen.getByText('Sonarr'), { button: 0 });

    expect(window.location.search).toContain('subtab=sonarr');
  });

  it('restores the sub-tab from the URL on mount, matching a page reload', () => {
    window.history.replaceState(null, '', '/?subtab=sonarr');
    const series: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, matched_count: 2, monitored_count: 4, occurrence_count: 0 };
    renderTab({ seriesItems: [series] });

    expect(screen.getByText('Example Series')).toBeInTheDocument();
  });
});

describe('RadarrSonarrTab sub-tab icons', () => {
  it('renders the Radarr and Sonarr icons but not for Résumé', () => {
    renderTab();

    expect(screen.getByText('Radarr').querySelector('img')).toBeInTheDocument();
    expect(screen.getByText('Sonarr').querySelector('img')).toBeInTheDocument();
    expect(screen.getByText('Summary').querySelector('img')).not.toBeInTheDocument();
  });
});

describe('RadarrSonarrTab search', () => {
  it('debounces the Films search input before calling setFilmsSearch', () => {
    vi.useFakeTimers();
    const setFilmsSearch = vi.fn();
    renderTab({ setFilmsSearch });
    fireEvent.mouseDown(screen.getByText('Radarr'), { button: 0 });

    const input = screen.getByPlaceholderText('Search movies by title...');
    fireEvent.change(input, { target: { value: 'matrix' } });

    expect(setFilmsSearch).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(300);
    });

    expect(setFilmsSearch).toHaveBeenCalledWith('matrix');
    vi.useRealTimers();
  });

  it('shows a search-specific empty state distinct from the no-items empty state', () => {
    renderTab({ filmsSearch: 'nomatch', filmsItems: [] });
    fireEvent.mouseDown(screen.getByText('Radarr'), { button: 0 });

    expect(screen.getByText('No movies match this search')).toBeInTheDocument();
    expect(screen.queryByText('No monitored movies found')).not.toBeInTheDocument();
  });

  it('shows the default empty state when there is no search term', () => {
    renderTab({ filmsItems: [] });
    fireEvent.mouseDown(screen.getByText('Radarr'), { button: 0 });

    expect(screen.getByText('No monitored movies found')).toBeInTheDocument();
  });
});

const seriesEpisodesFixture: SonarrSeriesEpisodesResponse = {
  episodes: [
    { season: 1, episode: 1, matched: true, occurrences: [] },
    { season: 1, episode: 2, matched: false, occurrences: [] },
    { season: 2, episode: 1, matched: true, occurrences: [{ id: 101, resolution: '1080p', state: 'downloaded' }] },
    { season: 2, episode: 2, matched: true, occurrences: [] },
  ],
};

const seriesFixture: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, matched_count: 3, monitored_count: 4, occurrence_count: 0 };

function openExampleSeries(overrides: Overrides = {}) {
  renderTab({ seriesItems: [seriesFixture], ...overrides });
  fireEvent.mouseDown(screen.getByText('Sonarr'), { button: 0 });
  fireEvent.click(screen.getByText('Example Series'));
}

describe('RadarrSonarrTab series season grouping', () => {
  it('groups episodes by season with correct per-season matched/total counts', async () => {
    vi.mocked(api.getSonarrSeriesEpisodes).mockResolvedValue(seriesEpisodesFixture);
    openExampleSeries();

    await waitFor(() => expect(screen.getByText('Season 1')).toBeInTheDocument());
    expect(screen.getByText('1/2')).toBeInTheDocument();
    expect(screen.getByText('Season 2')).toBeInTheDocument();
    expect(screen.getByText('2/2')).toBeInTheDocument();
  });

  it('allows only one season to be expanded at a time', async () => {
    vi.mocked(api.getSonarrSeriesEpisodes).mockResolvedValue(seriesEpisodesFixture);
    openExampleSeries();
    await waitFor(() => expect(screen.getByText('Season 1')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Season 1'));
    expect(screen.getByText('S01 E01')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Season 2'));
    expect(screen.queryByText('S01 E01')).not.toBeInTheDocument();
    expect(screen.getByText('S02 E01')).toBeInTheDocument();
  });

  it('allows only one episode to be expanded at a time within a season, with a visible active state', async () => {
    vi.mocked(api.getSonarrSeriesEpisodes).mockResolvedValue(seriesEpisodesFixture);
    openExampleSeries();
    await waitFor(() => expect(screen.getByText('Season 2')).toBeInTheDocument());
    fireEvent.click(screen.getByText('Season 2'));
    await waitFor(() => expect(screen.getByText('S02 E01')).toBeInTheDocument());

    fireEvent.click(screen.getByText('S02 E01'));
    expect(screen.getByText('S02 E01').closest('tr')).toHaveClass('clickable-row--active');

    fireEvent.click(screen.getByText('S02 E02'));
    expect(screen.getByText('S02 E01').closest('tr')).not.toHaveClass('clickable-row--active');
    expect(screen.getByText('S02 E02').closest('tr')).toHaveClass('clickable-row--active');
  });
});

describe('RadarrSonarrTab mobile sub-tab switcher', () => {
  it('renders with its own class and remains operable when isMobile is true', () => {
    vi.mocked(useIsMobile).mockReturnValue(true);
    renderTab({ seriesItems: [seriesFixture] });

    const tabsList = screen.getByText('Radarr').closest('.segmented-tabs-list');
    expect(tabsList).toHaveClass('radarr-sonarr-subtabs');

    fireEvent.mouseDown(screen.getByText('Sonarr'), { button: 0 });
    expect(screen.getByText('Example Series')).toBeInTheDocument();
  });
});

describe('RadarrSonarrTab mobile sidepanel lists', () => {
  it('renders movie occurrences as mobile list cards instead of a table', async () => {
    vi.mocked(useIsMobile).mockReturnValue(true);
    vi.mocked(api.getRadarrMovieMatches).mockResolvedValue({
      matched: true,
      occurrences: [{ id: 201, resolution: '1080p', state: 'downloaded' }],
    });
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, has_file: true, matched: true, occurrence_count: 0 };
    renderTab({ filmsItems: [movie] });
    fireEvent.mouseDown(screen.getByText('Radarr'), { button: 0 });
    fireEvent.click(screen.getByText('Example Movie'));

    await waitFor(() => expect(screen.getByText('1080p')).toBeInTheDocument());
    expect(screen.getByText('1080p').closest('table')).toBeNull();
    expect(screen.getByText('1080p').closest('.mobile-list-card')).not.toBeNull();
  });

  it('renders the season/episode list as mobile list cards instead of a table', async () => {
    vi.mocked(useIsMobile).mockReturnValue(true);
    vi.mocked(api.getSonarrSeriesEpisodes).mockResolvedValue(seriesEpisodesFixture);
    openExampleSeries();
    await waitFor(() => expect(screen.getByText('Season 1')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Season 1'));
    expect(screen.getByText('S01 E01')).toBeInTheDocument();
    expect(screen.getByText('S01 E01').closest('table')).toBeNull();
    expect(screen.getByText('S01 E01').closest('.mobile-list-card')).not.toBeNull();
  });
});
