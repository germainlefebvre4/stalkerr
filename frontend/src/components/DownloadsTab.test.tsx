import { afterEach, describe, expect, it } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { DownloadsTab } from './DownloadsTab';
import { DownloadEnriched } from '../types';

afterEach(cleanup);

function renderDownloadsTab(downloads: DownloadEnriched[]) {
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
