import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, act } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from './i18n';
import App from './App';

afterEach(() => {
  cleanup();
  localStorage.clear();
  window.history.replaceState(null, '', '/');
});

function setMatchMedia(matches: boolean) {
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches,
    media: query,
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })) as unknown as typeof window.matchMedia;
}

describe('App mobile fallback off the Erreurs tab', () => {
  it('falls back to another tab when mounted below the 768px breakpoint with errors active', async () => {
    window.history.replaceState(null, '', '/?tab=errors');
    setMatchMedia(true);

    await act(async () => {
      render(
        <I18nextProvider i18n={i18n}>
          <App />
        </I18nextProvider>
      );
    });

    // The Erreurs tab's own content must not be the active panel once the
    // mobile fallback effect has run.
    expect(screen.queryByText('Naming Errors')).not.toBeInTheDocument();
    // Its trigger is also absent from the mobile bottom tab bar.
    expect(screen.queryByText('Errors')).not.toBeInTheDocument();
  });

  it('keeps the Erreurs tab active on a desktop-width viewport', async () => {
    window.history.replaceState(null, '', '/?tab=errors');
    setMatchMedia(false);

    await act(async () => {
      render(
        <I18nextProvider i18n={i18n}>
          <App />
        </I18nextProvider>
      );
    });

    expect(screen.getByText('Naming Errors')).toBeInTheDocument();
  });
});

describe('App falls back off the removed Filtres tab', () => {
  it('falls back to Home on a mobile-width viewport ("filters" is no longer a valid tab)', async () => {
    window.history.replaceState(null, '', '/?tab=filters');
    setMatchMedia(true);

    await act(async () => {
      render(
        <I18nextProvider i18n={i18n}>
          <App />
        </I18nextProvider>
      );
    });

    expect(screen.getByText('Last Processing Run')).toBeInTheDocument();
  });

  it('falls back to Home on a desktop-width viewport ("filters" is no longer a valid tab)', async () => {
    window.history.replaceState(null, '', '/?tab=filters');
    setMatchMedia(false);

    await act(async () => {
      render(
        <I18nextProvider i18n={i18n}>
          <App />
        </I18nextProvider>
      );
    });

    expect(screen.getByText('Last Processing Run')).toBeInTheDocument();
  });
});

describe('App default tab', () => {
  it('activates Home when no tab param and no localStorage value are present', async () => {
    setMatchMedia(false);

    await act(async () => {
      render(
        <I18nextProvider i18n={i18n}>
          <App />
        </I18nextProvider>
      );
    });

    expect(screen.getByText('Last Processing Run')).toBeInTheDocument();
  });

  it('restores a previously selected tab over the Home default', async () => {
    window.history.replaceState(null, '', '/?tab=downloads');
    setMatchMedia(false);

    await act(async () => {
      render(
        <I18nextProvider i18n={i18n}>
          <App />
        </I18nextProvider>
      );
    });

    expect(screen.queryByText('Last Processing Run')).not.toBeInTheDocument();
  });
});

describe('App mobile bottom tab bar', () => {
  it('shows Home first and does not show Filtres', async () => {
    setMatchMedia(true);

    let container!: HTMLElement;
    await act(async () => {
      ({ container } = render(
        <I18nextProvider i18n={i18n}>
          <App />
        </I18nextProvider>
      ));
    });

    const labels = Array.from(container.querySelectorAll('.mobile-tab-bar-label')).map(el => el.textContent);
    expect(labels[0]).toBe('Home');
    expect(labels).not.toContain('Filters');
  });
});
