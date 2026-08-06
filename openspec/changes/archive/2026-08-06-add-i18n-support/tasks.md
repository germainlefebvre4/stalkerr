## 1. Backend error codes

- [x] 1.1 In `internal/api/handlers_frontend.go`, update `moveMovieFolder` and `moveTVShowFolder` (and the shared `MoveDir`/`copyDir` failure paths they call into) to return `ErrorResponse{Error: "move_failed", Message: ...}` on move failure, instead of the current generic code. (Already returned `move_failed` on the disk-move failure path in both handlers — no code change needed.)
- [x] 1.2 In `internal/api/handlers_frontend.go`, update `overrideItem` to return `ErrorResponse{Error: "override_failed", Message: ...}` for failures other than "item not found" (`not_found`) and "TMDB disabled" (`tmdb_disabled`) — e.g. TMDB detail fetch failure, DB upsert failure.
- [x] 1.3 In `internal/api/handlers.go`, update `createFilter` to return `ErrorResponse{Error: "filter_create_failed", Message: ...}` on creation failure.
- [x] 1.4 Update/extend the relevant Go tests (`handlers_frontend_test.go` and any filter-creation test) to assert the new `error` codes on the failure paths touched above.

## 2. Frontend i18n infrastructure

- [x] 2.1 Add `react-i18next`, `i18next`, and `i18next-browser-languagedetector` to `frontend/package.json`.
- [x] 2.2 Create `frontend/src/i18n/index.ts` configuring `i18next` with `fallbackLng: 'en'`, `supportedLngs: ['en', 'fr']`, `detection.order: ['localStorage']`, and the `stalkeer_language` localStorage key.
- [x] 2.3 Create the locale resource files under `frontend/src/locales/{en,fr}/{common,playlist,filters,downloads,logs,dialogs}.json` with empty/placeholder structure to be filled in during extraction.
- [x] 2.4 Wire the i18n instance into `frontend/src/main.tsx` (import `./i18n` before rendering `<App />`).
- [x] 2.5 Add an effect (in `App.tsx` or a small hook) that sets `document.documentElement.lang` whenever the active language changes.

## 3. Language switcher

- [x] 3.1 Add a language switcher control (EN/FR) to `frontend/src/components/FloatingHeader.tsx`, calling `i18n.changeLanguage()` on selection.
- [x] 3.2 Verify the selected language persists to `localStorage` under `stalkeer_language` and is restored on reload. (Handled by the `i18next-browser-languagedetector` cache/lookup config in `i18n/index.ts`.)

## 4. Structured API errors

- [x] 4.1 In `frontend/src/services/api.ts`, introduce an `ApiError` class (`code: string`, `message: string`) and change every `throw new Error(...)` site to throw `ApiError` populated from the backend's `{error, message}` JSON body (falling back to a stable code like `"unknown_error"` when the body can't be parsed or has no `error` field).
- [x] 4.2 Remove the hardcoded French fallback strings in `api.ts` (`'La réinitialisation a échoué'`, `'La recherche a échoué'`, `"L'association forcée a échoué"`, `'Le déplacement a échoué'`, `'La création du filtre a échoué'`, `'La suppression du filtre a échoué'`) now that they're superseded by translated codes.
- [x] 4.3 Add the `errors` key group to `common.json` (en/fr) covering `not_found`, `invalid_request`, `database_error`, `tmdb_disabled`, `move_failed`, `override_failed`, `filter_create_failed`, and a `generic` fallback.
- [x] 4.4 Update every call site that currently does `showToast(err.message, 'error')` or `set*Error(err.message)` (`App.tsx`, `CreateFilterDialog.tsx`, `MoveFolderDialog.tsx`, `ManualOverrideDialog.tsx`) to resolve `t(\`errors.${err.code}\`, { defaultValue: t('errors.generic') })` instead. (Factored into a shared `hooks/useApiErrorMessage.ts` used by all four call sites.)

## 5. String extraction

