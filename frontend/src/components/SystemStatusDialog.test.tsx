import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, waitFor } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { SystemStatusDialog } from './SystemStatusDialog';
import { api } from '../services/api';
import { SystemStatusResponse } from '../types';

vi.mock('../services/api', () => ({
  api: {
    getSystemStatus: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    code: string;
    constructor(code: string, message: string) {
      super(message);
      this.code = code;
    }
  },
}));

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  void i18n.changeLanguage('en');
});

const response: SystemStatusResponse = {
  database: { status: 'ok' },
  radarr: { status: 'ko', reason: 'unreachable' },
  sonarr: { status: 'not_configured' },
  tmdb: { status: 'ko', reason: 'unauthorized' },
  disk: [
    { paths: ['movies', 'tvshows'], available: 100, free: 100, total: 200, used_pct: 50 },
  ],
  version: '1.2.3',
  commit: 'abc1234',
  date: '2026-09-10_12:00:00',
};

function renderDialog(isOpen: boolean) {
  return render(
    <I18nextProvider i18n={i18n}>
      <SystemStatusDialog isOpen={isOpen} onOpenChange={() => {}} />
    </I18nextProvider>
  );
}

describe('SystemStatusDialog fetch behavior', () => {
  it('does not fetch before the dialog is opened', () => {
    renderDialog(false);
    expect(api.getSystemStatus).not.toHaveBeenCalled();
  });

  it('fetches exactly once when opened', async () => {
    vi.mocked(api.getSystemStatus).mockResolvedValue(response);
    renderDialog(true);

    await waitFor(() => expect(api.getSystemStatus).toHaveBeenCalledTimes(1));
  });

  it('fetches again on reopen after close', async () => {
    vi.mocked(api.getSystemStatus).mockResolvedValue(response);
    const { rerender } = render(
      <I18nextProvider i18n={i18n}>
        <SystemStatusDialog isOpen={true} onOpenChange={() => {}} />
      </I18nextProvider>
    );
    await waitFor(() => expect(api.getSystemStatus).toHaveBeenCalledTimes(1));

    rerender(
      <I18nextProvider i18n={i18n}>
        <SystemStatusDialog isOpen={false} onOpenChange={() => {}} />
      </I18nextProvider>
    );
    rerender(
      <I18nextProvider i18n={i18n}>
        <SystemStatusDialog isOpen={true} onOpenChange={() => {}} />
      </I18nextProvider>
    );

    await waitFor(() => expect(api.getSystemStatus).toHaveBeenCalledTimes(2));
  });
});

describe('SystemStatusDialog translations', () => {
  it('renders translated English rows/states/reasons, not raw keys', async () => {
    vi.mocked(api.getSystemStatus).mockResolvedValue(response);
    await i18n.changeLanguage('en');
    renderDialog(true);

    await waitFor(() => expect(screen.getByText('System Status')).toBeInTheDocument());
    expect(screen.getByText('Database')).toBeInTheDocument();
    expect(screen.getByText('Unreachable')).toBeInTheDocument();
    expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
    expect(screen.getByText('Not configured')).toBeInTheDocument();
    expect(screen.getByText('Build')).toBeInTheDocument();
    expect(screen.getByText('Version: 1.2.3')).toBeInTheDocument();
    expect(screen.getByText('Commit: abc1234')).toBeInTheDocument();
    expect(screen.getByText('Built: 2026-09-10_12:00:00')).toBeInTheDocument();
    expect(screen.queryByText(/systemStatus\./)).not.toBeInTheDocument();
  });

  it('renders translated French rows/states/reasons, not raw keys', async () => {
    vi.mocked(api.getSystemStatus).mockResolvedValue(response);
    await i18n.changeLanguage('fr');
    renderDialog(true);

    await waitFor(() => expect(screen.getByText('État du système')).toBeInTheDocument());
    expect(screen.getByText('Base de données')).toBeInTheDocument();
    expect(screen.getByText('Injoignable')).toBeInTheDocument();
    expect(screen.getByText('Identifiants invalides')).toBeInTheDocument();
    expect(screen.getByText('Non configuré')).toBeInTheDocument();
    expect(screen.getByText('Version installée')).toBeInTheDocument();
    expect(screen.getByText('Version: 1.2.3')).toBeInTheDocument();
    expect(screen.getByText('Commit: abc1234')).toBeInTheDocument();
    expect(screen.getByText('Construit le: 2026-09-10_12:00:00')).toBeInTheDocument();
    expect(screen.queryByText(/systemStatus\./)).not.toBeInTheDocument();
  });
});
