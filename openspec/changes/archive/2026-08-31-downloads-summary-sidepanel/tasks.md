## 1. Summary list component

- [x] 1.1 Create `frontend/src/components/DownloadsSummaryList.tsx` taking `downloads`, `loading`, `onRowClick`; render the desktop table (`table-flush`/`custom-table`, using `clickable-row`) with columns for type icon + title (+year, with fallback to file name / URL), status badge, and a compact progress indicator for `downloading`/`retrying` items — verify with `npm run build` (frontend) succeeding with no type errors
- [x] 1.2 In the same component, add the `useIsMobile()` branch rendering `mobile-list-card` items (title/year as main line, status badge), mirroring `PlaylistItemsTable.tsx`'s mobile branch — verify by rendering the component in a unit test at a mobile viewport and asserting `mobile-list-card` markup appears
- [x] 1.3 Add a `DownloadsSummaryList.test.tsx` covering: title fallback when `content.title` is absent, progress indicator only shown for `downloading`/`retrying`, and `onRowClick` firing with the clicked item's id — verify with `npm test -- DownloadsSummaryList`

## 2. Sidepanel wiring in DownloadsTab

- [x] 2.1 In `DownloadsTab.tsx`, replace the `expandedIds`/card-list state with `selectedId: number | null`, derive `selectedItem = downloads.find(d => d.id === selectedId) ?? null`, and render `DownloadsSummaryList` with `onRowClick={item => setSelectedId(item.id)}` in place of the current card-mapping JSX — verify the tab renders with no `download-card` markup left in the DOM
- [x] 2.2 Add the Radix `Dialog.Root`/`Dialog.Content` drawer (reusing `drawer-overlay`/`drawer-content` CSS, same as `PlaylistTab.tsx`) with `open={!!selectedItem}` and `onOpenChange={(open) => !open && setSelectedId(null)}` — verify clicking a row opens the drawer and the close button/overlay closes it
- [x] 2.3 Port the existing card's conditional detail blocks into the drawer body: status + progress section, file section (folder/file name, full path or URL fallback), technical specs section (format, resolution, size, duration, completion date), validation badges section (year/format/low-quality), genres — verify each section only renders when its underlying data is present (no empty boxes), matching `specs/downloads-details-sidepanel/spec.md`
- [x] 2.4 Move the error message block into the drawer, gated on `status === 'failed'` exactly as before — verify by adapting the existing "does not show the error banner for a completed download" / "shows the error banner for a failed download" assertions in `DownloadsTab.test.tsx` to first open the drawer (click the row) before asserting
- [x] 2.5 Move the "Move" and "Rename" buttons (calling the existing `onOpenMoveDialog`/`onOpenRenameDialog` props, unchanged) into the drawer's actions section, gated on `selectedItem.status === 'completed'` — verify the buttons no longer appear in `DownloadsSummaryList` rows and do appear in the drawer for a completed item
- [x] 2.6 Verify the auto-close behavior: when `selectedId` is set but no longer present in `downloads` (e.g. simulated by re-rendering with a filtered/updated `downloads` prop missing that id), the drawer closes — cover with a unit test asserting the dialog is no longer in the document

## 3. i18n

- [x] 3.1 Add any new translation keys needed for drawer section titles (status/progress, file, technical, validation, error, actions) to `frontend/src/locales/en/downloads.json` and `frontend/src/locales/fr/downloads.json`, reusing existing keys (`move`, `rename`, `folder`, `format`, badges, etc.) wherever their text still applies — verify with `npm run build` (i18next key usage has no missing-key console warnings in dev mode)

## 4. Cleanup and verification

- [x] 4.1 Remove now-unused CSS (`download-card` internals no longer referenced, `download-expand-btn`, `download-secondary`/`download-secondary--expanded`, `download-tech-row`, `download-path-mobile`, `download-size-mobile`, `download-progress-desktop`, `download-url-desktop` etc.) from `frontend/src/index.css` if no longer referenced anywhere — verify with a repo-wide grep confirming zero remaining references before deleting each class
- [x] 4.2 Run the full frontend test suite and confirm no regressions — verify with `npm test`
- [x] 4.3 Manually exercise the Downloads tab in the running app: summary rows for completed/downloading/failed items, opening the drawer for each, live progress update on an active download left open across a poll tick, Move/Rename from the drawer, and mobile viewport layout — verify by observing the described behavior in the browser
