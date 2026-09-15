## Context

See proposal.md - Why. Two independent mobile-only overflow sources on the Configuration page (`frontend/src/components/ConfigurationPage.tsx` and its child sections), both stemming from the same gap: nothing in this page's CSS stops content wider than the viewport from pushing the page itself into a horizontal scrollbar, unlike the rest of the app where `frontend-responsive-layout` closes that gap explicitly for every other screen.

- Tab navigation: `.settings-tabs-list` is `flex-wrap: wrap` by default, but a mobile media query (`index.css:1278-1288`) overrides it to `flex-wrap: nowrap; overflow-x: auto`, actively defeating the wrap that would otherwise happen for free.
- Technical values: `M3uSourcesSection.tsx` renders `file_path`/`url` inside `<code>`, and `SystemStatusSection.tsx` renders disk mount paths as plain text, both inside `.filter-card` / `.settings-group-card`. Neither has `overflow-wrap`/`word-break`, so a single unbroken long token (a URL with no spaces, a long Docker volume path) overflows its card and the page.

`ConfigurationPage.tsx` does not currently import `useIsMobile()` (from `frontend/src/hooks/useMediaQuery.ts`), unlike `RadarrSonarrTab.tsx` and `PlaylistTab.tsx` which already use it to branch rendering by viewport.

## Goals / Non-Goals

**Goals:**
- Eliminate horizontal page scroll on the Configuration page at mobile widths, for both the tab bar and long technical values.
- Keep the mobile tab control's information parity with desktop (override counts, restart-required marker).
- Reuse patterns already established elsewhere in this codebase (`useIsMobile()`, `.custom-select`) rather than introducing new ones.

**Non-Goals:**
- Redesigning the desktop tab bar or card visuals - unchanged.
- A general audit of every other page for the same overflow gap - scoped to the Configuration page only, per this change's proposal.
- Changing how M3U source or disk usage data is fetched or formatted - only how the existing strings are laid out.

## Decisions

### Mobile tab control: `useIsMobile()`-gated `<select>`, replacing `Tabs.List` entirely
`ConfigurationPage.tsx` imports `useIsMobile()` and renders either the existing `Tabs.List`/`Tabs.Trigger` bar (desktop) or a `<select className="custom-select">` (mobile) that calls the existing `setActiveSettingsTab` on change - the same function `Tabs.Root`'s `onValueChange` already calls. `Tabs.Content` panels are unchanged; only the trigger control differs.

Alternatives considered:
- **CSS-only** (`display: none` swap between two parallel trigger markups instead of a JS branch): rejected - would still mount two full sets of `Tabs.Trigger`/`<option>` elements and duplicate the override-count/restart-marker computation in two render paths, for no benefit since `useIsMobile()` is already the established pattern on this exact page family.
- **Keep pills, remove `overflow-x: auto` and let them wrap** (the cheapest possible fix, since `flex-wrap: wrap` is already the base behavior): rejected per the confirmed direction (dropdown) - two-row pills still cost vertical space and don't solve the badge/restart-marker legibility problem at small sizes as cleanly as text in a select option.

### Option label encodes override count and restart marker as text
Label composition mirrors `renderTabBadge`'s logic, serialized to a string: `${restartRequired ? '⚠ ' : ''}${label}${count > 0 ? ' (' + count + ')' : ''}`. No count suffix when `count === 0`, no leading marker when no override on that tab requires a restart - matching desktop's "no badge when zero" behavior.

### Overflow fix: `overflow-wrap: anywhere` on the technical-value containers
Applied to the `<code>` elements in `M3uSourcesSection.tsx` and the paths line in `SystemStatusSection.tsx`'s `DiskUsageRow` (via a shared utility class, e.g. `.settings-value-wrap`, added in `index.css` next to `.filter-card`).

Alternatives considered:
- **`word-break: break-all`**: breaks at any character boundary including mid-word in surrounding translated label text if the class were applied too broadly; `overflow-wrap: anywhere` only breaks when a line would otherwise overflow, preserving normal word breaking everywhere else - safer given these containers mix translated labels with raw technical values.
- **Truncate with ellipsis + `title` tooltip**: rejected - these are diagnostic values (a source URL, a disk mount path) a user may need to read or copy in full; tooltips are also unreliable on touch (the primary mobile input), so truncation would hide information with no reliable way to recover it on the viewport where this bug actually manifests.

## Risks / Trade-offs

- [Native `<select>` cannot show a colored badge/dot, only text] → Accepted per the confirmed direction; the text encoding (`⚠ Intégrations (3)`) is intentionally terser than desktop, which is consistent with how "Startup tab" already presents plain-text options on this same page.
- [Hiding `Tabs.List` on mobile removes the visible ARIA `tablist`/`tab` relationship for `Tabs.Content`] → Accepted trade-off: the `<select>` becomes the sole navigation control on mobile, matching the existing "Startup tab" `<select>` pattern already used on this same page; no regression relative to today, since today's mobile tab bar is a horizontally-scrolling `tablist` that is itself hard to operate with assistive tech.
- [`overflow-wrap: anywhere` on a class shared between M3U cards and disk usage rows could unintentionally affect other future content added to those containers] → Mitigate by scoping the new class to the specific value elements (the `<code>` tags, the paths line) rather than the whole card.

## Migration Plan

Frontend-only, no data migration. Ship as a normal PR; no feature flag needed since the change is purely presentational and reversible by reverting the CSS/markup diff. No rollback complexity beyond a standard revert.
