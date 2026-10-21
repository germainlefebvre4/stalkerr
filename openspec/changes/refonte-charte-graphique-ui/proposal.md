# Proposal

## Why

The dashboard looks like a generic SaaS template: Inter (declared but never loaded, so the system font is what renders), an indigo→violet gradient, glassmorphism, soft diffuse shadows, 12–24px radii, pastel badges and emoji icons. It has no identity of its own, while the product is a supervision tool (M3U ingestion, downloads, Radarr/Sonarr, errors, logs) whose name and 🛰️ favicon already suggest a "mission control" console. Six top-level sections also no longer fit comfortably in a segmented pill bar. A redesign is wanted now, and the token layer (`variables.css`) is already clean enough to carry most of it.

## What Changes

- Adopt a **"mission control"** visual identity: near-black blue-tinted surfaces separated by 1px rules, a single flat cyan "signal" accent, no gradient, no glass, no shadows, 4–6px radii.
- **Dark theme becomes the default**; the light theme is kept as a sober secondary theme (printed-console look) reachable through the existing toggle.
- **Desktop shell**: replace the glass header + segmented tab pills with a fixed left **sidebar** (logo, version, nav with lucide icons, error-count badge, API health LED, settings entry). Replace the KPI cards with a thin **telemetry bar**. The mobile layout (bottom tab bar, header row with title + settings icon, card lists) keeps its structure and is only re-skinned.
- **Status vocabulary**: pastel badges become **LED + short label**, with a pulse on active states.
- **Typography**: bundled IBM Plex Sans (UI) and IBM Plex Mono (sizes, rates, paths, hashes, timestamps), self-hosted in the bundle — no CDN, since this is a homelab app.
- **Icons**: emojis in tabs, header and favicon are replaced by `lucide-react` icons and an SVG favicon.
- **Motion**: the card hover lift is removed; "reduce animations" now covers the LED pulse and the toast entrance animation.
- **Inline style cleanup**: the 7 hard-coded colors (`ManualOverrideDialog`, `ErrorsSidepanel`, `MediaOccurrenceDrawer`) and the `rgba(99,102,241…)` literals move to tokens; the ~637 `style={{…}}` usages are migrated to classes screen by screen for radius, density and mono typography consistency. Most of the effect comes for free from the ~320 existing `var(--token)` usages.
- Delivered in four ordered phases, the app staying usable after each: (1) foundations, (2) shell, (3) primitives, (4) screens.

Out of scope: new product features, backend/API changes, i18n key changes beyond renaming a few labels, poster/artwork imagery.

## Capabilities

### New Capabilities
- `frontend-design-system`: the visual language of the dashboard — design tokens (dark and light), bundled typography, flat/no-shadow surface rules, LED status indicators, lucide iconography, favicon, and the rule that components consume tokens instead of hard-coded colors.

### Modified Capabilities
- `frontend-configuration-page`: Theme Toggle defaults to dark when no preference is stored; Reduce Motion no longer refers to the card hover lift but to the LED pulse and toast animation; Settings Header Entry Point is relocated to the sidebar on desktop (still on the title row on mobile).
- `frontend-responsive-layout`: at or above 768px the navigation is a left sidebar with a telemetry bar instead of a header + segmented tab pills + KPI cards; the bottom tab bar below 768px is unchanged; the "desktop unchanged" statements are narrowed to behavior (tables, cards, pagination) rather than visual chrome.
- `frontend-ihm-dashboard`: the requirement "Modern Light-Themed Radix UI Dialog for File Re-organization" is renamed to drop the "Light-Themed" qualifier (behavior unchanged).

## Impact

- **Code (frontend only)**: `frontend/src/variables.css`, `frontend/src/index.css` (1612 lines), `frontend/index.html` (favicon, theme bootstrap script defaulting to dark), `frontend/src/App.tsx` (shell, nav), `FloatingHeader.tsx` (replaced by a sidebar component), and the components with the heaviest inline styles (`RadarrSonarrTab`, `ErrorsSidepanel`, `MediaOccurrenceDrawer`, `ManualOverrideDialog`, `DownloadsTab`, Playlist/Home/Logs/Settings components). Existing component tests asserting on emojis, header or tab markup need updating.
- **Dependencies**: add self-hosted font packages (e.g. `@fontsource/ibm-plex-sans`, `@fontsource/ibm-plex-mono`); `lucide-react` is already installed.
- **Specs**: delta specs for the three modified capabilities above and one new capability.
- **Coordination**: the in-flight change `downloads-hide-cancelled-and-search` (16/17 tasks) touches `DownloadsTab.tsx`; finish and archive it before starting phase 4 on Downloads.
- **Users**: first-time and cleared-storage users will see the dark theme; users with a stored preference keep it.
