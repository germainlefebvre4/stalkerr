import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, cleanup, act } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { RadarrSonarrTab } from './RadarrSonarrTab';
import { RadarrMovieListItem, SonarrSeriesListItem, RadarrSonarrStats } from '../types';

afterEach(() => {
  cleanup();
  window.history.replaceState(null, '', '/');
});

interface Overrides {
  filmsItems?: RadarrMovieListItem[];
  filmsSearch?: string;
  setFilmsSearch?: (value: string) => void;
  seriesItems?: SonarrSeriesListItem[];
  seriesSearch?: string;
  setSeriesSearch?: (value: string) => void;
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
          seriesItems={overrides.seriesItems ?? []}
          seriesLoading={false}
          seriesError={null}
          seriesTotal={overrides.seriesItems?.length ?? 0}
          seriesPage={1}
          setSeriesPage={() => {}}
          seriesLimit={20}
          fetchSeries={() => {}}
          seriesSearch={overrides.seriesSearch ?? ''}
          setSeriesSearch={overrides.setSeriesSearch ?? (() => {})}
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
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, has_file: true, matched: true };
    const series: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, matched_count: 2, monitored_count: 4 };
    renderTab({ filmsItems: [movie], seriesItems: [series] });

    fireEvent.mouseDown(screen.getByText('Sonarr'), { button: 0 });

    expect(screen.getByText('Example Series')).toBeInTheDocument();
    expect(screen.queryByText('Example Movie')).not.toBeInTheDocument();
  });

  it('switching to the Radarr sub-tab shows Films content', () => {
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, has_file: true, matched: true };
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
    const series: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, matched_count: 2, monitored_count: 4 };
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
