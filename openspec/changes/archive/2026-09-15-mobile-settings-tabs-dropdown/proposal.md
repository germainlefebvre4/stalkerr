## Why

On mobile, the Configuration page forces horizontal scrolling in two places: the six-tab navigation bar (Général/Intégrations/Contenu/Notifications/Avancé/Système) overflows sideways instead of wrapping, and long technical values (M3U source file paths/URLs, disk mount paths) push their cards past the viewport edge. Both contradict the app-wide principle, already established for every other screen by `frontend-responsive-layout`, that browsing and switching sections should never require horizontal scrolling on a phone. The original Configuration page redesign explicitly deferred a mobile-specific tab treatment; this change delivers it, along with the related overflow fix surfaced while investigating it.

## What Changes

- Below the mobile breakpoint, replace the horizontally-scrolling tab pill bar with a native `<select>` that drives the same active-tab state as the desktop tabs, fully replacing the pill bar (not shown alongside it) on mobile.
- Each `<select>` option's label includes the same override count and restart-required signal shown on desktop tab badges, encoded as text (e.g. "⚠ Intégrations (3)") since a native `<option>` cannot render a colored badge.
- Ensure the M3U source card's file path and URL values wrap or break instead of forcing the page to scroll horizontally, regardless of value length.
- Ensure the "Système" tab's disk usage entries (mount paths) wrap or break instead of forcing the page to scroll horizontally, regardless of value length.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `frontend-configuration-page`: tab navigation switches to a native dropdown below the mobile breakpoint (instead of a horizontally-scrolling pill bar), carrying the same override-count/restart-required signal as text; the "Système" tab's disk usage values no longer force horizontal scroll.
- `frontend-app-settings-management`: the M3U source card's file path and URL values no longer force horizontal scroll.

## Impact

- `frontend/src/components/ConfigurationPage.tsx`: add a mobile-only `<select>` tab switcher wired to the existing `activeSettingsTab` state; needs `useIsMobile()` (already used elsewhere, e.g. `RadarrSonarrTab`, `PlaylistTab`) to choose between the select and the existing `Tabs.List`.
- `frontend/src/components/M3uSourcesSection.tsx` and `frontend/src/components/SystemStatusSection.tsx`: apply wrapping to the long technical text they render.
- `frontend/src/index.css`: remove or replace the mobile `overflow-x: auto` rule on `.settings-tabs-list`; add wrapping rules for the affected card text.
- No backend, API, or persisted-state changes; `activeSettingsTab` storage/URL behavior is unchanged, only the control that sets it gains a mobile variant.