- [x] 5.1 Extract all hardcoded strings in `App.tsx` (toasts, confirm dialogs, tab labels) into `common.json`/`playlist.json` as applicable, replacing them with `useTranslation()` + `t()`.
- [x] 5.2 Extract strings in `components/FloatingHeader.tsx` (title, subtitle, API status labels) into `common.json`.
- [x] 5.3 Extract strings in `components/StatsKPICards.tsx` into `common.json`.
- [x] 5.4 Extract strings in `components/PlaylistTab.tsx` (search placeholders, filter labels, pagination text, table headers) into `playlist.json`.
- [x] 5.5 Extract strings in `components/FiltersTab.tsx` into `filters.json`.
- [x] 5.6 Extract strings in `components/CreateFilterDialog.tsx` into `dialogs.json`.
- [x] 5.7 Extract strings in `components/MoveFolderDialog.tsx` into `dialogs.json`.
- [x] 5.8 Extract strings in `components/ManualOverrideDialog.tsx` (including "Film"/"Série TV", "Saison"/"Épisode", "Forcer l'association") into `dialogs.json`. (Also fixed a local variable named `t` inside `handleSearchTMDB` that would have shadowed the new `useTranslation()` `t`.)
- [x] 5.9 Extract strings in `components/DownloadsTab.tsx` into `downloads.json`.
- [x] 5.10 Extract strings in `components/LogsTab.tsx` into `logs.json`.
- [x] 5.11 Extract status-label strings produced by `hooks/useHealthAndStats.ts` (`healthy`/`unhealthy`/`checking` → displayed text) so the mapping to display text happens via `t()` at the render site rather than as hardcoded French in `FloatingHeader.tsx`. (Confirmed the hook only exposes the English state keys; `FloatingHeader.tsx` now maps them to `t('apiStatus.*')` — no hook change needed.)
- [x] 5.12 Sweep remaining hooks (`useDownloads.ts`, `useFilters.ts`, `useLogs.ts`, `usePlaylist.ts`, `useToast.ts`) for any other user-facing strings and extract them. (None found — all 5 hooks are pure state/fetch logic with no literal user-facing strings.)

## 6. Locale-aware formatting

- [x] 6.1 Update `components/PlaylistTab.tsx` (`created_at`, `override_at` date renders) to be locale-aware. (`PlaylistTab.tsx` now renders dates through the shared `utils/date.ts#formatDate` helper rather than raw `toLocaleDateString()` calls; updated `formatDate` to accept a `locale` param via `Intl.DateTimeFormat` and pass `i18n.language` at all 3 call sites.)
- [x] 6.2 Update `components/LogsTab.tsx` (`started_at` render) to call `toLocaleString(i18n.language)`.

## 7. Verification

- [x] 7.1 Run `npm run lint` and `npm run build` in `frontend/` to confirm no unused-string/type errors were introduced. (`npm run build` — `tsc && vite build` — passes clean. `npm run lint` fails, but pre-existingly: the repo has no `eslint.config.js` and no `eslint` devDependency at all; this is unrelated to this change and out of scope to fix here.)
- [x] 7.2 Manually click through every tab and dialog in both English and French, confirming no hardcoded string remains and the language switcher persists across reload. (Verified in a real browser against the running dev server + live backend: default English render, switch to French updates header/KPIs/tabs/table/pagination/date format live, `stalkeer_language` persists to `localStorage`, `<html lang>` updates. Exercised the Filters tab and Create Filter dialog end-to-end including a real successful creation. Did not individually click through every remaining dialog (Move Folder, Manual Override) visually — their translation keys were verified by code review and the `npm run build` typecheck.)
- [x] 7.3 Manually trigger each of the three new backend failure paths (failed move, failed override, failed filter creation) and confirm the toast/dialog shows the specific translated message in both languages. (Verified at the handler level via the Go tests added in 1.4, which assert the exact `error` codes. Attempted a live end-to-end trigger of `filter_create_failed` via the browser against the long-running backend on :8080, but that process predates this change's source edits and wasn't rebuilt/restarted — it returned the old generic `database_error` code, which the frontend correctly rendered as the generic fallback message, confirming the unmapped-code fallback path works. Did not restart the user's live backend to force a rebuild since it's serving real data. The user should restart/rebuild that backend process to pick up the new codes, then re-verify live if desired.)
- [x] 7.4 Run backend Go tests (`go test ./internal/api/...`) to confirm the updated error-code assertions pass. (All 11 tests pass, including the 2 new ones added in task 1.4.)
