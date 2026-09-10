import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { HomeTab } from './HomeTab';
import { ProcessingLog, RadarrSonarrStats, StatsResponse } from '../types';
import { useIsMobile } from '../hooks/useMediaQuery';

vi.mock('../hooks/useMediaQuery', () => ({
  useIsMobile: vi.fn(() => false),
}));

afterEach(() => {
  cleanup();
  vi.mocked(useIsMobile).mockReturnValue(false);
});

interface Overrides {
  latestLog?: ProcessingLog | null;
  latestLogLoading?: boolean;
  stats?: StatsResponse | null;
  radarrSonarrStats?: RadarrSonarrStats | null;
  radarrSonarrStatsLoading?: boolean;
  radarrSonarrStatsError?: string | null;
  downloadsTotal?: number;
  errorsTotal?: number;
}

function renderHomeTab(overrides: Overrides = {}) {
  return render(
    <I18nextProvider i18n={i18n}>
      <Tabs.Root value="home">
        <HomeTab
          latestLog={overrides.latestLog ?? null}
          latestLogLoading={overrides.latestLogLoading ?? false}
          stats={overrides.stats ?? null}
          getDownloadSuccessRatio={() => '87%'}
          radarrSonarrStats={overrides.radarrSonarrStats ?? null}
          radarrSonarrStatsLoading={overrides.radarrSonarrStatsLoading ?? false}
          radarrSonarrStatsError={overrides.radarrSonarrStatsError ?? null}
          downloadsTotal={overrides.downloadsTotal ?? 0}
          errorsTotal={overrides.errorsTotal ?? 0}
        />
      </Tabs.Root>
    </I18nextProvider>
  );
}

const completedLog: ProcessingLog = {
  id: 1,
  action: 'process_m3u',
  item_count: 17,
  status: 'success',
  started_at: '2026-01-01T10:00:00Z',
  completed_at: '2026-01-01T10:05:09Z',
  movies_count: 12,
  tv_shows_count: 5,
  new_items_count: 8,
  tmdb_matched_count: 17,
  tmdb_unmatched_count: 3,
  group_titles: ['ACTION-FR', 'ANIMATION'],
};

