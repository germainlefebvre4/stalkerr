# Spec Delta

## MODIFIED Requirements

### Requirement: Dedicated Diagnostic Sidepanel
Clicking a row in the Erreurs table SHALL open a sidepanel component dedicated to this tab (distinct from the Downloads tab's details sidepanel, capability `downloads-details-sidepanel`), which SHALL lead with a diagnostic section listing which reason(s) apply and their relevant values (e.g. detected vs. expected year, detected extension, or the duplicate group's sibling occurrence(s) per capability `duplicate-download-detection`), followed by the file location detail, then the matched content detail (title, year, genres, season/episode when applicable), then the download's status and dates, then the source `url`. The sidepanel SHALL NOT expose the "Déplacer" or "Renommer" actions, or any other corrective action, with the sole exception of the duplicate actions described below (verify duplicate files, and keep one duplicate file).

#### Scenario: Diagnostic section renders first
- **WHEN** the user clicks a row in the Erreurs table
- **THEN** the sidepanel SHALL open with the diagnostic reason(s) and their values as its first visible section, before file, content, and status detail

#### Scenario: No corrective actions are offered
- **WHEN** the sidepanel is open for a download whose only applicable reason(s) are `missing_year`, `year_mismatch`, or `unknown_format`
- **THEN** the sidepanel SHALL NOT render "Déplacer", "Renommer", "Associer", "Forcer le téléchargement", or any other action control

## ADDED Requirements

### Requirement: Keep One Duplicate File In The Sidepanel
When the sidepanel is open for a download flagged `duplicate_download`, it SHALL offer a way to keep one file of the duplicate group and delete the others (capability `duplicate-download-detection`). On request it SHALL fetch the group's candidate files and list each distinct file with its path, whether it exists, its size, and its latest completion date, marking the file whose name matches the library's canonical name when that information is available. The file flagged as suggested by the candidates response SHALL be pre-selected, the user SHALL be able to select any other existing file instead, and a file that does not exist SHALL NOT be selectable as the file to keep. Running the action SHALL require an explicit confirmation step that names the file that will be kept and the file(s) that will be deleted from disk. Once the request completes, the sidepanel SHALL show each occurrence's outcome (kept, deleted, or cancelled because its file was already missing) and the Radarr/Sonarr refresh outcome; a failed or not-configured refresh SHALL be shown as a distinct message telling the user the library was not refreshed.

#### Scenario: Suggested file is pre-selected
- **WHEN** the user opens the keep-one picker and the candidates response flags a file as suggested
- **THEN** that file SHALL be pre-selected as the file to keep, and the user SHALL be able to change the selection to another existing file

#### Scenario: A missing file cannot be kept
- **WHEN** a candidate's file does not exist on disk
- **THEN** the picker SHALL show it as missing and SHALL NOT allow selecting it as the file to keep

#### Scenario: Confirmation is required before deleting
- **WHEN** the user chooses to run the keep-one action
- **THEN** the sidepanel SHALL show a confirmation naming the file to keep and the file(s) to be deleted, and SHALL NOT call the keep-one endpoint until the user confirms

#### Scenario: Cancelling the confirmation changes nothing
- **WHEN** the user dismisses the confirmation
- **THEN** no request that modifies data SHALL be sent

#### Scenario: Outcomes are reflected after the action runs
- **WHEN** the keep-one request completes
- **THEN** the sidepanel SHALL show, for each occurrence, whether it was kept, deleted, or cancelled because its file was already missing, and SHALL show the refresh outcome

#### Scenario: Refresh failure is surfaced
- **WHEN** the keep-one request completes but reports the Radarr/Sonarr refresh as failed or not configured
- **THEN** the sidepanel SHALL show that the deletion succeeded and that the library was not refreshed

#### Scenario: Group no longer flagged after success
- **WHEN** the keep-one request completes successfully
- **THEN** the Erreurs list SHALL no longer show the group's download as a `duplicate_download` once refreshed
