## Context

`ManualOverrideDialog` (`frontend/src/components/ManualOverrideDialog.tsx`) currently drives one `ProcessedLine` at a time through `POST /api/v1/items/:id/override`. That endpoint already auto-extracts season/episode per item from its own `tvg_name` (via the Go `classifier`) whenever the request omits them, and the backend TMDB client caches responses by request URL, so calling the endpoint repeatedly with the same `tmdb_id` is cheap after the first call (see proposal.md - Impact). The playlist table (`PlaylistTab.tsx`) is server-paginated; `App.tsx` already holds the current page's items in the `playlist` array and does not currently pass it to the dialog. See proposal.md - Why for the underlying pain point.

## Goals / Non-Goals

**Goals:**
- Let one TMDB series choice be applied to several playlist entries in one confirmation, without a new backend endpoint.
- Keep the false-positive/false-negative correction fully in the user's hands (freely checkable list), not a fully-automatic bulk apply.
- Keep the candidate search cheap (no extra network round trip): computed client-side from data already loaded for the visible page.

**Non-Goals:**
- Finding candidates beyond the currently loaded playlist page (see proposal.md - explicitly out of scope; users needing wider coverage already have a page-size control and name filter to bring more episodes into view first).
- Resetting or re-triggering pipeline `State` for already-downloaded/failed items after a (re)association - out of scope, matches today's single-item override behavior.
- A dedicated bulk backend endpoint - the per-item endpoint's cost is already amortized by the existing TMDB response cache.

## Decisions

**Reuse the existing single-item endpoint in a sequential client-side loop, not a new bulk endpoint.**
Each `POST /items/:id/override` call already does the right per-item thing (own season/episode auto-extraction, own `ManualMapping` row). The only genuinely repeated work across a batch - the TMDB series detail + external-IDs lookup - is already memoized in the backend TMDB client by full request URL, so N calls with the same `tmdb_id` cost one real TMDB round trip. A new bulk endpoint would duplicate that logic server-side for no measurable benefit. Calls are issued sequentially (not `Promise.all`) so the per-item result can be attributed correctly for the outcome summary and so a systemic failure (e.g. TMDB briefly down) doesn't fire N simultaneous requests.

**Candidate detection reuses the dialog's existing title-cleaning regex, compared for equality, computed from already-loaded data.**
The same cleaning already applied to build the search query (strip provider prefixes, quality/codec tags, `SxxExx` markers) is applied to every other `tvshows` item on the current page; a candidate is pre-checked when its cleaned title equals the opened item's cleaned title. This is a heuristic, not a guarantee - which is why every candidate (matched or not) stays independently checkable, covering both false positives (uncheck) and false negatives (check manually). No new API call is introduced: the candidate pool is the `playlist` array `App.tsx` already fetched for the visible page.

**Batch outcome is reported as a single aggregated summary, not per-request toasts.**
Sequential requests resolve one by one; the dialog accumulates successes/failures and shows one summary (e.g. "11/12 associés") once the loop finishes, naming the failed item(s) if any. The table refresh (`onSuccess` → `fetchPlaylist()`) fires once after the loop, reflecting whatever subset actually succeeded.

**Wiring: pass the current page's `playlist` array into `ManualOverrideDialog` as a new prop.**
`App.tsx` already holds it; no new state or fetch is needed. The dialog filters it down to `content_type === 'tvshows'` candidates excluding the opened item.

## Risks / Trade-offs

- [Risk] A partially-failed batch leaves the playlist in a mixed state (some entries corrected, some not) → Mitigation: the outcome summary explicitly names the failed item(s) so the user can retry just those from the normal single-item flow; this mirrors how any multi-step network operation in this codebase already surfaces partial failure (e.g. downloads list shows per-item status).
- [Risk] Sequential requests mean a large batch (e.g. a 24-episode season) takes roughly 24× one request's latency to fully confirm → Mitigation: acceptable given today's TMDB cache makes each request cheap after the first, and the user already controls batch size via the page-size filter; a visible per-item progress indicator keeps the wait legible.
- [Risk] The cleaned-title equality heuristic under- or over-matches on unusual titles (e.g. inconsistent punctuation across episodes) → Mitigation: this is why the candidate list is fully user-editable rather than an automatic bulk-apply; incorrect pre-selection is a one-click correction, not a silent failure.
