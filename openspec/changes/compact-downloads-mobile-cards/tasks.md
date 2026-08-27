## 1. Expand/collapse state

- [ ] 1.1 Add local `expandedIds: Set<number>` state (`useState`) to `DownloadsTab`, with a toggle handler keyed by download `id`, and verify toggling one card's id adds/removes it from the set without affecting other ids
- [ ] 1.2 Add a per-card chevron/expand control (button, `aria-expanded` reflecting state) that calls the toggle handler and stops event propagation so it doesn't conflict with the "Déplacer" button, and verify clicking it toggles only that card

## 2. Mobile essential fields markup

- [ ] 2.1 Replace the desktop-only raw `item.url` line and the boxed `folder_name`/`file_name` display, for mobile, with a single line rendering `download_path` (falling back to existing `filepathBase`/`item.url` logic when `download_path` is absent), and verify the fallback still renders correctly for a download with no `download_path`
- [ ] 2.2 Add a size/progress line shown for every status: `downloaded / total (percent)` + progress bar for `downloading`/`retrying`, and fixed `file_size` for `completed`/`failed`/`pending`, with no duplicate size value rendered alongside the progress bar, and verify each status renders exactly one size/progress representation
- [ ] 2.3 Confirm title+year, status badge, error message (when failed), and the "Déplacer" button (when completed) remain rendered as before, now grouped with the fields above as the always-visible mobile section, and verify no existing desktop rendering of these fields changed

## 3. Collapsible secondary section

- [ ] 3.1 Wrap the format/resolution/duration line, the three validation badges (year, format, low quality), and the genres line in a single collapsible section keyed to the card's expanded state, and verify all three groups appear together when expanded and are absent from view when collapsed
- [ ] 3.2 Add mobile-only CSS (`index.css`, under the existing `@media (max-width: 767.98px)` block) to hide the collapsible section by default and show it when the card has the expanded class/state, including a `min-height: 44px` tappable area for the expand control
- [ ] 3.3 Add a CSS override (existing desktop breakpoint, `min-width: 768px` or equivalent) forcing the collapsible section always visible and hiding the expand control entirely on desktop, and verify desktop rendering is visually unchanged from before this change

## 4. Verification

- [ ] 4.1 Manually test in a mobile-width viewport: cards for completed, downloading, retrying, and failed downloads each show the correct essential fields and collapsed secondary section by default
- [ ] 4.2 Manually test expanding multiple cards independently, then wait for the 5s auto-refresh to fire, and verify previously-expanded cards remain expanded after the refetch
- [ ] 4.3 Manually test at a desktop-width viewport that all fields remain always visible with no expand control present, matching current behavior
- [ ] 4.4 Run the frontend build/typecheck (existing `npm run build` or equivalent) and verify it succeeds with no new TypeScript errors
