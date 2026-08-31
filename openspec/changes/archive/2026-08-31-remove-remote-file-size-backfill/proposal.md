## Why

Le probe de taille de fichier distant (`BackfillRemoteFileSize`, lancé automatiquement à chaque `process` nocturne) envoie des rafales de requêtes HTTP concurrentes vers le panel IPTV de l'utilisateur pour sonder la taille de chaque fichier VOD. Le panel IPTV répond par des erreurs 429/503 en masse, et une fois déclenché, ce throttling dure au moins 12h côté fournisseur. Le run du 2026-08-27 (passage en sondage concurrent, `concurrency=10`) a aggravé ce risque : sur 387 lignes sondées, 100% ont échoué en 429/503, et le bug du cooldown de 7 jours appliqué uniformément (y compris aux erreurs transitoires) a en plus gelé ces lignes pour une semaine.

La donnée elle-même (taille du fichier distant, affichée dans le sidepanel de la playlist) n'apporte pas de valeur suffisante pour justifier ce risque : elle met en péril la stabilité du compte IPTV de l'utilisateur (et potentiellement l'accès au flux pour d'autres usages du produit comme le téléchargement) pour un affichage secondaire. Plutôt que d'ajouter un correctif supplémentaire (circuit breaker, cooldown différencié, etc.) sur un système déjà repris il y a 4 jours, la fonctionnalité est retirée du produit.

## What Changes

- **BREAKING**: Suppression complète du probe de taille de fichier distant exécuté pendant `process` (`internal/processor/remote_file_size.go` et son appel dans `processor.go`).
- **BREAKING**: Suppression des champs `remote_file_size` / `remote_file_size_checked_at` du modèle `ProcessedLine`, de l'API (`ItemResponse`), et de la configuration (`m3u.remote_file_size.*`).
- **BREAKING**: Suppression de l'affichage de la taille du fichier distant dans le sidepanel de la playlist M3U (frontend), y compris l'état "indisponible".
- Les colonnes DB `remote_file_size` / `remote_file_size_checked_at` et leur index composite restent en base (le projet utilise `gorm.AutoMigrate`, qui n'ajoute jamais que des colonnes et n'en supprime jamais) ; elles deviennent orphelines et inertes, sans migration de suppression dans ce changement.

## Capabilities

### New Capabilities
(aucune)

### Modified Capabilities
- `m3u-playlist-details-sidepanel`: retire l'exigence "Affichage de la taille du fichier distant dans le sidepanel" (et ses scénarios) ; le sidepanel n'affiche plus aucune information de taille de fichier distant, y compris pour les contenus VOD.

### Removed Capabilities
- `remote-file-size-backfill`: la capacité de sondage automatique de la taille des fichiers distants est retirée dans son intégralité.
- `remote-file-size-storage`: la capacité de persistance/retry de la taille de fichier distant est retirée dans son intégralité.

## Impact

- Code affecté :
  - `internal/processor/remote_file_size.go`, `internal/processor/remote_file_size_test.go` (suppression complète)
  - `internal/processor/processor.go` (retrait de l'appel `BackfillRemoteFileSize` et des champs `Statistics.FileSizeBackfilled` / `FileSizeBackfillErrors`)
  - `internal/models/processed_line.go` (retrait des champs `RemoteFileSize` / `RemoteFileSizeCheckedAt` et de leur index)
  - `internal/config/config.go`, `internal/config/config_test.go` (retrait de `RemoteFileSizeConfig`, des bindings et des defaults `m3u.remote_file_size.*`)
  - `internal/api/dto.go`, `internal/api/handlers.go`, `internal/api/handlers_frontend_test.go` (retrait du champ `remote_file_size` de `ItemResponse`)
  - `config.yml.example` (retrait de la section `remote_file_size`)
  - `frontend/src/types.ts` (retrait du champ `remote_file_size`)
  - `frontend/src/components/PlaylistTab.tsx` (retrait du bloc d'affichage et du helper `formatRemoteFileSize`)
  - `frontend/src/locales/en/playlist.json`, `frontend/src/locales/fr/playlist.json` (retrait des clés `remoteFileSize` / `remoteFileSizeUnavailable`)
- Pas d'impact sur les autres pipelines (TMDB enrichment, Radarr/Sonarr sync, téléchargement) qui ne dépendent pas de `remote_file_size`.
- Aucune migration DB de suppression de colonne n'est incluse (cohérent avec l'usage actuel de `gorm.AutoMigrate` dans le projet, qui ne gère pas les suppressions de colonnes).
