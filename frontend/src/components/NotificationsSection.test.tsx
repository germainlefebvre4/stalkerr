import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { NotificationsSection } from './NotificationsSection';
import { SettingsField } from '../types';

const settings: SettingsField[] = [
  { key: 'notifications.enabled', value: true, sensitive: false, origin: 'config', restart_required: false },
  { key: 'notifications.ntfy.enabled', value: true, sensitive: false, origin: 'config', restart_required: false },
  { key: 'notifications.ntfy.server_url', value: 'https://ntfy.sh', sensitive: false, origin: 'config', restart_required: false },
  { key: 'notifications.ntfy.topic', value: 'stalkeer', sensitive: false, origin: 'config', restart_required: false },
  { key: 'notifications.ntfy.auth_token', is_set: false, sensitive: true, origin: 'config', restart_required: false },
];

describe('NotificationsSection', () => {
  afterEach(() => cleanup());

  it('renders every notification field with its effective value', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <NotificationsSection
          isExpanded
          settings={settings}
          loading={false}
          onSetSetting={vi.fn()}
          onClearSetting={vi.fn()}
        />
      </I18nextProvider>
    );

    expect(screen.getByDisplayValue('https://ntfy.sh')).toBeInTheDocument();
    expect(screen.getByDisplayValue('stalkeer')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Not set')).toBeInTheDocument();
    // Both boolean fields render as toggles.
    expect(screen.getAllByRole('checkbox').length).toBe(2);
  });

  it('renders nothing when collapsed', () => {
    const { container } = render(
      <I18nextProvider i18n={i18n}>
        <NotificationsSection
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
