## 1. Fix react/react-dom version mismatch

- [x] 1.1 In `frontend/package.json`, set `"react-dom": "^19.2.8"` and align `@types/react-dom` to the matching `^19.2.x` range, then run `npm install` and verify `npm ls react react-dom` shows the same resolved version for both.
- [x] 1.2 Start the dev server and load the app in a browser; verify the console no longer shows the "Incompatible React versions" exception and the UI renders instead of a blank page.

## 2. Add vite-env.d.ts

- [x] 2.1 Create `frontend/src/vite-env.d.ts` with `/// <reference types="vite/client" />` and verify `npx tsc --noEmit` no longer reports `TS2882` for the `./index.css` import.
- [x] 2.2 Run `npm run build` and verify it completes successfully.

## 3. Install and configure eslint

- [x] 3.1 Add `eslint`, `typescript-eslint`, `eslint-plugin-react-hooks`, and `eslint-plugin-react-refresh` as devDependencies and verify `npm install` succeeds.
- [x] 3.2 Add `frontend/eslint.config.js` (flat config) covering TS + React rules consistent with `tsconfig.json`, and verify `npm run lint` runs without an "eslint: not found"-style error.
- [x] 3.3 Fix any lint violations `npm run lint` reports on the current source tree and verify the command exits with status 0.

## 4. Verify

- [x] 4.1 Run `npm run build`, `npm run lint`, and `npm run dev` (manual browser check) together and verify all three succeed with no errors.
