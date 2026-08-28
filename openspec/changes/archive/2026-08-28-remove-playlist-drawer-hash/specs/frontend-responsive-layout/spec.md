## MODIFIED Requirements

### Requirement: Sidepanel Mobile Ergonomics

The Playlist details sidepanel SHALL continue to render as a right-anchored drawer on all viewport sizes. Below the mobile breakpoint, the drawer SHALL reduce its internal padding and typography sizing for the narrower width, and its close control SHALL have a tappable area of at least `44px` by `44px`. The drawer's own content SHALL NOT require horizontal scrolling to read at any viewport width: any unbreakable technical value (e.g. a hash or identifier) SHALL either wrap or be confined to its own explicitly scrollable sub-container, never forcing the drawer itself to overflow horizontally.

#### Scenario: Sidepanel close control is easily tappable on mobile
- **WHEN** the details sidepanel is open on a viewport narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the close control with a tappable area of at least `44px` by `44px`.

#### Scenario: Sidepanel content never introduces its own horizontal scroll on mobile
- **WHEN** the details sidepanel is open on a viewport narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the panel's content within the viewport width, with no horizontal scrollbar appearing inside the panel regardless of the length of any technical field it displays.
