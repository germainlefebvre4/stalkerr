## Context

The dashboard is a single React 19 + Vite SPA (`frontend/src`, ~2000 lines across `App.tsx`, 10 components, 6 hooks, `services/api.ts`). No i18n library is installed today; every user-facing string is hardcoded in French, while `index.html` declares `<html lang="en">`. Backend error responses already carry a machine-readable `ErrorResponse{ Error, Message }` (see `internal/api/dto.go`), but the frontend currently reads only `Message` (backend-authored, mostly English) and, in several flows, throws its own hardcoded French fallback string instead. See `proposal.md` for the full motivation.

## Goals / Non-Goals

**Goals:**
- Ship `react-i18next` as the translation layer for the whole frontend, with English as default and French as the only other supported language.
- Make error toasts and inline dialog errors resolve through the backend's existing `ErrorResponse.error` code instead of its `message` field.
- Keep date/number formatting consistent with the active UI language.

**Non-Goals:**
- No support for languages beyond English/French, no pluralization-heavy content, no RTL layout concerns.
- No rewrite of backend error architecture beyond adding three new, narrowly-scoped error codes (`move_failed`, `override_failed`, `filter_create_failed`) to flows that currently rely on hardcoded frontend French fallbacks. Existing generic codes (`database_error`, `not_found`, `invalid_request`, ...) are left as-is.
- No change to how errors are logged server-side; `ErrorResponse.message` keeps flowing to the client for debugging/console use, it's just no longer what gets rendered to the end user.

## Decisions

### 1. `react-i18next` + `i18next`, not a hand-rolled Context
The app is small (~2000 lines) and a minimal Context+`t()` hook was considered, but `react-i18next` was chosen because: it gives free interpolation (`t('reset.success', { title })`) needed for several existing dynamic messages (e.g. "movie with id %d was reset successfully, removed %d processed lines"), it has a mature language-detection plugin, and it keeps the door open if a third language is ever requested without a rewrite. The extra dependency weight is acceptable for a dashboard app.

### 2. Language detection with English forced as the true default
`i18next-browser-languagedetector` is used only to read a stored preference; the `detection.order` is configured as `['localStorage']` only (no `navigator` detector), so a browser set to e.g. German never causes an unsupported-language fallback surprise beyond the documented "always defaults to English" contract. `fallbackLng: 'en'` and `supportedLngs: ['en', 'fr']` enforce this at the i18next-config level too, as defense in depth.

### 3. Locale key: `stalkeer_language` in `localStorage`
Mirrors the existing `stalkeer_active_tab` / `stalkeer_playlist_limit` pattern already used in the codebase (`App.tsx`, `usePlaylist.ts`) rather than introducing a new persistence mechanism (e.g. a cookie, or a backend user-preference endpoint). This is a client-only preference; there is no per-user account system to persist it against server-side.

### 4. Translation resource layout: one namespace per feature area
```
frontend/src/locales/
  en/
    common.json      (app shell, header, toasts, generic error fallback)
    playlist.json
    filters.json
    downloads.json
    logs.json
    dialogs.json      (CreateFilterDialog, MoveFolderDialog, ManualOverrideDialog)
  fr/
    (same structure)
```
Namespacing by feature (rather than one giant `translation.json`) keeps files reviewable and matches how the codebase already separates concerns into `hooks/use*.ts` + one component per tab/dialog. `common.json` is always loaded; other namespaces load on demand via `useTranslation('playlist')` etc. — negligible bundle-size win, but keeps future growth contained.

### 5. Error rendering: code → i18n key mapping in a single place
`services/api.ts` is changed to throw a small `ApiError extends Error` carrying `{ code, message }` instead of a plain `Error(message)`. Call sites (`App.tsx`, `CreateFilterDialog`, `MoveFolderDialog`, `ManualOverrideDialog`) then call `t(\`errors.${err.code}\`, { defaultValue: t('errors.generic') })` against the `common` namespace's `errors` key. Unmapped codes automatically fall back to the translated generic message — no per-call-site fallback string needed anymore, which also removes the last hardcoded French fallback strings in that file (`'La réinitialisation a échoué'`, `'La recherche a échoué'`, etc.).

Codes considered "known" and given a specific translation from day one: `not_found`, `invalid_request`, `database_error`, `tmdb_disabled`, `move_failed`, `override_failed`, `filter_create_failed`. Any other/future backend code (including ones not enumerated here) renders the generic fallback rather than requiring a frontend change every time a new code is introduced server-side.

### 6. New backend error codes are additive, not replacements
`move_failed`, `override_failed`, and `filter_create_failed` are new distinct values for the existing `ErrorResponse.Error` field on the specific handlers named in the specs — they don't change the JSON shape, don't touch unrelated endpoints, and don't remove any existing code (`database_error` etc. still covers everything else in those same handler files that isn't the specific move/override/create-filter failure path).

### 7. Date formatting: pass `i18n.language` explicitly
Every existing `toLocaleDateString()` / `toLocaleString()` call site (`PlaylistTab.tsx`, `LogsTab.tsx`) is changed to `toLocaleDateString(i18n.language)` / `toLocaleString(i18n.language)`. No date library is introduced; `Intl` already resolves `"en"` and `"fr"` correctly via the runtime's built-in locale data.

## Risks / Trade-offs

- **[Risk]** Manually extracting ~150+ hardcoded strings across 10 components is mechanical, repetitive work with a real chance of missing a string or introducing a typo in a key. → **Mitigation**: extract file-by-file, and keep a `console.warn` on missing keys enabled in dev (`i18next` `saveMissing`/`missingKeyHandler` during development only) so gaps surface immediately while clicking through the app.
- **[Risk]** The three new backend error codes are scoped narrowly (one failure path per handler); other failure branches in those same handlers still return `database_error`/`invalid_request` and will still show the generic message. → **Mitigation**: acceptable per proposal scope; broadening error-code granularity further is not required by this change and can be a follow-up if specific flows prove confusing to users.
- **[Trade-off]** Loading translation namespaces on demand adds a small amount of async complexity (a namespace not yet loaded briefly shows fallback text) versus bundling everything upfront. → Accepted: the app is small enough that eager-loading all namespaces at startup is also a reasonable simpler alternative; either is fine, but namespaces are kept regardless for file organization.
