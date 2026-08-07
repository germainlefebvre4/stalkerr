## Context

Three independent pieces of frontend state need to survive a page refresh: the Playlist filters/pagination (already implemented ad hoc in `usePlaylist.ts` via its own `useEffect` that rebuilds `URLSearchParams` from scratch and calls `history.replaceState`), the active tab (currently `localStorage`-only, in `App.tsx`), and the Downloads filters (currently not persisted at all, in `useDownloads.ts`). See proposal.md - Why for why writing three separate from-scratch `URLSearchParams` rebuilds is unsafe: whichever effect fires last in a render cycle wins and drops the other two's query params.

## Goals / Non-Goals

**Goals:**
- One hook that all URL-backed state goes through, safe against being called from multiple independent React hooks/components without a shared store.
- No change to Playlist's existing URL param names or observable behavior.
- Downloads' new params can't collide with Playlist's existing ones.

**Non-Goals:**
- Not building a general-purpose routing library or introducing React Router.
- Not persisting ephemeral UI state (open drawer, selected row, scroll position, "go to page" input) — only state that already drives a data fetch or a filter.
- Not adding pagination to Downloads (it has none today; out of scope).

## Decisions

### 1. Read-fresh / merge / write, no shared store
`useURLState` does not hold `URLSearchParams` in a module-level singleton or React context. Every write reads `window.location.search` fresh at call time, sets/deletes only its own key(s), and calls `history.replaceState`. Because `history.replaceState` is synchronous and React runs effects within one commit sequentially, an effect that writes after another effect in the same commit sees that effect's write already applied. This removes the clobbering race without introducing a shared registry, context provider, or extra re-renders for unrelated consumers.

Alternative considered: a singleton store (e.g. a module-level `URLSearchParams` + subscriber list) that every hook reads/writes through. Rejected — it adds a second source of truth that must itself stay in sync with `window.location`, for no benefit over reading the live URL directly.

### 2. Hook shape: schema object, single hook per domain
```
const [state, patchState] = useURLState({
  page:   { default: 1,  parse: parseIntOr(1),        serialize: String },
  limit:  { default: 50, parse: parseIntOr(50),        serialize: String },
  search: { default: '', parse: String,                serialize: identity },
  ...
});
patchState({ page: 2 });        // merges into current URL state + writes only changed/valid keys
```
Each call to `patchState` reads the live URL, applies the patch, drops keys whose value equals their `default` (matches today's "omit when default" convention so URLs stay short), and writes back. A key is omitted from the parsed result (falls back to `default`) if `parse` throws or the value fails an optional `isValid` check — mirrors `usePlaylist`'s current `VALID_LIMITS`/`VALID_*` allow-list guards.

This one hook shape covers both the multi-key case (Playlist's 6 keys, Downloads' 3 keys) and the single-key case (active tab) — a 1-entry schema is just the degenerate case, so no separate single-key API is needed.

Alternative considered: a single-key `useURLParam(key, opts)` plus a multi-key `useURLState(schema)` as two APIs. Rejected for now — Playlist and Downloads both need the multi-key form, and the active tab is a 1-entry schema, so shipping only the schema form avoids maintaining two APIs for one behavior.

### 3. Namespacing to avoid collisions
Downloads' new params are named `dlStatus`, `dlType`, `dlProblem` (not `status`/`type`/`problem`) because Playlist already owns `type` (content type) and `state` (pipeline state) — reusing bare names would silently cross-wire the two tabs' filters if a link is shared. The active tab uses `tab`. No existing Playlist param uses that name.

### 4. Migrate `usePlaylist` onto the shared hook now, not later
Per the proposal, `usePlaylist.ts` is migrated onto `useURLState` in this same change rather than left on its bespoke implementation. Rationale: leaving it as a second, parallel implementation of the same read-URL/write-URL behavior is the exact duplication this change exists to remove, and the migration is behavior-preserving (same param names, same default-omission rule, same page-resets-on-filter-change rule) so the regression risk is contained to "does it still read/write the same six params the same way," which is directly testable.

## Risks / Trade-offs

- [Risk] A subtle behavioral difference slips in while migrating `usePlaylist`'s bespoke logic (e.g. the "reset page to 1 only when a filter actually changed, not on initial mount" guard) onto the generic hook → Mitigation: port that guard as-is inside `usePlaylist` (it stays a concern of *when* to patch state, not of the URL hook itself), and keep/extend the existing test coverage for page-reset-on-filter-change.
- [Risk] Two `patchState` calls from different hooks firing in the *same* React commit but in an order where one hasn't flushed before the other reads → Mitigation: both read/write synchronously against `window.location` inside the effect body (no `setTimeout`/microtask deferral), so within one commit's sequential effect execution this is safe by construction; there is no async gap for a stale read to occur.
- [Trade-off] `history.replaceState` is used (never `pushState`), consistent with today's Playlist behavior — browser back/forward does not step through filter changes. Acceptable since that's the existing, already-accepted behavior for Playlist.

## Migration Plan

1. Add `useURLState` with no consumers yet (pure addition, no behavior change possible).
2. Migrate `usePlaylist` onto it; verify identical URL read/write behavior (same param names, defaults, and page-reset rule).
3. Add `tab` to `App.tsx` via the shared hook, alongside the existing `localStorage` write (both are kept — `localStorage` remains the fallback when no `tab` param is present, e.g. a bookmarked bare URL).
4. Add `dlStatus`/`dlType`/`dlProblem` to `useDownloads` via the shared hook.

No backend changes, no data migration, no feature flag needed — this is a client-only, additive-then-refactor change. Rollback is reverting the frontend deploy.
