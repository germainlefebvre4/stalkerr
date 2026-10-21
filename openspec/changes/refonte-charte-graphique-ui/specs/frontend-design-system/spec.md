## Purpose

Defines the visual language of the dashboard ("mission control"): design tokens for the dark and light themes, bundled typography, flat surface rules, LED-style status indicators and iconography, so that every screen shares one coherent, self-contained look.

## ADDED Requirements

### Requirement: Mission Control Design Tokens
The frontend SHALL define its colors, surfaces, borders, radii and typography as design tokens, with one value set per theme (dark and light). The dark theme SHALL use a near-black, blue-tinted background with three surface levels separated by 1px rules, and a single flat cyan accent color that also denotes an in-progress state.

#### Scenario: Dark theme palette
- **WHEN** the dark theme is active
- **THEN** the page background, the two raised surface levels, the rule color, the three text levels and the accent SHALL resolve to the dark token values, and the accent SHALL be a single solid color

#### Scenario: Light theme is a sober secondary theme
- **WHEN** the light theme is active
- **THEN** the same token names SHALL resolve to a neutral, low-saturation light palette keeping the same single accent role and the same flat, borders-only surface rules as the dark theme

#### Scenario: Components consume tokens
- **WHEN** a component needs a color, radius or font
- **THEN** it SHALL obtain it from a design token, and no component SHALL hard-code a color literal that bypasses the active theme

### Requirement: Flat Surfaces
Surfaces SHALL be separated by 1px borders rather than shadows, gradients or translucency. Corner radii SHALL NOT exceed 6px, except for circular elements (LEDs, avatars). Overlay surfaces (drawers, dialogs, dropdowns, toasts) SHALL follow the same rule.

#### Scenario: No shadow, gradient or glass on surfaces
- **WHEN** any card, panel, header area, drawer or dialog is rendered in either theme
- **THEN** it SHALL have no box-shadow, no gradient background and no backdrop blur, and SHALL be delimited by a 1px border

#### Scenario: Interactive hover does not lift
- **WHEN** the pointer hovers a card or row
- **THEN** its position, size and shadow SHALL NOT change; only a border or background tone MAY change

### Requirement: Bundled Typography
The frontend SHALL render its UI text in IBM Plex Sans and its technical data in IBM Plex Mono, with both font families served from the application bundle and no request to a third-party font host. Technical data comprises file sizes, transfer rates, file paths, hashes, URLs, durations and timestamps.

#### Scenario: Fonts load without network access to third parties
- **WHEN** the dashboard is loaded on a host with no internet access
- **THEN** the UI text and technical data SHALL render in the bundled fonts, and no request SHALL be made to an external font service

#### Scenario: Technical data uses the monospace face
- **WHEN** a download's size, rate, path and timestamps are displayed
- **THEN** each of these values SHALL be rendered in the monospace face, with tabular numerals for numeric values

### Requirement: LED Status Indicators
Statuses (success, in progress, pending, failed, neutral) SHALL be displayed as a small LED dot followed by a short text label, colored from the status tokens. The text label SHALL always be present so that status is never conveyed by color alone. LEDs of active states SHALL pulse.

#### Scenario: Status shown as LED plus label
- **WHEN** a download with status "failed" is rendered
- **THEN** its status SHALL appear as a red LED dot followed by a short text label

#### Scenario: Active states pulse
- **WHEN** an item is in an active state (for example downloading) and animations are not reduced
- **THEN** its LED SHALL pulse

#### Scenario: Reduced animations stop the pulse
- **WHEN** animations are reduced, by the user preference or by the operating-system `prefers-reduced-motion` setting
- **THEN** LEDs SHALL be rendered steady, without pulsing

### Requirement: Vector Iconography
The frontend SHALL use vector icons from the bundled icon library for navigation, actions and the application mark, and SHALL NOT use emoji characters as icons in navigation, headers or the favicon.

#### Scenario: Navigation uses vector icons
- **WHEN** the navigation (sidebar or mobile bottom bar) is rendered
- **THEN** each section SHALL be identified by a vector icon and a text label, with no emoji character

#### Scenario: Favicon is a vector image
- **WHEN** the page is loaded
- **THEN** the browser tab icon SHALL be an SVG image consistent with the application mark, not an emoji

### Requirement: Contrast and Focus Visibility
Text and status colors SHALL meet a contrast ratio of at least 4.5:1 against their background in both themes, and every interactive element SHALL show a visible focus ring using the accent color when focused by keyboard.

#### Scenario: Keyboard focus is visible
- **WHEN** a user tabs to a button, link or input
- **THEN** a visible accent-colored focus ring SHALL be displayed around it in both themes

#### Scenario: Secondary text remains legible
- **WHEN** secondary or muted text is rendered on a surface in either theme
- **THEN** its contrast ratio against that surface SHALL be at least 4.5:1
