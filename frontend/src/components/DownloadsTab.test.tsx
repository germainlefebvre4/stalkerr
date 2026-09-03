import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { DownloadsTab } from './DownloadsTab';
import { DownloadEnriched } from '../types';

afterEach(cleanup);

interface PaginationOverrides {
  downloadsTotal?: number;
  downloadsPage?: number;
  setDownloadsPage?: (page: number | ((prev: number) => number)) => void;
  downloadsLimit?: number;
  setDownloadsLimit?: (limit: number) => void;
  onResyncPath?: (item: DownloadEnriched) => Promise<void>;
}

function renderDownloadsTab(downloads: DownloadEnriched[], overrides: PaginationOverrides = {}) {
  return render(
    <I18nextProvider i18n={i18n}>
      <Tabs.Root value="downloads">
        <DownloadsTab
          downloads={downloads}
          downloadsLoading={false}
          statusFilter=""
          setStatusFilter={() => {}}
          typeFilter=""
          setTypeFilter={() => {}}
          problemFilter=""
          setProblemFilter={() => {}}
          downloadsTotal={overrides.downloadsTotal ?? downloads.length}
          downloadsPage={overrides.downloadsPage ?? 1}
          setDownloadsPage={overrides.setDownloadsPage ?? (() => {})}
          downloadsLimit={overrides.downloadsLimit ?? 20}
          setDownloadsLimit={overrides.setDownloadsLimit ?? (() => {})}
          onFetchDownloads={() => {}}
          onOpenMoveDialog={() => {}}
          onOpenRenameDialog={() => {}}
          onResyncPath={overrides.onResyncPath ?? (() => Promise.resolve())}
        />
      </Tabs.Root>
    </I18nextProvider>
  );
}

function openDrawer(download: DownloadEnriched) {
  const title = download.content?.title || download.url;
  fireEvent.click(screen.getByText(new RegExp(title.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))));
}

const baseDownload: DownloadEnriched = {
  id: 1,
  url: 'http://example.com/movie.mkv',
  status: 'completed',
  retry_count: 1,
  updated_at: '2026-08-27T10:00:00Z',
};

describe('DownloadsTab error banner', () => {
  it('does not show the error banner for a completed download with a leftover error_message', () => {
    const download = {
      ...baseDownload,
      status: 'completed' as const,
      error_message: 'HTTP 429: rate limited',
    };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.queryByText(/HTTP 429: rate limited/)).not.toBeInTheDocument();
  });

  it('shows the error banner for a failed download with an error_message', () => {
    const download = {
      ...baseDownload,
      status: 'failed' as const,
      error_message: 'HTTP 429: rate limited',
    };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText(/HTTP 429: rate limited/)).toBeInTheDocument();
  });
});

describe('DownloadsTab sidepanel', () => {
  it('opens the drawer on row click and closes it via the close button', () => {
    const download = { ...baseDownload, content: { type: 'movies' as const, title: 'Test Movie' } };
    renderDownloadsTab([download]);

    expect(screen.queryByText('Download Details')).not.toBeInTheDocument();

    openDrawer(download);
    expect(screen.getByText('Download Details')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Close ✕'));
    expect(screen.queryByText('Download Details')).not.toBeInTheDocument();
  });

  it('does not show Move/Rename/Resync actions for a non-completed download', () => {
    const download = {
      ...baseDownload,
      status: 'downloading' as const,
      content: { type: 'movies' as const, title: 'In Progress Movie' },
      total_bytes: 100,
      bytes_downloaded: 20,
    };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.queryByText('Move ⇄')).not.toBeInTheDocument();
    expect(screen.queryByText('Rename ✎')).not.toBeInTheDocument();
    expect(screen.queryByText('Resync ⟲')).not.toBeInTheDocument();
  });

  it('shows Move/Rename/Resync actions in the drawer for a completed download', () => {
    const download = { ...baseDownload, status: 'completed' as const, content: { type: 'movies' as const, title: 'Done Movie' } };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText('Move ⇄')).toBeInTheDocument();
    expect(screen.getByText('Rename ✎')).toBeInTheDocument();
    expect(screen.getByText('Resync ⟲')).toBeInTheDocument();
  });

  it('calls onResyncPath with the selected download when Resync is clicked', () => {
    const download = { ...baseDownload, status: 'completed' as const, content: { type: 'movies' as const, title: 'Done Movie' } };
    const onResyncPath = vi.fn(() => Promise.resolve());
    renderDownloadsTab([download], { onResyncPath });
    openDrawer(download);

    fireEvent.click(screen.getByText('Resync ⟲'));

    expect(onResyncPath).toHaveBeenCalledWith(download);
  });

  it('closes the drawer automatically when the selected item leaves the downloads list', () => {
    const download = { ...baseDownload, content: { type: 'movies' as const, title: 'Vanishing Movie' } };
    const { rerender } = renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText('Download Details')).toBeInTheDocument();

    rerender(
      <I18nextProvider i18n={i18n}>
        <Tabs.Root value="downloads">
          <DownloadsTab
            downloads={[]}
            downloadsLoading={false}
            statusFilter=""
            setStatusFilter={() => {}}
            typeFilter=""
            setTypeFilter={() => {}}
            problemFilter=""
            setProblemFilter={() => {}}
            downloadsTotal={0}
            downloadsPage={1}
            setDownloadsPage={() => {}}
            downloadsLimit={20}
            setDownloadsLimit={() => {}}
            onFetchDownloads={() => {}}
            onOpenMoveDialog={() => {}}
            onOpenRenameDialog={() => {}}
            onResyncPath={() => Promise.resolve()}
          />
        </Tabs.Root>
      </I18nextProvider>
    );

    expect(screen.queryByText('Download Details')).not.toBeInTheDocument();
  });
});

