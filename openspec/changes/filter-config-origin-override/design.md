## Context

See proposal.md - Why. Current mechanics, for reference:

- `config.yml`'s `filter.group_title` / `filter.tvg_name` sections are the origin config, loaded once at startup via `filter.Manager.LoadFromConfig()` and never exposed through the API.
- `filter_configs` DB rows (`models.FilterConfig`) are always `is_runtime = true` — created only through `POST /api/v1/filters`, which today does not check for an existing row on the same attribute, so multiple runtime rows can accumulate per attribute.
- `filter.Manager.Matches()` picks the runtime filter set over the config filter set whenever at least one runtime filter exists for the attribute, and if several runtime filters exist for the same attribute they're ANDed together. This change doesn't alter that "runtime wins" resolution rule, it just makes "runtime" mean exactly one filter per attribute so the picture is always unambiguous.

## Goals / Non-Goals

**Goals:**
- Make the origin `config.yml` patterns readable through the API without touching the database.
- Guarantee at most one active runtime override per attribute, enforced server-side (not just in the UI), so the effective config for an attribute is always a single, well-defined set of patterns.
- Let the frontend show origin vs. active-override provenance per attribute, and let users seed a new filter from whichever is currently active.

**Non-Goals:**
- No change to how patterns within a single filter are evaluated (include/exclude regex semantics stay as-is).
- No support for more than two attributes (`group_title`, `tvg_name`) — that set stays fixed.
- No UI for editing `config.yml` itself; origin config remains read-only from the app's perspective.
- No new "edit filter" UI flow beyond what "Reprendre la config actuelle" + create covers; the existing `PATCH /api/v1/filters/:id` enforcement is a server-side safety net, not a newly exposed UI capability.

## Decisions

**Expose origin config via a new endpoint, not by embedding it in `GET /api/v1/filters`.**
`GET /api/v1/filters` returns DB rows (`[]models.FilterConfig`) and is used elsewhere (e.g., future filter management) as a straightforward list of runtime rows. Origin config has a different shape (fixed two attributes, always present, never has an id/timestamps) and a different source (`config.Get().Filter`, not the DB). A dedicated endpoint, `GET /api/v1/filters/system`, keeps both shapes simple: `{"group_title": {"include_patterns": [...], "exclude_patterns": [...]}, "tvg_name": {...}}`. Alternative considered: add an `origin` object into every `GET /api/v1/filters` response — rejected because origin config isn't per-row, it's per-attribute, and repeating it on every row (or returning it once awkwardly) is more confusing than a second small endpoint.

**Enforce "one active override per attribute" in the handler, not with a DB unique constraint.**
`createFilter`/`updateFilter` will, inside the same transaction as the insert/update, delete any other `filter_configs` row with `is_runtime = true` and the same `attribute` before committing the new/updated row. Alternative considered: a unique index on `attribute` (partial, `WHERE is_runtime`) with a migration and an upsert-style "delete conflicting row, retry" loop — rejected as unnecessary ceremony for a two-valued attribute column; the application-level delete-then-write inside a transaction gives the same guarantee without a schema migration, and this is the only write path that creates runtime filters.

**No new DB column for "is this config file-based or overridden" — keep origin purely computed from `config.yml`.**
Origin patterns are never written to the DB. The frontend combines `GET /api/v1/filters/system` (origin) with `GET /api/v1/filters` (active overrides, now guaranteed ≤1 per attribute) client-side to render each attribute's card. This avoids duplicating `config.yml` content into the database and keeps the origin config trivially always in sync with the deployed file.

**"Reprendre la config actuelle" is a form-only action, not a new API call.**
The dialog already has access to both the origin config and the active overrides (fetched by `FiltersTab`/`useFilters` for the list view) before it opens, so clicking the button just copies the relevant attribute's patterns into local form state. No new endpoint or request is needed for this.

## Risks / Trade-offs

- **[Risk] Deploying the "one active override per attribute" rule silently deletes a pre-existing second/third runtime filter on an attribute that currently has several stacked (ANDed) filters** → Mitigation: on migration, keep the AND-stacking behavior available as a manual read (log or one-time report of which rows would be dropped) is out of scope for this change; document in the changelog/release notes that only the most recently created runtime filter per attribute survives after upgrade, since `filter.Manager.Matches()` already treats "runtime present" as replacing origin regardless of count — this change only removes the ability to stack multiple runtime filters, it doesn't change whether origin applies.
- **[Risk] Auto-replacing an active override on create is destructive with no undo** → Mitigation: the frontend warns before submission (per proposal); the replaced filter is deleted, not recoverable, which is consistent with how `DELETE /api/v1/filters/:id` already behaves today.
- **[Trade-off] A second endpoint (`/filters/system`) instead of one unified response** → simpler payload shapes for both origin and runtime data, at the cost of the frontend needing two fetches instead of one; acceptable since both are small, infrequent, low-latency reads on a lightly-trafficked settings page.
