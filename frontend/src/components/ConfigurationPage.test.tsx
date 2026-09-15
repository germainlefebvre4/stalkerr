import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { ConfigurationPage, ConfigurationPageProps } from './ConfigurationPage';
import { api } from '../services/api';
import { useIsMobile } from '../hooks/useMediaQuery';
import { FilterConfig, SettingsField, M3uSource } from '../types';

vi.mock('../hooks/useMediaQuery', () => ({
  useIsMobile: vi.fn(() => false),
}));

vi.mock('../services/api', () => ({
  api: {
    getSettings: vi.fn().mockResolvedValue({ settings: [] }),
    setSetting: vi.fn(),
    clearSetting: vi.fn(),
    getM3uSources: vi.fn().mockResolvedValue({ sources: [] }),
    getM3uSourcesOrigin: vi.fn().mockResolvedValue({ sources: [] }),
    getSystemStatus: vi.fn().mockResolvedValue({
      database: { status: 'ok' }, radarr: { status: 'not_configured' }, sonarr: { status: 'not_configured' },
      tmdb: { status: 'not_configured' }, disk: [], version: '0.0.0', commit: 'test', date: 'test',
    }),
  },
  ApiError: class ApiError extends Error {
    code: string;
    constructor(code: string, message: string) {
      super(message);
      this.code = code;
    }
  },
}));

function baseProps(overrides: Partial<ConfigurationPageProps> = {}): ConfigurationPageProps {
  return {
    onBack: vi.fn(),
    theme: 'light', onSetTheme: vi.fn(),
    reduceMotion: false, onSetReduceMotion: vi.fn(),
    startupTab: 'home', onSetStartupTab: vi.fn(), tabOptions: [{ value: 'home', label: 'Home' }],
    playlistLimit: 10, onSetPlaylistLimit: vi.fn(),
    playlistView: 'items', onSetPlaylistView: vi.fn(),
    filters: [], filterOrigin: [], filtersLoading: false, onFetchFilters: vi.fn(),
    onDeleteFilter: vi.fn(), onOpenCreateFilter: vi.fn(),
    ...overrides,
  };
}

function renderPage(overrides: Partial<ConfigurationPageProps> = {}) {
  return render(
    <I18nextProvider i18n={i18n}>
      <ConfigurationPage {...baseProps(overrides)} />
    </I18nextProvider>
  );
}

