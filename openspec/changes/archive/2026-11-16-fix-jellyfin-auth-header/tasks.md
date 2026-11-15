# Tasks

## 1. Client Jellyfin

- [x] 1.1 Dans `newRequest` (`internal/external/jellyfin/jellyfin.go`), remplacer `X-Emby-Token` par `Authorization: MediaBrowser Token="<clé>"` et corriger le commentaire de `SystemStatus`. Vérifier avec `go build ./...` et `grep -rn "X-Emby-Token" internal --include=*.go` qui ne doit plus rien renvoyer hors tests.
- [x] 1.2 Mettre à jour `internal/external/jellyfin/jellyfin_test.go` (`NotifyPathsUpdated`, `SystemStatus`, `Sessions`) pour asserter le header `Authorization` attendu et l'absence de `X-Emby-Token`. Vérifier avec `go test ./internal/external/jellyfin/ -count=1`.
- [x] 1.3 Ajouter un test qui couvre une clé vide : la requête part avec `MediaBrowser Token=""` et un 401 est renvoyé comme `StatusError`. Vérifier avec `go test ./internal/external/jellyfin/ -run Empty -count=1`.

## 2. Consommateurs

- [x] 2.1 Mettre à jour `internal/api/integration_test_handlers_test.go` (`TestTestIntegrationConnectivity_JellyfinOK`) pour asserter le nouveau header. Vérifier avec `go test ./internal/api/ -run Jellyfin -count=1`.
- [x] 2.2 Mettre à jour `internal/downloader/jellyfin_notify_test.go` et `cmd/jellyfin_notify_test.go` pour asserter le nouveau header. Vérifier avec `go test ./internal/downloader/ -run Jellyfin -count=1`.

## 3. Vérification d'ensemble

- [x] 3.1 Lancer `go test ./... -count=1` et `go vet ./...`, qui doivent passer sans échec.
- [x] 3.2 Après déploiement, vérifier dans l'UI de configuration que « Test connection » affiche OK avec une clé valide saisie dans le champ. Vérifier aussi qu'un warning `failed to check Jellyfin active playback` n'apparaît plus dans les logs.

## Workflow follow-up

- Archiver la change une fois la revue terminée.
- Défaut secondaire à traiter dans une autre change si souhaité : clé vide envoyée au test pour un champ sensible déjà enregistré (`SettingsGroupCard.tsx`).
- Révoquer la clé API de test utilisée pendant le diagnostic.
