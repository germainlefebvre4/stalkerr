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

// jsdom does not implement ResizeObserver, which Radix's Tooltip (via
// @radix-ui/react-use-size) reads to measure its trigger. A no-op stub is
// enough for tests that don't assert on actual measured sizes.
if (typeof window !== 'undefined' && !window.ResizeObserver) {
  window.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
}
