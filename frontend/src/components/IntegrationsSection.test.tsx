import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { IntegrationsSection } from './IntegrationsSection';
import { SettingsField } from '../types';

const settings: SettingsField[] = [
  { key: 'radarr.url', value: 'http://radarr.example.com', sensitive: false, origin: 'config', restart_required: false },
  { key: 'radarr.enabled', value: true, sensitive: false, origin: 'config', restart_required: false },
  { key: 'sonarr.url', value: 'http://sonarr.example.com', sensitive: false, origin: 'config', restart_required: false },
  { key: 'tmdb.api_key', is_set: true, sensitive: true, origin: 'config', restart_required: true },
  { key: 'jellyfin.url', value: 'http://jellyfin.example.com', sensitive: false, origin: 'config', restart_required: false },
];

describe('IntegrationsSection', () => {
  afterEach(() => cleanup());

  it('renders a group with its fields for each of Radarr, Sonarr, TMDB, and Jellyfin', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <IntegrationsSection
          isExpanded
          settings={settings}
          loading={false}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
        />
      </I18nextProvider>
    );

    expect(screen.getByText('Radarr')).toBeInTheDocument();
    expect(screen.getByText('Sonarr')).toBeInTheDocument();
    expect(screen.getByText('TMDB')).toBeInTheDocument();
    expect(screen.getByText('Jellyfin')).toBeInTheDocument();

    expect(screen.getByDisplayValue('http://radarr.example.com')).toBeInTheDocument();
    expect(screen.getByDisplayValue('http://sonarr.example.com')).toBeInTheDocument();
    expect(screen.getByDisplayValue('http://jellyfin.example.com')).toBeInTheDocument();
    // The TMDB API key is sensitive: masked, never showing the raw value.
    expect(screen.getByPlaceholderText('Set')).toBeInTheDocument();
    // The boolean Radarr "enabled" field renders as a toggle, not a select.
    expect(screen.getAllByRole('checkbox').length).toBeGreaterThan(0);
  });

  it('renders nothing when collapsed', () => {
    const { container } = render(
      <I18nextProvider i18n={i18n}>
        <IntegrationsSection
          isExpanded={false}
          settings={settings}
          loading={false}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
        />
      </I18nextProvider>
    );

    expect(container).toBeEmptyDOMElement();
  });

  it('filters visible fields by the search query, hiding non-matching fields and groups with no match', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <IntegrationsSection
          isExpanded
          settings={settings}
          loading={false}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
          searchQuery="url"
        />
      </I18nextProvider>
    );

    // Radarr, Sonarr, and Jellyfin each have a matching "URL" field.
    expect(screen.getByDisplayValue('http://radarr.example.com')).toBeInTheDocument();
    expect(screen.getByDisplayValue('http://sonarr.example.com')).toBeInTheDocument();
    expect(screen.getByDisplayValue('http://jellyfin.example.com')).toBeInTheDocument();
    // Radarr's non-matching "Enabled" boolean field is hidden.
    expect(screen.queryAllByRole('checkbox').length).toBe(0);
    // TMDB's only present field (API key) doesn't match "url": the whole group hides.
    expect(screen.queryByText('TMDB')).not.toBeInTheDocument();
  });
});
