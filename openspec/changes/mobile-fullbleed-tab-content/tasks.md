## 1. Divider token

- [x] 1.1 Add a new border token to `frontend/src/variables.css` for the KPI-to-content seam (e.g. `--border-color-strong: #e2e8f0`), distinct from the existing `--border-color: #f1f5f9` which is too close to `--bg-app` to read as a divider, and verify the new token is defined alongside the other border tokens.

## 2. Mobile tab-panel styling

- [x] 2.1 In `frontend/src/index.css`, within the `@media (max-width: 767.98px)` block, remove `.tab-panel`'s card treatment (background, border, border-radius, box-shadow) and its side padding, and verify (via browser devtools computed styles) that none of those four properties resolve to a visible value at a mobile viewport width.
- [x] 2.2 Give mobile `.tab-panel` negative side margins equal to `.app-container`'s *actual rendered* mobile side padding (`-1.5rem`, not the `-0.75rem` from the (dead) CSS override — `App.tsx:186` sets `paddingLeft`/`paddingRight: '1.5rem'` inline unconditionally, which always wins over the `index.css:897-898` class rule), the same technique already used by `.table-flush` to bleed content past a padded ancestor (`index.css:860-864`), and verify the tab content's left/right edges align with the physical viewport edges rather than with the header/KPI cards above it.
- [x] 2.3 Add a `border-top: 1px solid var(--border-color-strong)` to mobile `.tab-panel` and verify it renders as a visibly distinct hairline against `--bg-app`, directly under the KPI block (collapsed toggle row or expanded grid).

## 3. Cross-tab and regression check

- [x] 3.1 At a viewport narrower than 768px, open each of the four tabs (Playlist, Filters, Logs, Downloads) and verify each one's content reaches the viewport's left/right edges with the new top divider, while `mobile-list-card`, `download-card`, and `filter-card` items keep their existing background/border/radius unchanged.
- [x] 3.2 At a viewport at or above 768px, verify `.tab-panel` is pixel-identical to current desktop behavior (card background, border, radius, shadow, padding all still present).
- [x] 3.3 Run `npm run lint` and `npm run build` in `frontend/` and verify both succeed with no new warnings or errors.
