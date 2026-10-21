import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, cleanup, act, waitFor } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { RadarrSonarrTab } from './RadarrSonarrTab';
import { RadarrMovieListItem, SonarrSeriesListItem, RadarrSonarrStats, SonarrSeriesEpisodesResponse, MatchStatusFilter, SuiviFilter, FileFilter } from '../types';
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
  filmsStatus?: SuiviFilter;
  setFilmsStatus?: (value: SuiviFilter) => void;
  filmsFile?: FileFilter;
  setFilmsFile?: (value: FileFilter) => void;
  setFilmsMissing?: () => void;
  resetFilmsFilters?: () => void;
  seriesItems?: SonarrSeriesListItem[];
  seriesSearch?: string;
  setSeriesSearch?: (value: string) => void;
  seriesFilter?: MatchStatusFilter;
  setSeriesFilter?: (value: MatchStatusFilter) => void;
  seriesStatus?: SuiviFilter;
  setSeriesStatus?: (value: SuiviFilter) => void;
  seriesFile?: FileFilter;
  setSeriesFile?: (value: FileFilter) => void;
  setSeriesMissing?: () => void;
  resetSeriesFilters?: () => void;
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
          filmsStatus={overrides.filmsStatus ?? new Set(['monitored'])}
          setFilmsStatus={overrides.setFilmsStatus ?? (() => {})}
          filmsFile={overrides.filmsFile ?? new Set()}
          setFilmsFile={overrides.setFilmsFile ?? (() => {})}
          setFilmsMissing={overrides.setFilmsMissing ?? (() => {})}
          resetFilmsFilters={overrides.resetFilmsFilters ?? (() => {})}
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
          seriesStatus={overrides.seriesStatus ?? new Set(['monitored'])}
          setSeriesStatus={overrides.setSeriesStatus ?? (() => {})}
          seriesFile={overrides.seriesFile ?? new Set()}
          setSeriesFile={overrides.setSeriesFile ?? (() => {})}
          setSeriesMissing={overrides.setSeriesMissing ?? (() => {})}
          resetSeriesFilters={overrides.resetSeriesFilters ?? (() => {})}
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
    const stats: RadarrSonarrStats = { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 5, sonarr_matched: null };
    renderTab({ stats });

    expect(screen.getByText('10')).toBeInTheDocument();
    expect(screen.getByText('7')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('shows a success-styled status badge and matched-ratio progress bar for Radarr, and a success badge for Sonarr', () => {
    const stats: RadarrSonarrStats = { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 5, sonarr_matched: null };
    renderTab({ stats });

    const successBadges = screen.getAllByText('Success');
    expect(successBadges).toHaveLength(2);
    successBadges.forEach(badge => expect(badge.className).toContain('badge-success'));

    const indicators = document.querySelectorAll('.progress-indicator');
    const widths = Array.from(indicators).map(el => (el as HTMLElement).style.width);
    expect(widths).toContain('70%');
  });

  it('shows a failure-styled status badge for the failing service while the other stays success', () => {
    const stats: RadarrSonarrStats = { radarr_monitored: null, radarr_matched: null, radarr_error: 'radarr_unreachable', sonarr_monitored: 5, sonarr_matched: null };
    renderTab({ stats });

    expect(screen.getByText('Failed').className).toContain('badge-failed');
    expect(screen.getByText('Success').className).toContain('badge-success');
  });

  it('renders the Sonarr matched/unmatched breakdown and progress bar when sonarr_matched is present', () => {
    const stats: RadarrSonarrStats = { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 8, sonarr_matched: 6 };
    renderTab({ stats });

    expect(screen.getByText('8')).toBeInTheDocument();
    expect(screen.getByText('6')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();

    const indicators = document.querySelectorAll('.progress-indicator');
    const widths = Array.from(indicators).map(el => (el as HTMLElement).style.width);
    expect(widths).toContain('75%');
  });

  it('falls back to the total-only display for the Sonarr card when sonarr_matched is null, without breaking the Radarr card', () => {
    const stats: RadarrSonarrStats = { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 5, sonarr_matched: null };
    renderTab({ stats });

    expect(screen.getByText('10')).toBeInTheDocument();
    expect(screen.getByText('7')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);
  });

  it('switching to the Sonarr sub-tab shows Séries content and hides Films', () => {
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, monitored: true, file_state: 'downloaded', matched: true, occurrence_count: 0 };
    const series: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, monitored: true, file_state: 'downloaded', episode_file_count: 4, episode_count: 4, matched_count: 2, monitored_count: 4, occurrence_count: 0 };
    renderTab({ filmsItems: [movie], seriesItems: [series] });

    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });

    expect(screen.getByText('Example Series')).toBeInTheDocument();
    expect(screen.queryByText('Example Movie')).not.toBeInTheDocument();
  });

  it('switching to the Radarr sub-tab shows Films content', () => {
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, monitored: true, file_state: 'downloaded', matched: true, occurrence_count: 0 };
    renderTab({ filmsItems: [movie] });

    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    expect(screen.getByText('Example Movie')).toBeInTheDocument();
  });

  it('clicking the Films (Radarr) summary card on the Résumé sub-tab switches to the Radarr sub-tab', () => {
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, monitored: true, file_state: 'downloaded', matched: true, occurrence_count: 0 };
    const stats: RadarrSonarrStats = { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 5, sonarr_matched: null };
    renderTab({ filmsItems: [movie], stats });

    fireEvent.click(screen.getByText('Films (Radarr)').closest('.home-card')!);

    expect(window.location.search).toContain('subtab=radarr');
    expect(screen.getByText('Example Movie')).toBeInTheDocument();
  });

  it('clicking the Séries (Sonarr) summary card on the Résumé sub-tab switches to the Sonarr sub-tab', () => {
    const series: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, monitored: true, file_state: 'downloaded', episode_file_count: 4, episode_count: 4, matched_count: 2, monitored_count: 4, occurrence_count: 0 };
    const stats: RadarrSonarrStats = { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 5, sonarr_matched: null };
    renderTab({ seriesItems: [series], stats });

    fireEvent.click(screen.getByText('Séries (Sonarr)').closest('.home-card')!);

    expect(window.location.search).toContain('subtab=sonarr');
    expect(screen.getByText('Example Series')).toBeInTheDocument();
  });

  it('the Résumé sub-tab cards present a clickable affordance', () => {
    const stats: RadarrSonarrStats = { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 5, sonarr_matched: null };
    renderTab({ stats });

    expect(screen.getByText('Films (Radarr)').closest('.home-card')).toHaveClass('home-card--clickable');
    expect(screen.getByText('Séries (Sonarr)').closest('.home-card')).toHaveClass('home-card--clickable');
  });
});

