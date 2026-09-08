## Context

See proposal.md - Why. The only relevant constraint: `.card` and `.tab-panel` are both applied to the same tab-content elements (`class="card tab-panel"`), and the mobile reset at `.tab-panel { box-shadow: none; }` (specificity 0,1,0) is weaker than `.card:hover`'s box-shadow (specificity 0,2,0), so it never wins regardless of source order.

## Goals / Non-Goals

**Goals:**
- Stop `.card:hover`'s box-shadow/lift from ever applying on touch-only input, including a tap-triggered stuck `:hover` state.
- Keep desktop mouse-hover behavior byte-for-byte identical.

**Non-Goals:**
- Reworking `.tab-panel`'s mobile reset or its specificity relative to `.card`.
- Touching any other `:hover` effect in `index.css` (e.g. `.filter-card:hover`, `.btn-primary:hover`) — out of scope for this fix; they don't affect the mobile tab-content shadow this change targets.

## Decisions

- **Gate `.card:hover` behind `@media (hover: hover) and (pointer: fine)` rather than restyling `.tab-panel`.** The media feature reflects actual input hardware (not viewport width), so it correctly distinguishes "real mouse" from "touch, including a stuck tap-hover" in every case, not just below a breakpoint. Alternatives considered:
  - Raising `.tab-panel`'s specificity (e.g. `.tab-panel.card`) to beat `.card:hover` under the existing mobile width media query — rejected because it only fixes the symptom on narrow viewports; a touch device with a wide viewport (e.g. a tablet) would still get stuck-hover shadows.
  - Removing `:hover` from `.card` entirely and using `.card:active` — rejected because it changes desktop affordance (hover-to-preview-lift) instead of just fixing the touch-specific bug.

## Risks / Trade-offs

- [Some touch devices with an attached mouse/trackpad (2-in-1 laptops, styluses reporting `pointer: fine`) may still get real, non-stuck hover] → Acceptable: that's a genuine fine-pointer/hover-capable input, so the lift/shadow behaving like desktop is correct, not a regression.
- [Browser support for `hover`/`pointer` media features] → No mitigation needed: supported in all browsers this project targets (Chrome, Firefox, Safari, Edge, all modern versions); unsupported browsers simply never match the block, which fails safe (no lift/shadow anywhere, never a stuck one).
