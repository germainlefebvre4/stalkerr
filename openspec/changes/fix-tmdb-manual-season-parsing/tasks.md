## 1. Season/episode extraction fix

- [ ] 1.1 In `frontend/src/components/ManualOverrideDialog.tsx`, widen the season/episode regex to tolerate whitespace or a dash between the season and episode markers (e.g. `S02E25`, `S02 E25`, `S02-E25`), case-insensitively.
- [ ] 1.2 Remove the hard-coded `setOverrideSeason('1')` fallback used when an episode marker is found without a matching season marker; leave both `overrideSeason` and `overrideEpisode` empty in that case.

## 2. Verification

- [ ] 2.1 Manually verify in the "Correction Manuelle TMDB" dialog: raw title `"Inspecteur Gadget S02 E25"` pre-populates Saison=2, Épisode=25.
- [ ] 2.2 Manually verify a title with only an episode marker and no season marker leaves both fields empty, and that submitting with both empty still resolves the correct season/episode via the backend fallback.
- [ ] 2.3 Manually verify the existing collapsed format (e.g. `S02E25`) still pre-populates correctly (no regression).
