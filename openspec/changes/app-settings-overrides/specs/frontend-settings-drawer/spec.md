## MODIFIED Requirements

### Requirement: Settings Drawer Container
Clicking the settings entry point SHALL open a right-anchored drawer containing, in order: Apparence, Langue, Filtres, Sources M3U, Intégrations, Notifications, Avancé, Préférences, Système, À propos. "Filtres", "Sources M3U", "Intégrations", "Notifications", "Avancé", and "Système" SHALL render as disclosure sections collapsed by default; "Apparence", "Langue", "Préférences", and "À propos" SHALL render always-expanded. The drawer SHALL provide an explicit close control and SHALL NOT change the app's currently active tab.

#### Scenario: Opening the drawer shows all sections in order
- **WHEN** the user clicks the settings entry point
- **THEN** the drawer SHALL open and display the ten sections in order Apparence, Langue, Filtres, Sources M3U, Intégrations, Notifications, Avancé, Préférences, Système, À propos, with "Filtres", "Sources M3U", "Intégrations", "Notifications", "Avancé", and "Système" collapsed

#### Scenario: Closing the drawer preserves the active tab
- **WHEN** the drawer is open while a given tab is active and the user closes the drawer
- **THEN** the same tab SHALL remain active, unchanged

#### Scenario: Expanding a collapsed section reveals its content in place
- **WHEN** the user expands the "Filtres", "Sources M3U", "Intégrations", "Notifications", "Avancé", or "Système" section
- **THEN** the drawer SHALL reveal that section's content without closing the drawer or navigating away from it
