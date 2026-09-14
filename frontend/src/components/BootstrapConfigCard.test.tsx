import { afterEach, describe, expect, it } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { BootstrapConfigCard } from './BootstrapConfigCard';
import { BootstrapField } from '../types';

const bootstrap: BootstrapField[] = [
  { key: 'database.host', value: 'db-host', sensitive: false, origin: 'config' },
  { key: 'database.password', is_set: true, sensitive: true, origin: 'config' },
  { key: 'api.port', value: 8080, sensitive: false, origin: 'config' },
];

describe('BootstrapConfigCard', () => {
  afterEach(() => cleanup());

  it('renders the bootstrap configuration read-only, with no edit control', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <BootstrapConfigCard bootstrap={bootstrap} />
      </I18nextProvider>
    );

    expect(screen.getByText('db-host')).toBeInTheDocument();
    expect(screen.getByText('8080')).toBeInTheDocument();
    // Sensitive bootstrap field is masked too.
    expect(screen.getByText('Set')).toBeInTheDocument();

    const bootstrapHeading = screen.getByText('Database & ports (read-only)');
    const bootstrapGroup = bootstrapHeading.closest('.settings-group-card') as HTMLElement;
    // The compact origin-indicator dot is a focusable button (for its hover/
    // focus tooltip) but not an edit control; everything else must be absent.
    expect(bootstrapGroup.querySelectorAll('input, select, button:not(.settings-origin-dot)').length).toBe(0);
  });

  it('filters rows by the search query', () => {
    render(
      <I18nextProvider i18n={i18n}>
        <BootstrapConfigCard bootstrap={bootstrap} searchQuery="port" />
      </I18nextProvider>
    );

    expect(screen.getByText('8080')).toBeInTheDocument();
    expect(screen.queryByText('db-host')).not.toBeInTheDocument();
  });
});
