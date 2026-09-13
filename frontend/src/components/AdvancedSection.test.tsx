import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { AdvancedSection } from './AdvancedSection';
import { SettingsField, BootstrapField } from '../types';

const settings: SettingsField[] = [
  { key: 'downloads.movies_path', value: '/media/movies', sensitive: false, origin: 'config', restart_required: false },
  { key: 'downloads.timeout', value: 300, sensitive: false, origin: 'config', restart_required: true },
  { key: 'logging.app.level', value: 'info', sensitive: false, origin: 'config', restart_required: false },
  { key: 'm3u.update_interval', value: 3600, sensitive: false, origin: 'config', restart_required: false },
];

const bootstrap: BootstrapField[] = [
  { key: 'database.host', value: 'db-host', sensitive: false, origin: 'config' },
  { key: 'database.password', is_set: true, sensitive: true, origin: 'config' },
  { key: 'api.port', value: 8080, sensitive: false, origin: 'config' },
];

describe('AdvancedSection', () => {
  afterEach(() => cleanup());

  it('renders Downloads, logging, and M3U fields as editable', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <AdvancedSection
          isExpanded
          settings={settings}
          bootstrap={bootstrap}
          loading={false}
          onFetchSettings={vi.fn()}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
        />
      </I18nextProvider>
    );

    expect(screen.getByDisplayValue('/media/movies')).toBeInTheDocument();
    expect(screen.getByDisplayValue('300')).toBeInTheDocument();
    expect(screen.getByDisplayValue('info')).toBeInTheDocument();
    expect(screen.getByDisplayValue('3600')).toBeInTheDocument();
    expect(screen.getByText('Restart required')).toBeInTheDocument();
  });

  it('renders the bootstrap configuration read-only, with no edit control', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <AdvancedSection
          isExpanded
          settings={settings}
          bootstrap={bootstrap}
          loading={false}
          onFetchSettings={vi.fn()}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
        />
      </I18nextProvider>
    );

    expect(screen.getByText('db-host')).toBeInTheDocument();
    expect(screen.getByText('8080')).toBeInTheDocument();
    // Sensitive bootstrap field is masked too.
    expect(screen.getByText('Set')).toBeInTheDocument();

    // No input/select/button edit controls anywhere in the bootstrap group.
    const bootstrapHeading = screen.getByText('Database & ports (read-only)');
    const bootstrapGroup = bootstrapHeading.parentElement as HTMLElement;
    expect(bootstrapGroup.querySelectorAll('input, select, button').length).toBe(0);
  });
});
