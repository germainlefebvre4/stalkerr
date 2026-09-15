## REMOVED Requirements

### Requirement: Settings Header Entry Point
**Reason**: Superseded by `frontend-configuration-page`'s "Settings Header Entry Point" requirement — the header icon now navigates to a dedicated Configuration page instead of opening a drawer.
**Migration**: No user action required. The icon stays in the same header position; clicking it now navigates instead of opening a drawer.

### Requirement: Settings Drawer Container
**Reason**: The settings drawer is replaced entirely by a dedicated Configuration page (see `frontend-configuration-page`'s "Configuration Page Container" requirement), reachable via the same header icon, so that a growing set of backend-settable configuration (Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads, logging, M3U sources) has room to be presented without being cramped into a slide-out panel.
**Migration**: No data migration required. Every section previously in the drawer (Apparence, Langue, Filtres, Préférences, Système, À propos) is carried over to the Configuration page, alongside the new sections this change adds.

### Requirement: Theme Toggle
**Reason**: Superseded by `frontend-configuration-page`'s "Theme Toggle" requirement, re-hosted on the Configuration page's "Apparence" section instead of the drawer's.
**Migration**: No action required; the stored theme preference and its behavior are unchanged.

### Requirement: Reduce Motion Toggle
**Reason**: Superseded by `frontend-configuration-page`'s "Reduce Motion Toggle" requirement, re-hosted on the Configuration page's "Apparence" section instead of the drawer's.
**Migration**: No action required; the stored preference and its behavior are unchanged.

### Requirement: Default Startup Tab Preference
**Reason**: Superseded by `frontend-configuration-page`'s "Default Startup Tab Preference" requirement, re-hosted on the Configuration page's "Préférences" section instead of the drawer's.
**Migration**: No action required; the stored preference and its behavior are unchanged.

### Requirement: Default Playlist Page Size and View Preferences
**Reason**: Superseded by `frontend-configuration-page`'s "Default Playlist Page Size and View Preferences" requirement, re-hosted on the Configuration page's "Préférences" section instead of the drawer's.
**Migration**: No action required; the stored preferences and their behavior are unchanged.

### Requirement: About Section
**Reason**: Superseded by `frontend-configuration-page`'s "About Section" requirement, re-hosted on the Configuration page's "À propos" section instead of the drawer's.
**Migration**: No action required; the displayed content is unchanged.
