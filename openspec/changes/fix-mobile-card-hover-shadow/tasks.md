## 1. Fix the hover rule

- [ ] 1.1 In `frontend/src/index.css`, wrap the `.card:hover` rule (currently unconditional, ~lines 72-75) in an `@media (hover: hover) and (pointer: fine)` guard, and verify no other selector in the file still applies a box-shadow or transform to `.card`/`.tab-panel` outside of that guard

## 2. Verify the fix

- [ ] 2.1 In Chrome DevTools with a mobile device emulation profile (touch input) at a viewport below 768px, tap into each tab's content area (Home, Downloads, Playlist, Filters, Logs, Errors, Radarr/Sonarr) and confirm no box-shadow or upward lift appears or persists after the tap, matching the `frontend-responsive-layout` "Full-Bleed Tab Content on Mobile" spec
- [ ] 2.2 On desktop with a real mouse, confirm the tab content area still lifts and shows the shadow on `:hover`, unchanged from current behavior
- [ ] 2.3 Run the frontend test suite (`npm test` in `frontend/`) and confirm it still passes
