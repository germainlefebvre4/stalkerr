import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { ErrorsSidepanel } from './ErrorsSidepanel';
import { DownloadEnriched } from '../types';

afterEach(cleanup);

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
    }
  });

  it('shows the detected and expected year in the diagnostic section for a year-related reason', () => {
    renderSidepanel(baseDownload);

    expect(screen.getByText(/Detected: 2019/)).toBeInTheDocument();
    expect(screen.getByText(/Expected: 2020/)).toBeInTheDocument();
  });
});
