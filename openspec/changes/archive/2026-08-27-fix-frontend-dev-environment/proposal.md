## Why

The frontend dev environment is broken in three independent ways: the app shows a blank page at runtime because React throws on a `react`/`react-dom` version mismatch, `npm run build` fails because TypeScript can't type the `./index.css` side-effect import (no `vite-env.d.ts`), and `npm run lint` fails because `eslint` was never installed or configured. None of this is new work in progress — all three predate the current uncommitted changes and block basic local development (`local-dev`'s "Frontend Make targets" requirement already claims `make front-build` and `make front-lint` invoke working npm scripts; today they don't).

## What Changes

- Pin `react` and `react-dom` (and their `@types/*` packages) to the same version range so React's runtime version check stops throwing and the app renders again.
- Add `frontend/src/vite-env.d.ts` with the Vite client type reference so `tsc` (and therefore `npm run build`) stops failing on the CSS side-effect import.
- Install and configure `eslint` (flat config, TypeScript + React rules consistent with the existing `tsconfig.json`) so `npm run lint` runs and passes.

## Capabilities

No capability requirements change. This is a bug fix restoring already-documented dev tooling behavior (`local-dev`'s existing requirement that `make front-build` / `make front-lint` work) rather than introducing new behavior — see `skip_specs: true` in `.openspec.yaml`.

## Impact

- `frontend/package.json`, `frontend/package-lock.json`: dependency version changes, new `eslint` devDependency.
- `frontend/src/vite-env.d.ts`: new file.
- New eslint flat config file (e.g. `frontend/eslint.config.js`).
- No API, schema, or user-facing behavior changes.
