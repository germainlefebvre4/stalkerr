import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { ErrorsSidepanel } from './ErrorsSidepanel';
import { DownloadEnriched } from '../types';
import { api } from '../services/api';

vi.mock('../services/api', () => ({
  api: {
    resolveDuplicateDownload: vi.fn(),
  },
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

function renderSidepanel(item: DownloadEnriched | null) {
  return render(
    <I18nextProvider i18n={i18n}>
      <ErrorsSidepanel item={item} onOpenChange={vi.fn()} />
    </I18nextProvider>
  );
}

const baseDownload: DownloadEnriched = {
  id: 1,
  url: 'http://example.com/movie.mkv',
  status: 'completed',
  retry_count: 0,
  updated_at: '2026-08-27T10:00:00Z',
  completed_at: '2026-08-27T10:05:00Z',
  content: { type: 'movies', title: 'Diagnostic Movie', year: 2020, genres: 'Action' },
  file_info: {
    extension: '.xyz',
    folder_name: 'Diagnostic Movie Folder',
    file_name: 'diagnostic.xyz',
    has_year_in_path: false,
    year_mismatch: false,
    detected_year: 2019,
    is_valid_format: false,
  },
};

describe('ErrorsSidepanel', () => {
  it('renders the diagnostic section before file, content, and status detail', () => {
    renderSidepanel(baseDownload);

    const container = screen.getByText(/Error Diagnostic/).closest('[role="dialog"]') as HTMLElement;
    const html = container.innerHTML;

    const diagnosticIndex = html.indexOf('Diagnostic');
    const fileIndex = html.indexOf('File');
    const contentIndex = html.indexOf('Matched Content');
    const statusIndex = html.indexOf('Status &amp; Dates');

    expect(diagnosticIndex).toBeGreaterThan(-1);
    expect(diagnosticIndex).toBeLessThan(fileIndex);
    expect(fileIndex).toBeLessThan(contentIndex);
    expect(contentIndex).toBeLessThan(statusIndex);
  });

  it('does not render any corrective action controls, regardless of status', () => {
    for (const status of ['completed', 'downloading', 'failed', 'pending', 'retrying'] as const) {
      cleanup();
      renderSidepanel({ ...baseDownload, status });

      expect(screen.queryByText(/Déplacer/)).not.toBeInTheDocument();
      expect(screen.queryByText(/Renommer/)).not.toBeInTheDocument();
      expect(screen.queryByText(/Associer/)).not.toBeInTheDocument();
      expect(screen.queryByText(/Forcer le téléchargement/)).not.toBeInTheDocument();
      expect(screen.queryByText('Move ⇄')).not.toBeInTheDocument();
      expect(screen.queryByText('Rename ✎')).not.toBeInTheDocument();
      expect(screen.queryByText(/Resynchroniser/)).not.toBeInTheDocument();
      expect(screen.queryByText('Resync ⟲')).not.toBeInTheDocument();
    }
  });

  it('shows the detected and expected year in the diagnostic section for a year-related reason', () => {
    renderSidepanel(baseDownload);

    expect(screen.getByText(/Detected: 2019/)).toBeInTheDocument();
    expect(screen.getByText(/Expected: 2020/)).toBeInTheDocument();
  });
});

const duplicateDownload: DownloadEnriched = {
  ...baseDownload,
  file_info: {
    extension: '.mkv',
    folder_name: 'Duplicate Movie Folder',
    file_name: 'duplicate.mkv',
    has_year_in_path: true,
    year_mismatch: false,
    is_valid_format: true,
  },
  duplicate_siblings: [
    { download_info_id: 2, download_path: '/media/movies/Duplicate Movie (1)/file.mkv', completed_at: '2026-08-20T10:00:00Z' },
  ],
};

describe('ErrorsSidepanel duplicate resolve action', () => {
  it('shows the sibling occurrence(s) in the diagnostic section for a duplicate_download item', () => {
    renderSidepanel(duplicateDownload);

    expect(screen.getByText(/\/media\/movies\/Duplicate Movie \(1\)\/file\.mkv/)).toBeInTheDocument();
  });

  it('renders the resolve action only for a duplicate_download item, not for the other three reasons', () => {
    renderSidepanel(duplicateDownload);
    expect(screen.getByText('Resolve duplicate')).toBeInTheDocument();
    cleanup();

    renderSidepanel(baseDownload); // missing_year + unknown_format only
    expect(screen.queryByText('Resolve duplicate')).not.toBeInTheDocument();
  });

  it('shows the per-occurrence outcome after the resolve action completes', async () => {
    vi.mocked(api.resolveDuplicateDownload).mockResolvedValue({
      occurrences: [
        { download_info_id: 1, download_path: '/media/movies/Duplicate Movie/file.mkv', outcome: 'cancelled' },
        { download_info_id: 2, download_path: '/media/movies/Duplicate Movie (1)/file.mkv', outcome: 'unchanged' },
      ],
    });

    renderSidepanel(duplicateDownload);

    fireEvent.click(screen.getByText('Resolve duplicate'));

    await waitFor(() => {
      expect(screen.getByText('Cancelled (file missing)')).toBeInTheDocument();
      expect(screen.getByText('Unchanged (file present)')).toBeInTheDocument();
    });
    expect(api.resolveDuplicateDownload).toHaveBeenCalledWith(duplicateDownload.id);
  });
});
