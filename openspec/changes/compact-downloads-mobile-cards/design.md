## Context

`DownloadsTab.tsx` is currently a stateless presentational component: it receives `downloads` and filter state via props and renders one card per item with no internal state. The Downloads list auto-refreshes every 5 seconds (existing behavior, `downloads-display-ui`), replacing the `downloads` array on each fetch. Mobile/desktop layout switches purely via CSS media queries (`max-width: 767.98px` in `index.css`) — there is no JS-level viewport branching today. See `proposal.md` for motivation and `specs/frontend-responsive-layout/spec.md` for the target behavior.

## Goals / Non-Goals

**Goals:**
- Track which download cards are expanded, independently per card, surviving the 5s auto-refresh re-render.
- Keep the collapse/expand purely a mobile-viewport concern with zero behavior change on desktop.

**Non-Goals:**
- No change to data fetching, filters, or the auto-refresh interval.
- No new shared/global expand-state (e.g. no "expand all" control) — out of scope per the conversation that shaped this change.

## Decisions

**Expand state: `Set<number>` of expanded download ids, local `useState` in `DownloadsTab`.**
Using the download's `id` (stable across refetches) rather than array index avoids state misattachment when items reorder or the list changes between polls. Alternative considered: per-card local state via a wrapper component — rejected as unnecessary complexity for a single flat list; a single `Set` in the parent is simpler and keeps the existing single-component structure intact.

**Collapse/expand is CSS-visibility only, not conditional unmount.**
The secondary fields (format/resolution/duration, validation badges, genres) render in the DOM at all times; a class toggle (e.g. `.download-card--expanded`) controls their visibility via CSS. This avoids remount cost on every toggle and keeps desktop rendering (always expanded) a no-op case of the same markup. Alternative considered: conditional rendering (`{expanded && <div>...}`) — rejected because it would require duplicating the "always visible on desktop" logic as a second code path instead of a single CSS breakpoint override.

**Mobile-only via CSS, no JS viewport check.**
Consistent with the existing pattern in `index.css` (`@media (max-width: 767.98px)`): the accordion trigger (tap target, chevron) is rendered unconditionally but is inert/hidden on desktop via CSS, and the secondary-fields section is force-visible on desktop via the same media query overriding the collapsed class. This avoids introducing a `useMediaQuery`-style hook that doesn't otherwise exist in this codebase.

**`download_path` replaces `item.url` and the `file_info.folder_name`/`file_name` box in the mobile essential section.**
`download_path` is already present on `DownloadEnriched` (currently unused for display, only as a title fallback). Reusing it avoids any backend change. Desktop keeps its current separate URL line and folder/file box unchanged — this substitution is mobile-only markup.

## Risks / Trade-offs

- [Expanded-state `Set` grows unbounded across many refetches if a download disappears while expanded] → Low impact (small in-memory set, cleared on tab unmount); no cleanup needed given typical list sizes.
- [Toggling by tapping anywhere on the card could conflict with tapping the "Déplacer" button, which lives inside the card] → Mitigate by making the button call `stopPropagation` (or making only a dedicated chevron control — not the whole card — the tap target; DownloadsTab already isn't expected to define this at spec level, this is an implementation-time choice within the constraints above).
- [`download_path` may be absent for very old/pending records where the path isn't yet known] → Fall back to the existing title-derivation logic already in place (`filepathBase`) or to `item.url`, consistent with how the component already tolerates missing fields.

## Open Questions

None — the tap-target-vs-button conflict above is resolved during implementation per the mitigation noted, without changing specs or task breakdown.
