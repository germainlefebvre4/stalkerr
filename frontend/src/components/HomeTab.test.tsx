import { afterEach, describe, expect, it } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { HomeTab } from './HomeTab';
import { ProcessingLog, RadarrSonarrStats, StatsResponse } from '../types';

afterEach(cleanup);

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

    expect(screen.getByText((_, node) => node?.textContent === 'In progress - In progress...')).toBeInTheDocument();
    expect(screen.queryByText(/Duration:/)).not.toBeInTheDocument();
  });

  it('renders a per-service Radarr/Sonarr error independently, without hiding the other service', () => {
    renderHomeTab({
      radarrSonarrStats: {
        radarr_monitored: null,
        radarr_matched: null,
        radarr_error: 'radarr_unreachable',
        sonarr_monitored: 42,
      },
    });

    expect(screen.getByText('Failed to reach Radarr. Please try again.', { exact: false })).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
  });
});
