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

    expect(screen.queryByText('📥 Download Details')).not.toBeInTheDocument();

    openDrawer(download);
    expect(screen.getByText('📥 Download Details')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Close ✕'));
    expect(screen.queryByText('📥 Download Details')).not.toBeInTheDocument();
  });

  it('does not show Move/Rename actions for a non-completed download', () => {
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
  });

  it('shows Move/Rename actions in the drawer for a completed download', () => {
    const download = { ...baseDownload, status: 'completed' as const, content: { type: 'movies' as const, title: 'Done Movie' } };
    renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText('Move ⇄')).toBeInTheDocument();
    expect(screen.getByText('Rename ✎')).toBeInTheDocument();
  });

  it('closes the drawer automatically when the selected item leaves the downloads list', () => {
    const download = { ...baseDownload, content: { type: 'movies' as const, title: 'Vanishing Movie' } };
    const { rerender } = renderDownloadsTab([download]);
    openDrawer(download);

    expect(screen.getByText('📥 Download Details')).toBeInTheDocument();

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
          />
        </Tabs.Root>
      </I18nextProvider>
    );

    expect(screen.queryByText('📥 Download Details')).not.toBeInTheDocument();
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
