import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { ArrFilterBar } from './ArrFilterBar';

afterEach(cleanup);

function renderBar(overrides: Partial<React.ComponentProps<typeof ArrFilterBar>> = {}) {
  const props: React.ComponentProps<typeof ArrFilterBar> = {
    service: 'radarr',
    searchPlaceholder: 'Search',
    searchValue: '',
    onSearchChange: vi.fn(),
    matchStatus: '',
    onMatchStatusChange: vi.fn(),
    suivi: new Set(['monitored']),
    onSuiviChange: vi.fn(),
    suiviOptions: [{ value: 'monitored', label: 'Monitored' }, { value: 'unmonitored', label: 'Unmonitored' }],
    file: new Set(),
    onFileChange: vi.fn(),
    fileOptions: [{ value: 'downloaded', label: 'Downloaded' }, { value: 'none', label: 'Missing file' }],
    missingActive: false,
    onMissing: vi.fn(),
    isDefault: false,
    onReset: vi.fn(),
    ...overrides,
  };
  render(<I18nextProvider i18n={i18n}><ArrFilterBar {...props} /></I18nextProvider>);
  return props;
}

describe('ArrFilterBar', () => {
  it('shows the service and Playlist group captions and a labelled match-status select', () => {
    renderBar();
    expect(screen.getByRole('group', { name: 'Radarr' })).toBeInTheDocument();
    expect(screen.getByRole('group', { name: 'Playlist' })).toBeInTheDocument();
    expect(screen.getByLabelText('Match')).toBeInTheDocument();
  });

  it('uses the Sonarr caption for the series section', () => {
    renderBar({ service: 'sonarr' });
    expect(screen.getByRole('group', { name: 'Sonarr' })).toBeInTheDocument();
  });

  it('forwards search, match-status, Manquants and reset callbacks', () => {
    const props = renderBar();
    fireEvent.change(screen.getByPlaceholderText('Search'), { target: { value: 'abc' } });
    expect(props.onSearchChange).toHaveBeenCalledWith('abc');
    fireEvent.change(screen.getByLabelText('Match'), { target: { value: 'matched' } });
    expect(props.onMatchStatusChange).toHaveBeenCalledWith('matched');
    fireEvent.click(screen.getByRole('button', { name: 'Missing' }));
    expect(props.onMissing).toHaveBeenCalledTimes(1);
    fireEvent.click(screen.getByRole('button', { name: 'Reset' }));
    expect(props.onReset).toHaveBeenCalledTimes(1);
  });

  it('reflects the active Manquants state', () => {
    renderBar({ missingActive: true });
    expect(screen.getByRole('button', { name: 'Missing' })).toHaveAttribute('aria-pressed', 'true');
  });

  it('disables reset at defaults', () => {
    renderBar({ isDefault: true });
    expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
  });

  it('renders the Suivi and Fichier triggers', () => {
    renderBar();
    expect(screen.getByRole('button', { name: /Monitoring: Monitored/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /File: /i })).toBeInTheDocument();
  });
});
