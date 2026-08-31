import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { DownloadsSummaryList } from './DownloadsSummaryList';
import { DownloadEnriched } from '../types';

function renderList(downloads: DownloadEnriched[], onRowClick = vi.fn()) {
  render(
    <I18nextProvider i18n={i18n}>
      <DownloadsSummaryList downloads={downloads} loading={false} onRowClick={onRowClick} />
    </I18nextProvider>
  );
  return { onRowClick };
}

const baseDownload: DownloadEnriched = {
  id: 1,
  url: 'http://example.com/movie.mkv',
  status: 'completed',
  retry_count: 0,
  updated_at: '2026-08-27T10:00:00Z',
};

describe('DownloadsSummaryList', () => {
  const originalMatchMedia = window.matchMedia;

  afterEach(() => {
    cleanup();
    window.matchMedia = originalMatchMedia;
  });

  function setMatchMedia(matches: boolean) {
    window.matchMedia = vi.fn().mockImplementation((query: string) => ({
      matches,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })) as unknown as typeof window.matchMedia;
  }

  it('falls back to the file name when content.title is absent', () => {
    setMatchMedia(false);
    renderList([
      { ...baseDownload, download_path: '/data/downloads/Some.Movie.2020.mkv' },
    ]);

    expect(screen.getByText(/Some\.Movie\.2020\.mkv/)).toBeInTheDocument();
  });

  it('only shows a progress indicator for downloading/retrying items', () => {
    setMatchMedia(false);
    renderList([
      { ...baseDownload, id: 1, status: 'completed' },
      { ...baseDownload, id: 2, status: 'downloading', total_bytes: 100, bytes_downloaded: 50 },
    ]);

    expect(screen.getAllByText(/50%/)).toHaveLength(1);
  });

  it('fires onRowClick with the clicked item on desktop', () => {
    setMatchMedia(false);
    const item: DownloadEnriched = { ...baseDownload, id: 42, content: { type: 'movies', title: 'Test Movie' } };
    const { onRowClick } = renderList([item]);

    fireEvent.click(screen.getByText(/Test Movie/));

    expect(onRowClick).toHaveBeenCalledWith(item);
    expect(onRowClick.mock.calls[0][0].id).toBe(42);
  });

  it('renders mobile-list-card markup at a mobile viewport', () => {
    setMatchMedia(true);
    renderList([
      { ...baseDownload, content: { type: 'movies', title: 'Mobile Movie' } },
    ]);

    expect(document.querySelector('.mobile-list-card')).toBeInTheDocument();
    expect(screen.getByText(/Mobile Movie/)).toBeInTheDocument();
  });

  it('fires onRowClick with the clicked item on mobile', () => {
    setMatchMedia(true);
    const item: DownloadEnriched = { ...baseDownload, id: 7, content: { type: 'movies', title: 'Mobile Movie' } };
    const { onRowClick } = renderList([item]);

    fireEvent.click(screen.getByText(/Mobile Movie/));

    expect(onRowClick).toHaveBeenCalledWith(item);
  });
});