describe('ConfigurationPage', () => {
  afterEach(() => {
    cleanup();
    localStorage.clear();
    window.history.replaceState(null, '', '/');
    vi.mocked(api.getSettings).mockReset().mockResolvedValue({ settings: [] });
    vi.mocked(api.getM3uSources).mockReset().mockResolvedValue({ sources: [] });
    vi.mocked(api.getM3uSourcesOrigin).mockReset().mockResolvedValue({ sources: [] });
    vi.mocked(useIsMobile).mockReturnValue(false);
  });

  it('renders the six tabs in order', () => {
    renderPage();

    const tabs = screen.getAllByRole('tab').map(el => el.textContent);
    expect(tabs).toEqual(['General', 'Integrations', 'Content', 'Notifications', 'Advanced', 'System']);
  });

  it('defaults to the "General" tab and shows Appearance/Language/Preferences as cards', () => {
    renderPage();

    expect(screen.getByRole('tab', { name: 'General' })).toHaveAttribute('data-state', 'active');
    expect(screen.getByText('Appearance')).toBeInTheDocument();
    expect(screen.getAllByText('Language').length).toBeGreaterThan(0);
    expect(screen.getByText('Preferences')).toBeInTheDocument();
  });

  it('calls onBack when the close button is clicked', () => {
    const onBack = vi.fn();
    renderPage({ onBack });
    fireEvent.click(screen.getByText('Close'));
    expect(onBack).toHaveBeenCalled();
  });

  it('prefetches settings, M3U sources, and filters once on mount, regardless of the initially active tab', async () => {
    localStorage.setItem('stalkeer_configuration_tab', 'system');
    const onFetchFilters = vi.fn();

    renderPage({ onFetchFilters });

    await waitFor(() => {
      expect(api.getSettings).toHaveBeenCalledTimes(1);
      expect(api.getM3uSources).toHaveBeenCalledTimes(1);
      expect(onFetchFilters).toHaveBeenCalledTimes(1);
    });
  });

  describe('summary banner', () => {
    it('states no settings are overridden when nothing carries an override', () => {
      renderPage({ filters: [], filterOrigin: [] });
      expect(screen.getByText('No settings are currently overridden.')).toBeInTheDocument();
    });

    it('shows the total override count across settings, M3U sources, and filters', async () => {
      vi.mocked(api.getSettings).mockResolvedValueOnce({
        settings: [
          { key: 'radarr.url', value: 'http://x', sensitive: false, origin: 'interface', restart_required: false },
        ] as SettingsField[],
      });
      vi.mocked(api.getM3uSources).mockResolvedValueOnce({
        sources: [{ name: 'a', is_runtime: true } as M3uSource],
      });

      const filters: FilterConfig[] = [{ id: 1, name: 'f1', attribute: 'group_title', is_runtime: true }];

      renderPage({ filters, filterOrigin: [] });

      await waitFor(() => expect(screen.getByText('3 active overrides')).toBeInTheDocument());
    });

    it('shows a distinct restart-required count', async () => {
      vi.mocked(api.getSettings).mockResolvedValueOnce({
        settings: [
          { key: 'radarr.url', value: 'http://x', sensitive: false, origin: 'interface', restart_required: true },
        ] as SettingsField[],
      });

      renderPage();

      await waitFor(() => expect(screen.getByText('1 require a restart')).toBeInTheDocument());
    });
  });

  describe('per-tab override badges', () => {
    it('shows no badge on a tab with zero overrides', () => {
      renderPage();
      const generalTab = screen.getByRole('tab', { name: 'General' });
      expect(generalTab.querySelector('.settings-tab-badge')).toBeNull();
    });

    it('shows a count badge and a restart-required marker on a tab with an overridden, restart-required field', async () => {
      vi.mocked(api.getSettings).mockResolvedValueOnce({
        settings: [
          { key: 'radarr.url', value: 'http://x', sensitive: false, origin: 'interface', restart_required: true },
        ] as SettingsField[],
      });

      renderPage();

      const integrationsTab = await screen.findByRole('tab', { name: /Integrations/ });
      await waitFor(() => expect(integrationsTab.querySelector('.settings-tab-badge')).toHaveTextContent('1'));
      expect(integrationsTab.querySelector('.settings-tab-restart-dot')).not.toBeNull();
    });

    it('sums M3U sources and filters into the "Content" tab badge', async () => {
      vi.mocked(api.getM3uSources).mockResolvedValueOnce({
        sources: [{ name: 'a', is_runtime: true } as M3uSource],
      });
      const filters: FilterConfig[] = [{ id: 1, name: 'f1', attribute: 'group_title', is_runtime: true }];

      renderPage({ filters });

      const contentTab = await screen.findByRole('tab', { name: /Content/ });
      await waitFor(() => expect(contentTab.querySelector('.settings-tab-badge')).toHaveTextContent('2'));
    });
  });

  describe('quick search', () => {
    it('filters visible fields while typing, clears on empty input, and persists across a tab switch', async () => {
      vi.mocked(api.getSettings).mockResolvedValue({
        settings: [
          { key: 'radarr.url', value: 'http://radarr.example.com', sensitive: false, origin: 'config', restart_required: false },
          { key: 'sonarr.url', value: 'http://sonarr.example.com', sensitive: false, origin: 'config', restart_required: false },
        ] as SettingsField[],
      });

      renderPage();

      fireEvent.mouseDown(screen.getByRole('tab', { name: /Integrations/ }));
      await screen.findByDisplayValue('http://radarr.example.com');

      const search = screen.getByPlaceholderText('Search settings…');
      fireEvent.change(search, { target: { value: 'url' } });
      expect(screen.getByDisplayValue('http://radarr.example.com')).toBeInTheDocument();

      // Switching tabs keeps the query.
      fireEvent.mouseDown(screen.getByRole('tab', { name: 'General' }));
      fireEvent.mouseDown(screen.getByRole('tab', { name: /Integrations/ }));
      expect(screen.getByPlaceholderText('Search settings…')).toHaveValue('url');
      expect(screen.getByDisplayValue('http://radarr.example.com')).toBeInTheDocument();

      fireEvent.change(search, { target: { value: '' } });
      expect(screen.getByDisplayValue('http://radarr.example.com')).toBeInTheDocument();
      expect(screen.getByDisplayValue('http://sonarr.example.com')).toBeInTheDocument();
    });
  });

  describe('settingsTab URL/localStorage precedence', () => {
    it('uses the URL settingsTab over localStorage and the default', () => {
      localStorage.setItem('stalkeer_configuration_tab', 'advanced');
      window.history.replaceState(null, '', '/?settingsTab=notifications');

      renderPage();

      expect(screen.getByRole('tab', { name: /Notifications/ })).toHaveAttribute('data-state', 'active');
    });

    it('falls back to localStorage when no URL settingsTab is present', () => {
      localStorage.setItem('stalkeer_configuration_tab', 'advanced');

      renderPage();

      expect(screen.getByRole('tab', { name: /Advanced/ })).toHaveAttribute('data-state', 'active');
    });

    it('falls back to "General" when neither the URL nor localStorage carries a tab', () => {
      renderPage();
      expect(screen.getByRole('tab', { name: 'General' })).toHaveAttribute('data-state', 'active');
    });

    it('updates the URL and localStorage when a tab is selected', () => {
      renderPage();

      fireEvent.mouseDown(screen.getByRole('tab', { name: /Advanced/ }));

      expect(localStorage.getItem('stalkeer_configuration_tab')).toBe('advanced');
      expect(new URLSearchParams(window.location.search).get('settingsTab')).toBe('advanced');
    });
  });

  describe('mobile tab dropdown', () => {
    beforeEach(() => {
      vi.mocked(useIsMobile).mockReturnValue(true);
    });

    it('renders a select with the six tabs instead of the tab bar', () => {
      renderPage();

      expect(screen.queryByRole('tablist')).toBeNull();
      const select = screen.getByRole('combobox', { name: 'Settings section' });
      const options = Array.from(select.querySelectorAll('option')).map(o => o.textContent);
      expect(options).toEqual(['General', 'Integrations', 'Content', 'Notifications', 'Advanced', 'System']);
    });

    it('encodes the override count and restart marker as option text', async () => {
      vi.mocked(api.getSettings).mockResolvedValueOnce({
        settings: [
          { key: 'radarr.url', value: 'http://x', sensitive: false, origin: 'interface', restart_required: true },
        ] as SettingsField[],
      });

      renderPage();

      const select = screen.getByRole('combobox', { name: 'Settings section' }) as HTMLSelectElement;
      await waitFor(() => {
        const integrationsOption = Array.from(select.querySelectorAll('option')).find(o => o.value === 'integrations');
        expect(integrationsOption?.textContent).toBe('⚠ Integrations (1)');
      });
    });

    it('switches the active tab, updates the URL, and matches desktop tab-switch behavior', () => {
      renderPage();

      const select = screen.getByRole('combobox', { name: 'Settings section' }) as HTMLSelectElement;
      fireEvent.change(select, { target: { value: 'system' } });

      expect(select.value).toBe('system');
      expect(localStorage.getItem('stalkeer_configuration_tab')).toBe('system');
      expect(new URLSearchParams(window.location.search).get('settingsTab')).toBe('system');
      expect(screen.getByText('Stalkeer Dashboard')).toBeInTheDocument();
    });
  });
});
