import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { ConfigurationPage } from './ConfigurationPage';

function renderPage() {
  return render(
    <I18nextProvider i18n={i18n}>
      <ConfigurationPage
        onBack={vi.fn()}
        theme="light" onSetTheme={vi.fn()}
        reduceMotion={false} onSetReduceMotion={vi.fn()}
        startupTab="home" onSetStartupTab={vi.fn()} tabOptions={[{ value: 'home', label: 'Home' }]}
        playlistLimit={10} onSetPlaylistLimit={vi.fn()}
        playlistView="items" onSetPlaylistView={vi.fn()}
        filters={[]} filtersLoading={false} onFetchFilters={vi.fn()}
        onDeleteFilter={vi.fn()} onOpenCreateFilter={vi.fn()}
      />
    </I18nextProvider>
  );
}

describe('ConfigurationPage', () => {
  afterEach(() => cleanup());

  it('renders all ten sections in the fixed order', () => {
    renderPage();

    const titles = Array.from(
      document.querySelectorAll('.settings-section-header, .advanced-filters-toggle-label')
    ).map(el => el.textContent);

    expect(titles).toEqual([
      'Appearance', 'Language', 'Filters', 'M3U Sources', 'Integrations',
      'Notifications', 'Advanced', 'Preferences', 'System', 'About',
    ]);
  });

  it('collapses Filters, M3U Sources, Integrations, Notifications, Advanced, and System by default', () => {
    renderPage();

    const collapsed = screen.getAllByRole('button', { expanded: false })
      .map(btn => btn.querySelector('.advanced-filters-toggle-label')?.textContent);

    expect(collapsed).toEqual(['Filters', 'M3U Sources', 'Integrations', 'Notifications', 'Advanced', 'System']);
  });

  it('leaves Appearance, Language, Preferences, and About always expanded (not disclosure toggles)', () => {
    renderPage();

    const alwaysExpandedHeaders = Array.from(document.querySelectorAll('.settings-section-header'));
    expect(alwaysExpandedHeaders.map(el => el.textContent)).toEqual([
      'Appearance', 'Language', 'Preferences', 'About',
    ]);
    // None of these are disclosure toggle buttons.
    alwaysExpandedHeaders.forEach(el => {
      expect(el.tagName).toBe('H3');
      expect(el).not.toHaveAttribute('aria-expanded');
    });
  });

  it('calls onBack when the close button is clicked', () => {
    // Rendered separately so onBack is captured for this assertion.
    const onBack = vi.fn();
    render(
      <I18nextProvider i18n={i18n}>
        <ConfigurationPage
          onBack={onBack}
          theme="light" onSetTheme={vi.fn()}
          reduceMotion={false} onSetReduceMotion={vi.fn()}
          startupTab="home" onSetStartupTab={vi.fn()} tabOptions={[{ value: 'home', label: 'Home' }]}
          playlistLimit={10} onSetPlaylistLimit={vi.fn()}
          playlistView="items" onSetPlaylistView={vi.fn()}
          filters={[]} filtersLoading={false} onFetchFilters={vi.fn()}
          onDeleteFilter={vi.fn()} onOpenCreateFilter={vi.fn()}
        />
      </I18nextProvider>
    );

    screen.getByText('Close').click();
    expect(onBack).toHaveBeenCalled();
  });
});
