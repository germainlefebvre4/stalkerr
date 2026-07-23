## MODIFIED Requirements

### Requirement: Real-time Monitoring Dashboard
The frontend SHALL present dedicated tabs or tables to display active downloads (incorporating file-size progress bars) and execution logs, using polling to automatically keep content fresh. The frontend SHALL also persist the active tab state in the browser's `localStorage` and restore it upon page refresh.

#### Scenario: Display active download progress bar and logs status
- **WHEN** the user navigates to the "Downloads" view
- **THEN** the frontend SHALL render active progress bars indicating the percentage of completion for elements in the `downloading` state, and automatically re-fetch the data every 10 seconds.

#### Scenario: Persist and restore active tab on refresh
- **WHEN** the user selects a tab and refreshes the browser page
- **THEN** the frontend SHALL write the selected tab name to `localStorage` upon switch, and read/restore it upon mount to active the correct view.