describe('DownloadsTab file section paths', () => {
  it('shows target and staging path for an in-progress download', () => {
    const download = {
      ...baseDownload,
      status: 'downloading' as const,
      content: { type: 'movies' as const, title: 'In Progress Movie' },
      target_path: '/media/movies/In.Progress.2024/movie.mkv',
      staging_path: '/tmp/stalkeer-download-abc/download.tmp',
    };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText(/\/media\/movies\/In\.Progress\.2024\/movie\.mkv/)).toBeInTheDocument();
    expect(screen.getByText(/\/tmp\/stalkeer-download-abc\/download\.tmp/)).toBeInTheDocument();
    expect(screen.queryByText(download.url)).not.toBeInTheDocument();
  });

  it('shows target and staging path for a failed download', () => {
    const download = {
      ...baseDownload,
      status: 'failed' as const,
      content: { type: 'movies' as const, title: 'Failed Movie' },
      target_path: '/media/movies/Failed.2024/movie.mkv',
      staging_path: '/tmp/stalkeer-download-def/download.tmp',
    };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText(/\/media\/movies\/Failed\.2024\/movie\.mkv/)).toBeInTheDocument();
    expect(screen.getByText(/\/tmp\/stalkeer-download-def\/download\.tmp/)).toBeInTheDocument();
    expect(screen.queryByText(download.url)).not.toBeInTheDocument();
  });

  it('shows only download_path for a completed download, even if target/staging are somehow present', () => {
    const download = {
      ...baseDownload,
      status: 'completed' as const,
      content: { type: 'movies' as const, title: 'Done Movie' },
      download_path: '/media/movies/Done.2024/movie.mkv',
      target_path: '/media/movies/Done.2024/movie.mkv',
      staging_path: '/tmp/stalkeer-download-ghi/download.tmp',
    };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText('/media/movies/Done.2024/movie.mkv')).toBeInTheDocument();
    expect(screen.queryByText(/\/tmp\/stalkeer-download-ghi\/download\.tmp/)).not.toBeInTheDocument();
  });

  it('prefers target/staging path over a pre-filled download_path on a non-completed download (force-download resume prefill)', () => {
    // force-download persists an extensionless download_path up front (as a
    // resume-target base path) before the transfer even starts, so its mere
    // presence must not be read as "the file is done".
    const download = {
      ...baseDownload,
      status: 'downloading' as const,
      content: { type: 'movies' as const, title: 'Prefilled Movie' },
      download_path: '/downloads/radarr/Prefilled Movie (2024)/Prefilled Movie (2024) [1080p]',
      target_path: '/downloads/radarr/Prefilled Movie (2024)/Prefilled Movie (2024) [1080p].mkv',
      staging_path: '/tmp/stalkeer-download-xyz/download.tmp',
    };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText(/Prefilled Movie \(2024\) \[1080p\]\.mkv/)).toBeInTheDocument();
    expect(screen.queryByText('/downloads/radarr/Prefilled Movie (2024)/Prefilled Movie (2024) [1080p]')).not.toBeInTheDocument();
  });

  it('falls back to the url when none of download_path, target_path, or staging_path are present', () => {
    const download = {
      ...baseDownload,
      status: 'pending' as const,
      content: { type: 'movies' as const, title: 'Pending Movie' },
    };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText(download.url)).toBeInTheDocument();
  });
});

describe('DownloadsTab pagination', () => {
  it('renders pagination controls below the list when there are downloads', () => {
    renderDownloadsTab([], { downloadsTotal: 45, downloadsPage: 1, downloadsLimit: 20 });

    expect(screen.getByTitle('Next page')).toBeInTheDocument();
    expect(screen.getByDisplayValue('20')).toBeInTheDocument();
  });

  it('does not render pagination controls when there are no downloads', () => {
    renderDownloadsTab([], { downloadsTotal: 0 });

    expect(screen.queryByTitle('Next page')).not.toBeInTheDocument();
  });

  it('navigating to another page calls setDownloadsPage with the target page', () => {
    const setDownloadsPage = vi.fn();
    renderDownloadsTab([], { downloadsTotal: 45, downloadsPage: 1, downloadsLimit: 20, setDownloadsPage });

    fireEvent.click(screen.getByRole('button', { name: '2' }));

    expect(setDownloadsPage).toHaveBeenCalledWith(2);
  });

  it('changing the items-per-page value calls setDownloadsLimit with the new limit', () => {
    const setDownloadsLimit = vi.fn();
    renderDownloadsTab([], { downloadsTotal: 45, downloadsPage: 1, downloadsLimit: 20, setDownloadsLimit });

    fireEvent.change(screen.getByDisplayValue('20'), { target: { value: '50' } });

    expect(setDownloadsLimit).toHaveBeenCalledWith(50);
  });
});