describe('RadarrSonarrTab sub-tab persistence', () => {
  it('persists the selected sub-tab in the URL query string', () => {
    renderTab();

    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });

    expect(window.location.search).toContain('subtab=sonarr');
  });

  it('restores the sub-tab from the URL on mount, matching a page reload', () => {
    window.history.replaceState(null, '', '/?subtab=sonarr');
    const series: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, monitored: true, file_state: 'downloaded', episode_file_count: 4, episode_count: 4, matched_count: 2, monitored_count: 4, occurrence_count: 0 };
    renderTab({ seriesItems: [series] });

    expect(screen.getByText('Example Series')).toBeInTheDocument();
  });
});

describe('RadarrSonarrTab sub-tab icons', () => {
  it('renders the Radarr and Sonarr icons but not for Résumé', () => {
    renderTab();

    expect(screen.getByRole('tab', { name: 'Radarr' }).querySelector('img')).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Sonarr' }).querySelector('img')).toBeInTheDocument();
    expect(screen.getByText('Summary').querySelector('img')).not.toBeInTheDocument();
  });
});

describe('RadarrSonarrTab Suivi and Fichier filters', () => {
  it('renders the Suivi and Fichier dropdowns and the Manquants preset in both the Films and Séries filter bars', () => {
    renderTab();

    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });
    expect(screen.getByText('Monitoring: Monitored')).toBeInTheDocument();
    expect(screen.getByText('File: No filter')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Missing' })).toBeInTheDocument();

    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });
    expect(screen.getByText('Monitoring: Monitored')).toBeInTheDocument();
    expect(screen.getByText('File: No filter')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Missing' })).toBeInTheDocument();
  });

  it('calling setFilmsStatus from the Suivi dropdown reports the updated selection', () => {
    const setFilmsStatus = vi.fn();
    renderTab({ setFilmsStatus });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    fireEvent.keyDown(screen.getByText('Monitoring: Monitored'), { key: 'Enter' });
    fireEvent.click(screen.getByText('Unmonitored'));

    expect(setFilmsStatus).toHaveBeenCalledWith(new Set(['monitored', 'unmonitored']));
  }, 30000);

  it('calling setFilmsFile from the Fichier dropdown reports the selection', () => {
    const setFilmsFile = vi.fn();
    renderTab({ setFilmsFile });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    fireEvent.keyDown(screen.getByText('File: No filter'), { key: 'Enter' });
    fireEvent.click(screen.getByText('Downloaded'));

    expect(setFilmsFile).toHaveBeenCalledWith(new Set(['downloaded']));
  }, 30000);

  it('offers a Partial option for Séries but not for Films', () => {
    renderTab();
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });
    fireEvent.keyDown(screen.getByText('File: No filter'), { key: 'Enter' });
    expect(screen.queryByText('Partial')).not.toBeInTheDocument();
    cleanup();

    renderTab();
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });
    fireEvent.keyDown(screen.getByText('File: No filter'), { key: 'Enter' });
    expect(screen.getByText('Partial')).toBeInTheDocument();
  }, 60000);

  it('the Manquants preset calls its setter and is not active by default', () => {
    const setSeriesMissing = vi.fn();
    renderTab({ setSeriesMissing });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });

    const preset = screen.getByRole('button', { name: 'Missing' });
    expect(preset).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(preset);

    expect(setSeriesMissing).toHaveBeenCalledTimes(1);
  });

  it('the Films Manquants preset is active when Suivi is monitored only and Fichier is none', () => {
    renderTab({ filmsStatus: new Set(['monitored']), filmsFile: new Set(['none']) });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    expect(screen.getByRole('button', { name: 'Missing' })).toHaveAttribute('aria-pressed', 'true');
  });

  it('the Séries Manquants preset is active only for exactly partial + none', () => {
    renderTab({ seriesStatus: new Set(['monitored']), seriesFile: new Set(['partial', 'none']) });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });
    expect(screen.getByRole('button', { name: 'Missing' })).toHaveAttribute('aria-pressed', 'true');
    cleanup();

    renderTab({ seriesStatus: new Set(['monitored']), seriesFile: new Set(['none']) });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });
    expect(screen.getByRole('button', { name: 'Missing' })).toHaveAttribute('aria-pressed', 'false');
    cleanup();

    renderTab({ seriesStatus: new Set(['monitored', 'unmonitored']), seriesFile: new Set(['partial', 'none']) });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });
    expect(screen.getByRole('button', { name: 'Missing' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('a selection with no matching items shows the ordinary empty-results state, not an error', () => {
    renderTab({ filmsItems: [], filmsStatus: new Set(['unmonitored']), filmsFile: new Set(['downloaded']) });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    expect(screen.getByText('No monitored movies found')).toBeInTheDocument();
  });

  it('renders Suivi and Fichier as distinct columns in the Films table, with Absent red only when monitored', () => {
    const movies: RadarrMovieListItem[] = [
      { radarr_id: 1, title: 'Absent Monitored', year: 2020, monitored: true, file_state: 'none', matched: false, occurrence_count: 0 },
      { radarr_id: 2, title: 'Downloaded Monitored', year: 2021, monitored: true, file_state: 'downloaded', matched: false, occurrence_count: 0 },
      { radarr_id: 3, title: 'Absent Unmonitored', year: 2022, monitored: false, file_state: 'none', matched: false, occurrence_count: 0 },
      { radarr_id: 4, title: 'Downloaded Unmonitored', year: 2023, monitored: false, file_state: 'downloaded', matched: false, occurrence_count: 0 },
    ];
    renderTab({ filmsItems: movies, filmsStatus: new Set(['monitored', 'unmonitored']) });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    const rowBadges = (title: string) => {
      const cells = screen.getByText(title).closest('tr')!.querySelectorAll('td');
      return { suivi: cells[2].querySelector('.badge')!, fichier: cells[3].querySelector('.badge')! };
    };

    let b = rowBadges('Absent Monitored');
    expect(b.suivi.textContent).toBe('Monitored');
    expect(b.fichier.textContent).toBe('Absent');
    expect(b.fichier.className).toContain('badge-failed');

    b = rowBadges('Downloaded Monitored');
    expect(b.fichier.textContent).toBe('Downloaded');
    expect(b.fichier.className).toContain('badge-success');

    b = rowBadges('Absent Unmonitored');
    expect(b.suivi.textContent).toBe('Unmonitored');
    expect(b.suivi.className).toContain('badge-neutral');
    expect(b.fichier.className).toContain('badge-neutral');

    b = rowBadges('Downloaded Unmonitored');
    expect(b.fichier.className).toContain('badge-success');
  });

  it('renders a partial series as an orange Fichier badge with the downloaded/aired ratio', () => {
    const series: SonarrSeriesListItem[] = [
      { sonarr_id: 1, title: 'Partial Series', year: 2019, monitored: true, file_state: 'partial', episode_file_count: 18, episode_count: 28, matched_count: 2, monitored_count: 4, occurrence_count: 0 },
      { sonarr_id: 2, title: 'Absent Series', year: 2019, monitored: true, file_state: 'none', episode_file_count: 0, episode_count: 8, matched_count: 0, monitored_count: 8, occurrence_count: 0 },
    ];
    renderTab({ seriesItems: series });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });

    const partial = screen.getByText('Partial 18/28');
    expect(partial.className).toContain('badge-pending');

    const absentCells = screen.getByText('Absent Series').closest('tr')!.querySelectorAll('td');
    expect(absentCells[3].querySelector('.badge')!.className).toContain('badge-failed');
  });

  it('renders Suivi and Fichier as two compact status dots on mobile cards', () => {
    vi.mocked(useIsMobile).mockReturnValue(true);
    const movies: RadarrMovieListItem[] = [
      { radarr_id: 1, title: 'Mobile Movie', year: 2020, monitored: true, file_state: 'none', matched: false, occurrence_count: 0 },
    ];
    renderTab({ filmsItems: movies });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    const card = screen.getByText('Mobile Movie').closest('.mobile-list-card')!;
    const dots = card.querySelectorAll('.status-dot');
    expect(dots).toHaveLength(2);
    expect(dots[0]).toHaveAttribute('aria-label', 'Monitored');
    expect(dots[1]).toHaveAttribute('aria-label', 'Absent');
    expect(dots[1].className).toContain('badge-failed');
  });

  it('shows Suivi and Fichier in the sidepanel header and per episode row', async () => {
    vi.mocked(api.getSonarrSeriesEpisodes).mockResolvedValue({
      episodes: [
        { season: 1, episode: 1, monitored: true, file_state: 'downloaded', matched: true, occurrences: [] },
        { season: 1, episode: 2, monitored: true, file_state: 'none', matched: false, occurrences: [] },
      ],
    });
    const partialSeries: SonarrSeriesListItem = { ...seriesFixture, file_state: 'partial', episode_file_count: 1, episode_count: 2 };
    renderTab({ seriesItems: [partialSeries] });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });
    fireEvent.click(screen.getByText('Example Series'));

    await waitFor(() => expect(screen.getByText('Season 1')).toBeInTheDocument());
    // Header carries the series ratio badge (the table row behind the drawer also has one).
    expect(screen.getAllByText('Partial 1/2').length).toBeGreaterThanOrEqual(2);

    fireEvent.click(screen.getByText('Season 1'));
    const downloadedRow = screen.getByText('S01 E01').closest('tr')!;
    expect(downloadedRow.querySelector('.badge-success')?.textContent).toBe('Downloaded');
    const absentRow = screen.getByText('S01 E02').closest('tr')!;
    expect(absentRow.querySelector('.badge-failed')?.textContent).toBe('Absent');
  });
});

