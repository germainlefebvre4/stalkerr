# Proposal

## Why

Le client Jellyfin n'envoie que le header legacy `X-Emby-Token`. Jellyfin 12.4.0 le rejette en 401 même avec une clé API valide, alors que `Authorization: MediaBrowser Token="<clé>"` est accepté (vérifié depuis le pod stalkerr sur `/System/Info` et `/Sessions`). Résultat : le bouton « Test connection » de la configuration Jellyfin affiche `KO – Invalid credentials` avec une clé valide, et toute l'intégration Jellyfin est en réalité inopérante.

## What Changes

- `newRequest` du client Jellyfin authentifie chaque requête avec `Authorization: MediaBrowser Token="<clé>"` à la place de `X-Emby-Token`, sans repli sur l'ancien header.
- Les tests existants qui assertent `X-Emby-Token` sont mis à jour pour asserter le nouveau header, et un test vérifie l'absence de l'ancien.
- Le commentaire de `SystemStatus` qui mentionne `X-Emby-Token` est corrigé.
- Aucun changement de comportement observable pour les utilisateurs de versions de Jellyfin qui acceptaient déjà les deux formats.

Hors périmètre, noté seulement : pour un champ sensible déjà enregistré et non retouché, le front envoie une clé vide au test de connexion (`SettingsGroupCard.tsx`), ce qui donne aussi un 401. C'est conforme à la spec `integration-connectivity-test` actuelle (l'endpoint n'utilise jamais la configuration enregistrée) et mérite sa propre change si on veut le traiter.

## Capabilities

### New Capabilities
- `jellyfin-client-authentication`: comment le client Jellyfin authentifie ses requêtes vers le serveur, commun à la vérification de connexion, à la détection de lecture en cours et à la notification de scan de bibliothèque.

### Modified Capabilities

## Impact

- Code : `internal/external/jellyfin/jellyfin.go` (`newRequest`, commentaire de `SystemStatus`).
- Tests : `internal/external/jellyfin/jellyfin_test.go`, `internal/api/integration_test_handlers_test.go`, `internal/downloader/jellyfin_notify_test.go`.
- Fonctionnalités rétablies : test d'intégration Jellyfin (`POST /api/v1/settings/integrations/test`), `ActivePlayback` (throttling adaptatif, qui échouait en silence en mode fail-open), `NotifyPathsUpdated`.
- Non vérifié en live : `POST /Library/Media/Updated`, car il déclencherait un rescan réel de la bibliothèque. La vérification se fait par test unitaire sur le header envoyé.
- Aucune dépendance ni API publique de stalkeer modifiée.
