import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import * as Tabs from '@radix-ui/react-tabs';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { DownloadsTab } from './DownloadsTab';
import { DownloadEnriched } from '../types';

function renderDownloadsTab(download: DownloadEnriched) {
  return render(
    <I18nextProvider i18n={i18n}>
      <Tabs.Root value="downloads">
        <DownloadsTab
          downloads={[download]}
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

const baseDownload: DownloadEnriched = {
  id: 1,
  url: 'http://example.com/movie.mkv',
  status: 'completed',
  retry_count: 1,
  updated_at: '2026-08-27T10:00:00Z',
};

describe('DownloadsTab error banner', () => {
  it('does not show the error banner for a completed download with a leftover error_message', () => {
    renderDownloadsTab({
      ...baseDownload,
      status: 'completed',
      error_message: 'HTTP 429: rate limited',
    });

    expect(screen.queryByText(/HTTP 429: rate limited/)).not.toBeInTheDocument();
  });

  it('shows the error banner for a failed download with an error_message', () => {
    renderDownloadsTab({
      ...baseDownload,
      status: 'failed',
      error_message: 'HTTP 429: rate limited',
    });

    expect(screen.getByText(/HTTP 429: rate limited/)).toBeInTheDocument();
  });
});
