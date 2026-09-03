import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { ErrorsTable } from './ErrorsTable';
import { DownloadEnriched } from '../types';

afterEach(cleanup);

function renderTable(downloads: DownloadEnriched[], onRowClick = vi.fn()) {
  render(
    <I18nextProvider i18n={i18n}>
      <ErrorsTable downloads={downloads} loading={false} onRowClick={onRowClick} />
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

describe('ErrorsTable', () => {
  it('renders one badge per applicable reason for an item with multiple problems', () => {
    const download: DownloadEnriched = {
      ...baseDownload,
      content: { type: 'movies', title: 'Multi Problem Movie' },
      file_info: {
        extension: '.xyz',
        folder_name: 'Multi Problem Movie',
        file_name: 'multi.xyz',
        has_year_in_path: false,
        year_mismatch: false,
        is_valid_format: false,
      },
    };
    renderTable([download]);

    expect(screen.getByText(/Missing year/)).toBeInTheDocument();
    expect(screen.getByText(/Invalid extension/)).toBeInTheDocument();
    expect(screen.queryByText(/Inconsistent year/)).not.toBeInTheDocument();
  });

  it('falls back to the file name derived from file_info when content.title is unavailable', () => {
    const download: DownloadEnriched = {
      ...baseDownload,
      file_info: {
        extension: '.mkv',
        folder_name: 'SomeFolder',
        file_name: 'some.file.mkv',
        has_year_in_path: false,
        year_mismatch: false,
        is_valid_format: true,
      },
    };
    renderTable([download]);

    expect(screen.getByText(/some\.file\.mkv/)).toBeInTheDocument();
  });

  it('fires onRowClick with the clicked item', () => {
    const item: DownloadEnriched = {
      ...baseDownload,
      id: 42,
      content: { type: 'movies', title: 'Clickable Movie' },
      file_info: {
        extension: '.mkv',
        folder_name: 'Clickable.Movie.Folder',
        file_name: 'clickable.mkv',
        has_year_in_path: false,
        year_mismatch: false,
        is_valid_format: true,
      },
    };
    const { onRowClick } = renderTable([item]);

    fireEvent.click(screen.getByText(/Clickable Movie/));

    expect(onRowClick).toHaveBeenCalledWith(item);
  });
});
