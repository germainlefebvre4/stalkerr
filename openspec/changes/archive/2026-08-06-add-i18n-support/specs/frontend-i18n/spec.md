## Purpose

Lets the dashboard render its interface in either English or French, defaulting to English, and lets error messages coming from the backend resolve to a translated, consistent message regardless of the language the backend produced them in.

## ADDED Requirements

### Requirement: Default and Available Languages
The frontend SHALL support exactly two UI languages, English and French, and SHALL default to English when no language preference has been stored yet.

#### Scenario: First visit with no stored preference
- **WHEN** a user opens the dashboard for the first time, with no `stalkeer_language` value in `localStorage`
- **THEN** the frontend SHALL render all UI text in English and set the `<html lang>` attribute to `"en"`

#### Scenario: Unsupported browser language falls back to English
- **WHEN** the user's browser reports a language other than English or French (e.g. `de`)
- **THEN** the frontend SHALL still default to English rather than that browser language

### Requirement: Language Switching and Persistence
The frontend SHALL provide a language switcher control in the dashboard header that lets the user choose between English and French, and SHALL persist the chosen language across sessions.

#### Scenario: User switches the active language
- **WHEN** the user selects "Français" from the language switcher in the header
- **THEN** the frontend SHALL immediately re-render all visible UI text in French, update the `<html lang>` attribute to `"fr"`, and store `"fr"` under the `stalkeer_language` key in `localStorage`

#### Scenario: Returning visit with a stored preference
- **WHEN** a user who previously selected French reopens the dashboard in a new browser session
- **THEN** the frontend SHALL read `stalkeer_language` from `localStorage` and render the UI in French without requiring the user to switch again

### Requirement: Translated Error Messages from Backend Error Codes
The frontend SHALL resolve user-facing error messages (toasts and inline dialog errors) from the machine-readable `error` code returned by the backend's `ErrorResponse`, rather than from its free-text `message` field, and SHALL render a translated generic message for any `error` code that has no specific translation.

#### Scenario: Known error code renders a specific translated message
- **WHEN** an API call fails and the backend responds with `{"error": "not_found", "message": "item with id 42 not found"}` while the active language is French
- **THEN** the frontend SHALL display the French translation mapped to the `not_found` code rather than the raw English `message` text

#### Scenario: Unknown error code falls back to a generic translated message
- **WHEN** an API call fails and the backend responds with an `error` code that has no entry in the translation catalog
- **THEN** the frontend SHALL display the generic translated fallback message for the active language instead of leaving the toast untranslated or empty

### Requirement: Locale-Aware Date and Number Formatting
The frontend SHALL format displayed dates and timestamps according to the currently active UI language rather than the browser's own locale.

#### Scenario: Dates follow the active UI language, not the browser locale
- **WHEN** the active UI language is French and the user's browser locale is `en-US`
- **THEN** playlist and log timestamps SHALL render using French date/time formatting conventions

#### Scenario: Switching language reformats already-visible dates
- **WHEN** the user switches the active language from English to French while a table of dated items is visible
- **THEN** the frontend SHALL re-render the visible dates using French formatting conventions without requiring a page reload
