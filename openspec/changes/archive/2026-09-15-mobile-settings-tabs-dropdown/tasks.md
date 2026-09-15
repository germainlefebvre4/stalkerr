## 1. Mobile tab dropdown

- [x] 1.1 Import `useIsMobile()` into `ConfigurationPage.tsx` and branch tab navigation rendering: desktop keeps `Tabs.List`/`Tabs.Trigger`, mobile renders a `<select className="custom-select">` wired to `setActiveSettingsTab`
- [x] 1.2 Build the mobile `<select>`'s option labels from the same data as `renderTabBadge` (override count, restart-required), serialized as text (e.g. `⚠ Intégrations (3)`), with no count suffix when zero and no marker when no restart is required — verify by toggling settings overrides and checking the option text updates
- [x] 1.3 Remove the `.settings-tabs-list` / `.segmented-tabs-trigger` mobile `overflow-x: auto` override in `index.css` (no longer reachable once the pill bar is desktop-only) and verify no horizontal scrollbar appears on the Configuration page at a mobile viewport width in the browser
- [x] 1.4 Verify selecting a tab from the mobile dropdown updates the active tab, the page URL, and matches the tab selected via keyboard/URL restore behavior already covered by existing `ConfigurationPage.test.tsx` tests

## 2. Technical value overflow fix

- [x] 2.1 Add a scoped wrapping utility class (e.g. `.settings-value-wrap`) to `index.css` using `overflow-wrap: anywhere`, applied only to the value elements, not whole cards
- [x] 2.2 Apply the class to the `<code>` elements rendering `source.file_path` and `source.url` in `M3uSourcesSection.tsx` and verify a long, space-free URL wraps within its card instead of overflowing
- [x] 2.3 Apply the class to the paths line in `SystemStatusSection.tsx`'s `DiskUsageRow` and verify a long, space-free mount path wraps within its card instead of overflowing
- [x] 2.4 Manually verify in the browser, at a mobile viewport width, that the Contenu tab (with a long-URL M3U source) and the Système tab (with a long disk path) show no page-level horizontal scrollbar

## 3. Tests

- [x] 3.1 Add/update a test in `ConfigurationPage.test.tsx` covering the mobile dropdown rendering and tab-switch behavior (mocking `useIsMobile()` as it's already mocked/used elsewhere in this test suite)
- [x] 3.2 Run the frontend test suite and verify it passes with no regressions to existing Configuration page tests
