## Why

In the "Correction Manuelle TMDB" dialog, the season/episode pre-fill for TV shows only matches the strict `SxxExx` pattern (no separator). When the raw title uses a space or dash between season and episode (e.g. `"Inspecteur Gadget S02 E25"`), the strict match fails and a fallback path pre-fills the episode number correctly but hard-codes the season to `1`, silently proposing a wrong season before the user validates.

## What Changes

- Widen the frontend season/episode extraction regex in `ManualOverrideDialog.tsx` to tolerate whitespace/dash between the season and episode markers (e.g. `S02 E25`, `S02-E25`).
- Remove the hard-coded `season = '1'` guess used when only an episode number is found. When season cannot be confidently extracted, both season and episode fields are left empty instead of showing a wrong value.
- No backend change: when season/episode are left empty by the frontend, the existing backend fallback (classifier-based extraction) already resolves them correctly on submit.

## Capabilities

### Modified Capabilities
- `tmdb-manual-override`: Requirement "Frontend Interactive Manual Override Modal Dialog" changes point 7 — season/episode pre-fill must tolerate separators between season and episode markers, and must never pre-fill a guessed/incorrect season when the pattern doesn't confidently match.

## Impact

- Affected file: `frontend/src/components/ManualOverrideDialog.tsx` (season/episode extraction logic only).
- No API contract change, no database change.
- No changes to `internal/classifier/classifier.go` or the override endpoint — their existing fallback behavior is what makes leaving the fields empty safe.
