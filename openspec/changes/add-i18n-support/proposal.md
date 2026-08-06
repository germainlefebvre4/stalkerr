## Why

The dashboard's UI text is hardcoded in French across every component (`App.tsx`, dialogs, tabs), while `index.html` declares `<html lang="en">` and several backend error messages surfacing in toasts are in English — an already-inconsistent, single-language experience. We need a proper i18n layer so the UI defaults to English with French available as a second language, and so error messages resolve consistently regardless of which language produced them on the backend.

## What Changes

- Introduce `react-i18next` + `i18next` in the frontend and extract every hardcoded string (in `App.tsx`, the 10 components, and the hooks that produce user-facing status labels) into locale resource files.
- Default language is English (`fallbackLng: 'en'`); available languages are English and French. The active language is selectable via a switcher in `FloatingHeader`, persisted to `localStorage` (same pattern as the existing `stalkeer_active_tab` key).
- Frontend error handling switches from displaying the raw `err.message` string to resolving the backend's existing `ErrorResponse.error` code through the translation catalog, with a generic translated fallback for any unmapped code.
- Backend gains more specific `ErrorResponse.error` codes for the three flows that currently rely on frontend-hardcoded French fallback text: folder move, manual override, and filter creation — replacing/supplementing their current generic codes so the frontend can render a precise translated message instead of a generic one.
- `toLocaleDateString`/`toLocaleString` call sites pass the active i18n language explicitly, so displayed date/time formats follow the selected UI language instead of the browser's locale.
- `<html lang>` is updated to reflect the active language.

## Capabilities

### New Capabilities
- `frontend-i18n`: Translation infrastructure (react-i18next setup, `en`/`fr` resource files), the default/available language contract, the language switcher control and its persistence, locale-aware date formatting, and the contract that frontend error toasts resolve backend error codes through the translation catalog (with a generic fallback for unmapped codes).

### Modified Capabilities
- `api-media-management`: The "Move Media Parent Folder" requirement's error response gains a distinct machine-readable error code (e.g. `move_failed`) instead of relying on the frontend's hardcoded French fallback message.
- `tmdb-manual-override`: The "Manual Override API Endpoint" requirement's error response gains a distinct machine-readable error code (e.g. `override_failed`) for override-specific failures, alongside the existing `tmdb_disabled` code.
- `frontend-filters-management`: The "Create Filter Configuration" requirement's error response gains a distinct machine-readable error code (e.g. `filter_create_failed`).

## Impact

- **Frontend**: new dependencies (`react-i18next`, `i18next`, and a language-detection helper); new `src/i18n/` setup and `src/locales/en/*.json` + `src/locales/fr/*.json` resource files; every component and hook that currently emits user-facing text is refactored to use `useTranslation`/`t()`; `services/api.ts` is refactored to surface the backend's error code alongside its message so callers can translate it.
- **Backend**: `internal/api/handlers_frontend.go` (folder move), the manual-override handler, and the filter-creation handler in `internal/api/handlers.go` gain new `ErrorResponse.error` codes. The existing generic codes (`database_error`, `not_found`, `invalid_request`, ...) are otherwise unchanged.
- No database schema changes.
- No breaking removal: the `ErrorResponse.message` field is kept for logging/fallback; only the `error` code gains granularity for these three flows.