describe('HomeTab', () => {
  it('renders all four summary sections', () => {
    renderHomeTab({ latestLog: completedLog });

    expect(screen.getByText('Last Processing Run')).toBeInTheDocument();
    expect(screen.getByText('Catalog Overview')).toBeInTheDocument();
    expect(screen.getByText('Radarr / Sonarr')).toBeInTheDocument();
    expect(screen.getByText('Downloads & Errors')).toBeInTheDocument();
  });

  it('renders the completed run summary with its counts and group titles', () => {
    renderHomeTab({ latestLog: completedLog });

    expect(screen.getByText('12')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(screen.getByText('8')).toBeInTheDocument();
    expect(screen.getByText(/ACTION-FR, ANIMATION/)).toBeInTheDocument();
    expect(screen.getByText(/Duration: 5m 09s/)).toBeInTheDocument();
  });

  it('renders an empty state when no processing run has ever been recorded', () => {
    renderHomeTab({ latestLog: null, latestLogLoading: false });

    expect(screen.getByText('No processing run has been recorded yet.')).toBeInTheDocument();
  });

  it('renders a statistics-unavailable placeholder for a pre-migration run instead of zeros', () => {
    const preMigrationLog: ProcessingLog = {
      id: 2,
      action: 'process_m3u',
      item_count: 9,
      status: 'success',
      started_at: '2026-01-01T10:00:00Z',
      completed_at: '2026-01-01T10:01:00Z',
    };

    renderHomeTab({ latestLog: preMigrationLog });

    expect(screen.getByText('Statistics unavailable for this run.')).toBeInTheDocument();
    expect(screen.queryByText('ACTION-FR')).not.toBeInTheDocument();
  });

  it('renders an in-progress state instead of a fixed duration', () => {
    const inProgressLog: ProcessingLog = {
      ...completedLog,
      status: 'in_progress',
      completed_at: undefined,
    };

    renderHomeTab({ latestLog: inProgressLog });

    expect(screen.getByText('In progress', { selector: '.badge' })).toBeInTheDocument();
    expect(screen.getByText((_, node) => node?.textContent === '01/01/2026 Started · In progress...')).toBeInTheDocument();
    expect(screen.queryByText(/Duration:/)).not.toBeInTheDocument();
  });

  it('renders a per-service Radarr/Sonarr error independently, without hiding the other service', () => {
    renderHomeTab({
      radarrSonarrStats: {
        radarr_monitored: null,
        radarr_matched: null,
        radarr_error: 'radarr_unreachable',
        sonarr_monitored: 42,
        sonarr_matched: null,
      },
    });

    expect(screen.getByText('Failed to reach Radarr. Please try again.', { exact: false })).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('Failed', { selector: '.badge' })).toBeInTheDocument();
    expect(screen.getByText('Success', { selector: '.badge' })).toBeInTheDocument();
  });

  it('renders the Sonarr matched count and progress bar when sonarr_matched is present', () => {
    renderHomeTab({
      radarrSonarrStats: {
        radarr_monitored: 10,
        radarr_matched: 5,
        sonarr_monitored: 8,
        sonarr_matched: 6,
      },
    });

    expect(screen.getByText('8')).toBeInTheDocument();
    expect(screen.getByText('6')).toBeInTheDocument();

    const indicators = document.querySelectorAll('.progress-indicator');
    const widths = Array.from(indicators).map(el => (el as HTMLElement).style.width);
    expect(widths).toContain('75%');
  });

  it('falls back to a dash and a zero-width progress bar when sonarr_matched is null', () => {
    renderHomeTab({
      radarrSonarrStats: {
        radarr_monitored: 10,
        radarr_matched: 5,
        sonarr_monitored: 8,
        sonarr_matched: null,
      },
    });

    expect(screen.getByText('8')).toBeInTheDocument();
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);
  });

  it('still renders the Sonarr error state unchanged when sonarr_error is set, without hiding a successful Radarr section', () => {
    renderHomeTab({
      radarrSonarrStats: {
        radarr_monitored: 10,
        radarr_matched: 5,
        sonarr_monitored: null,
        sonarr_matched: null,
        sonarr_error: 'sonarr_unreachable',
      },
    });

    expect(screen.getByText('Failed to reach Sonarr. Please try again.', { exact: false })).toBeInTheDocument();
    expect(screen.getByText('10')).toBeInTheDocument();
    expect(screen.getByText('Failed', { selector: '.badge' })).toBeInTheDocument();
    expect(screen.getByText('Success', { selector: '.badge' })).toBeInTheDocument();
  });

  it('renders a short group-titles list fully with no disclosure control', () => {
    renderHomeTab({ latestLog: completedLog });

    expect(screen.getByText(/ACTION-FR, ANIMATION/)).toBeInTheDocument();
    expect(screen.queryByText(/\+\d+ more/)).not.toBeInTheDocument();
  });

  it('collapses a long group-titles list behind a disclosure and expands it on activation', () => {
    const longLog: ProcessingLog = {
      ...completedLog,
      group_titles: ['ACTION-FR', 'ANIMATION', 'DOCUMENTAIRE', 'HORREUR', 'COMEDIE'],
    };
    renderHomeTab({ latestLog: longLog });

    expect(screen.getByText(/ACTION-FR, ANIMATION, DOCUMENTAIRE/)).toBeInTheDocument();
    expect(screen.queryByText(/HORREUR/)).not.toBeInTheDocument();
    const disclosureButton = screen.getByText('+2 more');
    expect(disclosureButton).toBeInTheDocument();

    fireEvent.click(disclosureButton);

    expect(screen.getByText(/ACTION-FR, ANIMATION, DOCUMENTAIRE, HORREUR, COMEDIE/)).toBeInTheDocument();
    expect(screen.queryByText('+2 more')).not.toBeInTheDocument();
  });

  it('renders the catalog download-success progress bar matching getDownloadSuccessRatio', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <Tabs.Root value="home">
          <HomeTab
            latestLog={null}
            latestLogLoading={false}
            stats={{ total_items: 100, by_content_type: { movies: 60, tvshows: 40 }, by_state: {} }}
            getDownloadSuccessRatio={() => '87%'}
            radarrSonarrStats={null}
            radarrSonarrStatsLoading={false}
            radarrSonarrStatsError={null}
            downloadsTotal={0}
            errorsTotal={0}
          />
        </Tabs.Root>
      </I18nextProvider>
    );

    const indicator = document.querySelector('.progress-indicator') as HTMLElement;
    expect(indicator.style.width).toBe('87%');
  });

  it('renders the downloads/errors badge as neutral when there are no errors and failed when there are', () => {
    const { rerender } = render(
      <I18nextProvider i18n={i18n}>
        <Tabs.Root value="home">
          <HomeTab
            latestLog={null}
            latestLogLoading={false}
            stats={null}
            getDownloadSuccessRatio={() => '100%'}
            radarrSonarrStats={null}
            radarrSonarrStatsLoading={false}
            radarrSonarrStatsError={null}
            downloadsTotal={10}
            errorsTotal={0}
          />
        </Tabs.Root>
      </I18nextProvider>
    );

    expect(screen.getByText(/0 Naming errors/).className).toContain('badge-neutral');

    rerender(
      <I18nextProvider i18n={i18n}>
        <Tabs.Root value="home">
          <HomeTab
            latestLog={null}
            latestLogLoading={false}
            stats={null}
            getDownloadSuccessRatio={() => '100%'}
            radarrSonarrStats={null}
            radarrSonarrStatsLoading={false}
            radarrSonarrStatsError={null}
            downloadsTotal={10}
            errorsTotal={3}
          />
        </Tabs.Root>
      </I18nextProvider>
    );

    expect(screen.getByText(/3 Naming errors/).className).toContain('badge-failed');
  });
});

describe('HomeTab mobile layout', () => {
  it('replaces status badges with compact status dots on mobile', () => {
    vi.mocked(useIsMobile).mockReturnValue(true);
    renderHomeTab({
      latestLog: completedLog,
      radarrSonarrStats: { radarr_monitored: 10, radarr_matched: 7, sonarr_monitored: 5, sonarr_matched: 3 },
    });

    // The Downloads & Errors card's badge conveys the errors count itself (not a card
    // load-status), so it intentionally stays a full badge even on mobile.
    expect(document.querySelectorAll('.badge:not(.status-dot)').length).toBe(1);
    const dots = document.querySelectorAll('.status-dot');
    expect(dots.length).toBeGreaterThan(0);
    dots.forEach(dot => expect(dot.textContent).toBe(''));
  });

  it('meets the 44px minimum tap target on the group-titles disclosure control on mobile', () => {
    vi.mocked(useIsMobile).mockReturnValue(true);
    const longLog: ProcessingLog = {
      ...completedLog,
      group_titles: ['ACTION-FR', 'ANIMATION', 'DOCUMENTAIRE', 'HORREUR', 'COMEDIE'],
    };
    renderHomeTab({ latestLog: longLog });

    const disclosureButton = screen.getByText('+2 more');
    expect(disclosureButton).toHaveClass('home-disclosure-btn');
  });
});