describe('RadarrSonarrTab search', () => {
  it('debounces the Films search input before calling setFilmsSearch', () => {
    vi.useFakeTimers();
    const setFilmsSearch = vi.fn();
    renderTab({ setFilmsSearch });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

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
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    expect(screen.getByText('No movies match this search')).toBeInTheDocument();
    expect(screen.queryByText('No monitored movies found')).not.toBeInTheDocument();
  });

  it('shows the default empty state when there is no search term', () => {
    renderTab({ filmsItems: [] });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    expect(screen.getByText('No monitored movies found')).toBeInTheDocument();
  });
});

const seriesEpisodesFixture: SonarrSeriesEpisodesResponse = {
  episodes: [
    { season: 1, episode: 1, monitored: true, file_state: 'downloaded', matched: true, occurrences: [] },
    { season: 1, episode: 2, monitored: true, file_state: 'downloaded', matched: false, occurrences: [] },
    { season: 2, episode: 1, monitored: true, file_state: 'downloaded', matched: true, occurrences: [{ id: 101, resolution: '1080p', state: 'downloaded' }] },
    { season: 2, episode: 2, monitored: true, file_state: 'downloaded', matched: true, occurrences: [] },
  ],
};

const seriesFixture: SonarrSeriesListItem = { sonarr_id: 1, title: 'Example Series', year: 2019, monitored: true, file_state: 'downloaded', episode_file_count: 4, episode_count: 4, matched_count: 3, monitored_count: 4, occurrence_count: 0 };

