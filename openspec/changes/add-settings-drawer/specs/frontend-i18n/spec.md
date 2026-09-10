## MODIFIED Requirements

### Requirement: Language Switching and Persistence
The frontend SHALL provide a language switcher control within the Settings drawer's "Langue" section, reachable from the header's single settings icon, that lets the user choose between English and French, and SHALL persist the chosen language across sessions.

#### Scenario: User switches the active language
- **WHEN** the user opens the Settings drawer and selects "Français" from the language switcher in its "Langue" section
- **THEN** the frontend SHALL immediately re-render all visible UI text in French, update the `<html lang>` attribute to `"fr"`, and store `"fr"` under the `stalkeer_language` key in `localStorage`

#### Scenario: Returning visit with a stored preference
- **WHEN** a user who previously selected French reopens the dashboard in a new browser session
- **THEN** the frontend SHALL read `stalkeer_language` from `localStorage` and render the UI in French without requiring the user to switch again or open the Settings drawer
