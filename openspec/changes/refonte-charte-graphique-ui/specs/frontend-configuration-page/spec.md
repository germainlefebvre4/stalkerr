## ADDED Requirements

### Requirement: Dark Theme by Default
The Configuration page's "Apparence" section SHALL provide a dark/light theme toggle. Selecting a theme SHALL immediately re-render the app with that theme's tokens and persist the choice across sessions. Absent a stored preference, the app SHALL use the dark theme, applied before the first paint.

#### Scenario: Switching to light theme re-renders immediately
- **WHEN** the user selects the light option in the Apparence section
- **THEN** the frontend SHALL immediately apply the light theme's tokens across the app and persist the choice

#### Scenario: Stored theme preference is restored on next visit
- **WHEN** a user who previously selected the light theme reopens the dashboard in a new browser session
- **THEN** the frontend SHALL render in the light theme without requiring the user to switch again

#### Scenario: No stored preference defaults to dark
- **WHEN** a user opens the dashboard for the first time, with no theme preference stored
- **THEN** the frontend SHALL render in the dark theme, with no visible switch from another theme during load

### Requirement: Reduce Animations Toggle
The Configuration page's "Apparence" section SHALL provide a "reduce animations" toggle. When enabled, the frontend SHALL disable the toast notification's entrance animation and the pulse of active-state LEDs, and SHALL persist the choice across sessions.

#### Scenario: Enabling reduce animations disables the toast entrance animation
- **WHEN** the user enables "reduce animations" and an action subsequently triggers a toast notification
- **THEN** the toast SHALL appear without its entrance animation

#### Scenario: Enabling reduce animations stops the LED pulse
- **WHEN** the user enables "reduce animations" while an item is in an active state such as downloading
- **THEN** that item's status LED SHALL be rendered steady, without pulsing

#### Scenario: Reduce animations preference persists across sessions
- **WHEN** a user who previously enabled "reduce animations" reopens the dashboard in a new browser session
- **THEN** the frontend SHALL keep animations reduced without requiring the user to re-enable the toggle

## MODIFIED Requirements

### Requirement: Settings Header Entry Point
The app SHALL display exactly one settings entry point (icon). On desktop viewports it SHALL be located in the footer of the navigation sidebar; on mobile viewports it SHALL be located on the same header row as the app title, without wrapping to its own row. This entry point SHALL be visually distinct from the icon used by the "Logs" section.

#### Scenario: Single entry point on desktop
- **WHEN** the app is viewed on a desktop-width viewport
- **THEN** the sidebar footer SHALL show exactly one settings icon, and no other settings icon SHALL be visible

#### Scenario: Single entry point shares the title row on mobile
- **WHEN** the app is viewed on a mobile-width viewport
- **THEN** the settings icon SHALL be visible on the same header row as the app title, without wrapping to its own row

#### Scenario: Icon is distinct from the Logs tab icon
- **WHEN** the settings entry point and the navigation (desktop sidebar or mobile bottom tab bar) are both visible
- **THEN** the settings icon SHALL be visually distinguishable from the "Logs" section's icon

## REMOVED Requirements

### Requirement: Theme Toggle
**Reason**: The default theme changes from light to dark with the mission control redesign; the requirement is replaced by "Dark Theme by Default".
**Migration**: Users with a stored theme preference keep it. Users without one see the dark theme; the toggle in "Apparence" switches to light.

### Requirement: Reduce Motion Toggle
**Reason**: The card hover lift no longer exists in the new design system; the toggle now governs LED pulse and toast animation. Replaced by "Reduce Animations Toggle".
**Migration**: The persisted preference and the toggle itself are unchanged; only the set of animations it disables changes.
