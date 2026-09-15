import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { AdvancedSection } from './AdvancedSection';
import { SettingsField } from '../types';

const settings: SettingsField[] = [
  { key: 'downloads.movies_path', value: '/media/movies', sensitive: false, origin: 'config', restart_required: false },
  { key: 'downloads.timeout', value: 300, sensitive: false, origin: 'config', restart_required: true },
  { key: 'logging.app.level', value: 'info', sensitive: false, origin: 'config', restart_required: false },
  { key: 'logging.database.level', value: 'warn', sensitive: false, origin: 'config', restart_required: false },
  { key: 'm3u.update_interval', value: 3600, sensitive: false, origin: 'config', restart_required: false },
];

describe('AdvancedSection', () => {
  afterEach(() => cleanup());

  it('renders Downloads, logging, and M3U fields as editable', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <AdvancedSection
          isExpanded
          settings={settings}
          loading={false}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
        />
      </I18nextProvider>
    );

    expect(screen.getByDisplayValue('/media/movies')).toBeInTheDocument();
    expect(screen.getByDisplayValue('300')).toBeInTheDocument();
    expect(screen.getByDisplayValue('3600')).toBeInTheDocument();
    expect(screen.getByText('Restart required')).toBeInTheDocument();
  });

  it('renders the app and database log level fields as a select with the four backend-accepted levels', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <AdvancedSection
          isExpanded
          settings={settings}
          loading={false}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
        />
      </I18nextProvider>
    );

    const selects = screen.getAllByRole('combobox', { name: 'App log level' })
      .concat(screen.getAllByRole('combobox', { name: 'Database log level' }));
    expect(selects).toHaveLength(2);
    for (const select of selects) {
      const optionLabels = Array.from(select.querySelectorAll('option')).map(o => o.textContent);
      expect(optionLabels).toEqual(['Debug', 'Info', 'Warning', 'Error']);
    }
  });

  it('no longer renders the bootstrap (read-only database/ports) group', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <AdvancedSection
          isExpanded
          settings={settings}
          loading={false}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
        />
      </I18nextProvider>
    );

    expect(screen.queryByText('Database & ports (read-only)')).not.toBeInTheDocument();
  });

  it('renders nothing when collapsed', () => {
    const { container } = render(
      <I18nextProvider i18n={i18n}>
        <AdvancedSection
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
});
