## MODIFIED Requirements

### Requirement: M3U Sources Management
The frontend SHALL provide a "Sources M3U" section within the Configuration page's "Contenu" tab, listing the effective M3U sources (origin and runtime), and allowing the user to create a new source, edit an existing source (creating or replacing its runtime override), and delete a runtime-only source or a runtime override (reverting an overridden source to its origin definition). If the name submitted for a new source already identifies an existing runtime-only source, the frontend SHALL warn the user that the existing runtime source will be replaced before submitting, using the same trigger, dialog layout, and inline replace-warning presentation as the Filtres creation dialog (see `frontend-filters-management`). In the create/edit dialog, the source's enabled/disabled control SHALL appear before its other configuration fields (name excepted). A source's displayed file path and URL SHALL wrap or break as needed to stay within their card's width, regardless of length, and SHALL NOT force the Configuration page to scroll horizontally.

#### Scenario: Viewing the effective sources list
- **WHEN** the user views the "Contenu" tab's "Sources M3U" section
- **THEN** the frontend SHALL display every effective source with its name, file path, and download settings, distinguishing origin-only sources from runtime-overridden or runtime-only ones

#### Scenario: Auth password status shown only when set
- **WHEN** the user views a source card for a source that currently has no `download.auth_password` set
- **THEN** the frontend SHALL display no auth-password-related text for that source, rather than indicating that it is unset

#### Scenario: Auth password status shown when set
- **WHEN** the user views a source card for a source that currently has `download.auth_password` set
- **THEN** the frontend SHALL indicate that a password is set, without displaying its raw value

#### Scenario: Editing a source without changing its auth password
- **WHEN** the user edits a source's other fields without typing into the auth password field
- **THEN** the frontend SHALL omit `download.auth_password` from the submitted request, so the backend keeps the current effective password rather than clearing it

#### Scenario: Editing a source to clear its auth password
- **WHEN** the user explicitly clears the auth password field before submitting
- **THEN** the frontend SHALL submit `download.auth_password` as an explicit empty value, distinct from omitting it, so the backend clears the stored password

#### Scenario: Creating a new source
- **WHEN** the user submits a new source with a name that does not exist in the origin configuration or among existing runtime sources
- **THEN** the frontend SHALL create it as a runtime-only source and add it to the displayed list

#### Scenario: Editing an origin-defined source
- **WHEN** the user edits a source that currently comes from `config.yml`
- **THEN** the frontend SHALL create a runtime override for that source's name and mark it as coming from the interface

#### Scenario: Warning before replacing an existing runtime-only source
- **WHEN** the user submits a new source whose name matches an existing runtime-only source
- **THEN** the frontend SHALL warn, using the same inline dialog warning presentation as Filtres, that the existing runtime source will be replaced, and SHALL replace it in place (not create a second one) if the user confirms

#### Scenario: Deleting an override reverts to origin
- **WHEN** the user deletes a runtime override for a source that also has an origin definition
- **THEN** the frontend SHALL display that source's origin definition again

#### Scenario: Enabled control appears first in the dialog
- **WHEN** the user opens the create or edit dialog for an M3U source
- **THEN** the enabled/disabled control SHALL appear immediately after the name field and before the file path, URL, authentication, and tuning fields

#### Scenario: A long source URL wraps instead of overflowing
- **WHEN** a source's `url` has no natural break point and is wider than its card
- **THEN** the frontend SHALL wrap or break that URL within the card, and the Configuration page SHALL NOT gain a horizontal scrollbar
