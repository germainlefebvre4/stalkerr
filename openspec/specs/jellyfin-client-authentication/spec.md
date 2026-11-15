# jellyfin-client-authentication Specification

## Purpose

Définit comment Stalkeer s'authentifie auprès d'un serveur Jellyfin avec une clé API, afin que la vérification de connexion, la détection de lecture en cours et la notification de scan fonctionnent avec les versions récentes de Jellyfin.

## Requirements

### Requirement: Jellyfin requests authenticate with the MediaBrowser Authorization scheme
Every request Stalkeer sends to a Jellyfin server SHALL carry the configured API key in an `Authorization` header using the `MediaBrowser` scheme, in the form `Authorization: MediaBrowser Token="<api key>"`. Stalkeer SHALL NOT send the legacy `X-Emby-Token` header.

#### Scenario: Connectivity check with a valid key on a current Jellyfin
- **WHEN** the connectivity check is run against a Jellyfin 12.x server with a valid API key
- **THEN** the request carries `Authorization: MediaBrowser Token="<api key>"` and the check reports OK

#### Scenario: Active-playback check authenticates
- **WHEN** Stalkeer polls Jellyfin for connected sessions with a valid API key
- **THEN** the request carries the `MediaBrowser` Authorization header and the sessions are returned instead of an authentication failure

#### Scenario: Library notification authenticates
- **WHEN** Stalkeer notifies Jellyfin of updated library paths
- **THEN** the request carries the `MediaBrowser` Authorization header

#### Scenario: Legacy header is not sent
- **WHEN** Stalkeer sends any request to Jellyfin
- **THEN** the request has no `X-Emby-Token` header

### Requirement: Rejected or empty keys still surface as an authentication failure
When the configured API key is empty or Jellyfin rejects it, Stalkeer SHALL still send the request and SHALL report the resulting 401 as an authentication failure, using the same unauthorized classification as for the other integrations.

#### Scenario: Jellyfin rejects the key
- **WHEN** Jellyfin answers an authenticated request with HTTP 401
- **THEN** the connectivity check reports KO with the reason `unauthorized`

#### Scenario: No API key configured
- **WHEN** the API key is empty
- **THEN** the request is still sent with an empty token and a 401 from Jellyfin is reported as `unauthorized`, not as a different error