function openExampleSeries(overrides: Overrides = {}) {
  renderTab({ seriesItems: [seriesFixture], ...overrides });
  fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });
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

    const tabsList = screen.getByRole('tab', { name: 'Radarr' }).closest('.segmented-tabs-list');
    expect(tabsList).toHaveClass('radarr-sonarr-subtabs');

    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });
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
    const movie: RadarrMovieListItem = { radarr_id: 1, title: 'Example Movie', year: 2020, monitored: true, file_state: 'downloaded', matched: true, occurrence_count: 0 };
    renderTab({ filmsItems: [movie] });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });
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

describe('RadarrSonarrTab filter bar reset', () => {
  it('Films reset calls only the Films handler; the search text is kept', () => {
    const resetFilmsFilters = vi.fn();
    const resetSeriesFilters = vi.fn();
    renderTab({ resetFilmsFilters, resetSeriesFilters, filmsFilter: 'matched', filmsSearch: 'matrix' });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });

    fireEvent.click(screen.getByRole('button', { name: 'Reset' }));

    expect(resetFilmsFilters).toHaveBeenCalledTimes(1);
    expect(resetSeriesFilters).not.toHaveBeenCalled();
    expect(screen.getByDisplayValue('matrix')).toBeInTheDocument();
  });

  it('Séries reset calls only the Séries handler', () => {
    const resetFilmsFilters = vi.fn();
    const resetSeriesFilters = vi.fn();
    renderTab({ resetFilmsFilters, resetSeriesFilters, seriesFile: new Set(['none']) });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Sonarr' }), { button: 0 });

    fireEvent.click(screen.getByRole('button', { name: 'Reset' }));

    expect(resetSeriesFilters).toHaveBeenCalledTimes(1);
    expect(resetFilmsFilters).not.toHaveBeenCalled();
  });

  it('reset is disabled at the default filters and enabled once one differs', () => {
    renderTab();
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });
    expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
    cleanup();

    renderTab({ filmsStatus: new Set(['monitored', 'unmonitored']) });
    fireEvent.mouseDown(screen.getByRole('tab', { name: 'Radarr' }), { button: 0 });
    expect(screen.getByRole('button', { name: 'Reset' })).toBeEnabled();
  });
});
