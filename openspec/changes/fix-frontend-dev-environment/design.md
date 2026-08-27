## Context

`frontend/package.json` currently has `"react": "^19.2.8"` and `"react-dom": "^19.0.0"`, which resolve to `react@19.2.8` / `react-dom@19.2.7` in `node_modules`. React 19 throws at runtime (not just a warning) when the two don't match exactly, which is the blank-page bug (see proposal.md - Why). There is no `frontend/src/vite-env.d.ts` and no eslint installation or config at all - `npm run lint` calls a binary that isn't a dependency.

## Goals / Non-Goals

**Goals:**
- Make `react` and `react-dom` resolve to the same version, and keep them from drifting apart again on routine dependency bumps.
- Make `npm run build` (`tsc && vite build`) pass with zero new TypeScript errors.
- Make `npm run lint` actually run and pass on the current source tree.

**Non-Goals:**
- Auditing or fixing every eslint rule violation the new config might theoretically allow for future code - only get the current tree passing.
- Changing React major version or any other dependency not implicated in these three failures.
- Adding CI enforcement for build/lint (out of scope; local-dev's Make targets already exist).

## Decisions

**Version pin: bump `react-dom` up to match `react`, not the reverse.**
Set both `"react": "^19.2.8"` and `"react-dom": "^19.2.8"` (and `@types/react`/`@types/react-dom` to the matching `^19.2.x` pair). Alternative considered: downgrade `react`'s range to `^19.0.0` to match the looser `react-dom` range - rejected because it would re-widen the range that already caused the drift (the next `npm install` could re-resolve them apart again the same way). Anchoring both to the same tight caret range keeps them locked together going forward.

**eslint: flat config matching Vite's official React+TS template.**
Add `eslint.config.js` using `typescript-eslint`, `eslint-plugin-react-hooks`, and `eslint-plugin-react-refresh`, mirroring what `npm create vite@latest -- --template react-ts` scaffolds today. Alternative considered: a minimal custom config with just `@eslint/js` - rejected because the existing `tsconfig.json` already uses strict TS + React JSX conventions that the standard template's rule set is designed to match, and using the ecosystem-standard config is less to maintain than a bespoke one.

**vite-env.d.ts: standard single-line reference.**
`/// <reference types="vite/client" />`, matching what Vite scaffolds by default. No alternative worth considering - this is the standard fix for this exact TS2882 error.

## Risks / Trade-offs

- [Bumping react-dom to 19.2.8 could theoretically introduce its own regressions] → Same-minor-version bump within an already-adopted major/minor line (19.2.x); low risk, and the app is currently unusable (blank page) without it.
- [New eslint config may surface pre-existing lint violations in files untouched by this change] → Fix any that surface in currently-checked-in source so `npm run lint` exits 0; do not silence rules just to pass.
