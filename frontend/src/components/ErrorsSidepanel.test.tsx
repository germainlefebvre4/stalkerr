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
    getDuplicateCandidates: vi.fn(),
    keepDuplicateDownload: vi.fn(),
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

describe('ErrorsSidepanel keep-one duplicate action', () => {
  const candidates = {
    candidates: [
      { download_path: '/media/a/Star Wars (1977).mkv', exists: true, size_bytes: 2048, download_info_ids: [1], matches_library_name: true, suggested: true, completed_at: '2026-08-20T10:00:00Z' },
      { download_path: '/media/b/La Guerre des etoiles.mp4', exists: true, size_bytes: 1024, download_info_ids: [2], matches_library_name: false, suggested: false },
      { download_path: '/media/c/gone.mkv', exists: false, download_info_ids: [3], suggested: false },
    ],
  };

  async function openPicker() {
    vi.mocked(api.getDuplicateCandidates).mockResolvedValue(candidates);
    renderSidepanel(duplicateDownload);
    fireEvent.click(screen.getByText('Keep one, delete the others'));
    await waitFor(() => expect(screen.getByText('Choose the file to keep:')).toBeInTheDocument());
  }

  it('offers the action only for duplicate_download items', () => {
    renderSidepanel(baseDownload);
    expect(screen.queryByText('Keep one, delete the others')).not.toBeInTheDocument();
    expect(screen.queryByText('Resolve duplicate')).not.toBeInTheDocument();
  });

  it('pre-selects the suggested file, allows changing it, and disables a missing file', async () => {
    await openPicker();
    const radios = screen.getAllByRole('radio') as HTMLInputElement[];
    expect(radios[0].checked).toBe(true);
    expect(radios[2].disabled).toBe(true);
    expect(screen.getByText('Missing')).toBeInTheDocument();

    fireEvent.click(radios[1]);
    expect(radios[1].checked).toBe(true);
  });

  it('pre-selects nothing when no file is suggested', async () => {
    vi.mocked(api.getDuplicateCandidates).mockResolvedValue({
      candidates: candidates.candidates.map(c => ({ ...c, suggested: false })),
    });
    renderSidepanel(duplicateDownload);
    fireEvent.click(screen.getByText('Keep one, delete the others'));
    await waitFor(() => expect(screen.getByText('Choose the file to keep:')).toBeInTheDocument());

    (screen.getAllByRole('radio') as HTMLInputElement[]).forEach(r => expect(r.checked).toBe(false));
    expect((screen.getByText('Continue') as HTMLButtonElement).disabled).toBe(true);
  });

  it('sends no modifying request when the confirmation is dismissed', async () => {
    await openPicker();
    fireEvent.click(screen.getByText('Continue'));
    expect(screen.getByText('Confirm deletion')).toBeInTheDocument();
    expect(screen.getAllByText('/media/a/Star Wars (1977).mkv').length).toBeGreaterThan(0);
    expect(screen.getByText('/media/b/La Guerre des etoiles.mp4')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Cancel'));
    expect(api.keepDuplicateDownload).not.toHaveBeenCalled();
  });

  it('confirms, sends the selected occurrence id, and renders outcomes with the refresh state', async () => {
    vi.mocked(api.keepDuplicateDownload).mockResolvedValue({
      occurrences: [
        { download_info_id: 1, download_path: '/media/a/Star Wars (1977).mkv', outcome: 'kept' },
        { download_info_id: 2, download_path: '/media/b/La Guerre des etoiles.mp4', outcome: 'deleted' },
        { download_info_id: 3, download_path: '/media/c/gone.mkv', outcome: 'cancelled_file_missing' },
      ],
      refresh: 'requested',
    });
    await openPicker();
    fireEvent.click(screen.getByText('Continue'));
    fireEvent.click(screen.getByText('Delete and keep the selected file'));

    await waitFor(() => expect(screen.getByText('Kept')).toBeInTheDocument());
    expect(api.keepDuplicateDownload).toHaveBeenCalledWith(1);
    expect(screen.getByText('Deleted')).toBeInTheDocument();
    expect(screen.getByText('Cancelled (file already missing)')).toBeInTheDocument();
    expect(screen.getByText('Radarr/Sonarr refresh requested.')).toBeInTheDocument();
  });

  it('shows a distinct message when the library refresh failed', async () => {
    vi.mocked(api.keepDuplicateDownload).mockResolvedValue({
      occurrences: [{ download_info_id: 1, download_path: '/media/a/Star Wars (1977).mkv', outcome: 'kept' }],
      refresh: 'failed',
    });
    await openPicker();
    fireEvent.click(screen.getByText('Continue'));
    fireEvent.click(screen.getByText('Delete and keep the selected file'));

    await waitFor(() => expect(screen.getByText(/the library was not refreshed/)).toBeInTheDocument());
  });

  it('refreshes the list on close after a run', async () => {
    vi.mocked(api.keepDuplicateDownload).mockResolvedValue({
      occurrences: [{ download_info_id: 1, download_path: '/media/a/Star Wars (1977).mkv', outcome: 'kept' }],
      refresh: 'requested',
    });
    const onChanged = vi.fn();
    vi.mocked(api.getDuplicateCandidates).mockResolvedValue(candidates);
    render(
      <I18nextProvider i18n={i18n}>
        <ErrorsSidepanel item={duplicateDownload} onOpenChange={vi.fn()} onChanged={onChanged} />
      </I18nextProvider>
    );
    fireEvent.click(screen.getByText('Keep one, delete the others'));
    await waitFor(() => expect(screen.getByText('Continue')).toBeInTheDocument());
    fireEvent.click(screen.getByText('Continue'));
    fireEvent.click(screen.getByText('Delete and keep the selected file'));
    await waitFor(() => expect(screen.getByText('Kept')).toBeInTheDocument());
    expect(onChanged).not.toHaveBeenCalled();

    fireEvent.click(screen.getByText('Close ✕'));
    expect(onChanged).toHaveBeenCalledTimes(1);
  });
});

describe('errors locale parity', () => {
  it('en and fr errors.json expose the same keys', async () => {
    const en = (await import('../locales/en/errors.json')).default;
    const fr = (await import('../locales/fr/errors.json')).default;
    const keys = (o: object, prefix = ''): string[] =>
      Object.entries(o).flatMap(([k, v]) => (typeof v === 'object' && v !== null ? keys(v, `${prefix}${k}.`) : [`${prefix}${k}`])).sort();
    expect(keys(fr)).toEqual(keys(en));
  });
});

