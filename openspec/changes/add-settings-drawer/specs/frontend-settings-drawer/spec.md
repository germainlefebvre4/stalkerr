## Purpose

Gives the dashboard a single, discoverable home — one header entry point and a collapsible drawer — for cross-cutting preferences (appearance, language, startup defaults) and for the relocated filters management and system diagnostics surfaces.

## ADDED Requirements

### Requirement: Settings Header Entry Point
The app header SHALL display exactly one settings entry point (icon), replacing the former separate system-status icon and language selector, positioned identically on desktop and mobile viewports: sharing the same header row as the app title on both. This entry point SHALL be visually distinct from the icon used by the "Logs" tab.

#### Scenario: Single entry point on desktop
- **WHEN** the app is viewed on a desktop-width viewport
- **THEN** the header SHALL show exactly one settings icon, and SHALL NOT show a separate system-status icon or language selector

#### Scenario: Single entry point shares the title row on mobile
- **WHEN** the app is viewed on a mobile-width viewport
- **THEN** the settings icon SHALL be visible on the same header row as the app title, without wrapping to its own row

#### Scenario: Icon is distinct from the Logs tab icon
- **WHEN** the header and the tab bar (desktop segmented tabs or mobile bottom tab bar) are both visible
- **THEN** the settings icon SHALL be visually distinguishable from the "Logs" tab's icon

### Requirement: Settings Drawer Container
Clicking the settings entry point SHALL open a right-anchored drawer containing, in order: Apparence, Langue, Filtres, Préférences, Système, À propos. "Filtres" and "Système" SHALL render as disclosure sections collapsed by default; "Apparence", "Langue", "Préférences", and "À propos" SHALL render always-expanded. The drawer SHALL provide an explicit close control and SHALL NOT change the app's currently active tab.

#### Scenario: Opening the drawer shows all sections in order
- **WHEN** the user clicks the settings entry point
- **THEN** the drawer SHALL open and display the six sections in the order Apparence, Langue, Filtres, Préférences, Système, À propos, with "Filtres" and "Système" collapsed

#### Scenario: Closing the drawer preserves the active tab
- **WHEN** the drawer is open while a given tab is active and the user closes the drawer
- **THEN** the same tab SHALL remain active, unchanged

#### Scenario: Expanding a collapsed section reveals its content in place
- **WHEN** the user expands the "Filtres" or "Système" section
- **THEN** the drawer SHALL reveal that section's content without closing the drawer or navigating away from it

### Requirement: Theme Toggle
The drawer's "Apparence" section SHALL provide a light/dark theme toggle. Selecting a theme SHALL immediately re-render the app using that theme's tokens and SHALL persist the choice across sessions. Absent a stored preference, the app SHALL use the light theme.

#### Scenario: Switching to dark theme re-renders immediately
- **WHEN** the user selects the dark option in the Apparence section
- **THEN** the frontend SHALL immediately apply the dark theme's tokens across the app and persist the choice

#### Scenario: Stored theme preference is restored on next visit
- **WHEN** a user who previously selected the dark theme reopens the dashboard in a new browser session
- **THEN** the frontend SHALL render in the dark theme without requiring the user to switch again

#### Scenario: No stored preference defaults to light
- **WHEN** a user opens the dashboard for the first time, with no theme preference stored
- **THEN** the frontend SHALL render in the light theme

### Requirement: Reduce Motion Toggle
The drawer's "Apparence" section SHALL provide a "reduce animations" toggle. When enabled, the frontend SHALL disable the toast notification's entrance animation and the card hover lift/transition effect, and SHALL persist the choice across sessions.

#### Scenario: Enabling reduce animations disables the toast entrance animation
- **WHEN** the user enables "reduce animations" and an action subsequently triggers a toast notification
- **THEN** the toast SHALL appear without its entrance animation

#### Scenario: Enabling reduce animations disables the card hover lift
- **WHEN** the user enables "reduce animations"
- **THEN** cards SHALL NOT apply their hover lift/transform/shadow transition

#### Scenario: Preference persists across sessions
- **WHEN** a user who previously enabled "reduce animations" reopens the dashboard in a new browser session
- **THEN** the frontend SHALL keep animations reduced without requiring the user to re-enable the toggle

### Requirement: Default Startup Tab Preference
The drawer's "Préférences" section SHALL display and let the user set the tab the app opens to on its next load (absent a `tab` URL parameter), using the same persisted value that normal tab navigation already updates. Setting it from "Préférences" SHALL NOT change the currently active tab in the session in which it was set.

#### Scenario: Setting the default tab does not navigate the current session
- **WHEN** the user selects "Downloads" as the default startup tab in Préférences while "Home" is the currently active tab
- **THEN** the frontend SHALL persist "downloads" as the startup tab, and "Home" SHALL remain the active tab in the current session

#### Scenario: Default tab is used on next load
- **WHEN** the app is reloaded with no `tab` URL parameter, after the user set "Downloads" as the default startup tab
- **THEN** the frontend SHALL open directly to the Downloads tab

### Requirement: Default Playlist Page Size and View Preferences
The drawer's "Préférences" section SHALL let the user set the default Playlist page size (10, 50, or 100 entries) and the default Playlist view (list or grouped), using the same persisted preferences the Playlist tab's own controls already read and write. Changing either value from "Préférences" SHALL immediately apply if the Playlist tab is currently mounted, and SHALL otherwise be used the next time the Playlist tab is opened.

#### Scenario: Changing page size applies live when Playlist is open
- **WHEN** the Playlist tab is currently active and the user changes the default page size to `50` in Préférences
- **THEN** the Playlist tab's currently displayed page size SHALL update to `50`

#### Scenario: Changing default view applies without visiting Playlist
- **WHEN** the Playlist tab is not currently mounted and the user sets the default view to "Grouped" in Préférences
- **THEN** the frontend SHALL persist "grouped" as the default view, and the Playlist tab SHALL open in the grouped view the next time it is activated

### Requirement: About Section
The drawer's "À propos" section SHALL display the application's name and a link to its source repository.

#### Scenario: About section shows app identity and repository link
- **WHEN** the user expands the drawer and views the "À propos" section
- **THEN** the section SHALL display the application name and a link that opens the source repository
