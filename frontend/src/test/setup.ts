import '@testing-library/jest-dom/vitest';

// jsdom does not implement matchMedia; default to "no match" (desktop) so
// components using useIsMobile()/useMediaQuery() don't throw when rendered
// in tests that don't care about responsive behavior.
if (typeof window !== 'undefined' && !window.matchMedia) {
  window.matchMedia = ((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof window.matchMedia;
}
