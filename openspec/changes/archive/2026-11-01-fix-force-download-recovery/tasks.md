# Tasks

## 1. Réactivation d'une occurrence annulée (handler)

- [x] 1.1 Dans `internal/api/force_download.go`, brancher « `DownloadInfoID` déjà présent » de `persistForceDownloadPath` pour remettre `status='pending'`, `retry_count=0`, `error_message=NULL`, `locked_at/locked_by=NULL`, en plus de `download_path` et `is_forced`, et repasser la `ProcessedLine` de `cancelled`/`failed` à `processed`, le tout dans une transaction. Vérifier avec un test de `force_download_test.go` : un `DownloadInfo` `cancelled` + ligne `cancelled` deviennent `pending`/`processed`, `retry_count` vaut 0, le POST répond 202.
- [x] 1.2 Ajouter un test : un re-force sur une occurrence `downloaded` reste refusé (409 `already_downloaded`) et ne modifie aucun enregistrement ; un re-force d'une occurrence annulée n'affecte pas les autres occurrences annulées du même film. Vérifier : `go test ./internal/api -run ForceDownloadItem`.
- [x] 1.3 Vérifier que l'enregistrement réactivé apparaît dans `GET /api/v1/downloads?status=active` (test de `handlers_frontend_test.go` ou test d'intégration du handler). Vérifier : le test passe.

## 2. Scheduler : exemption des forcés et candidat unique

- [x] 2.1 Ajouter à `scheduler.Item` (`internal/scheduler/types.go`) les champs `ForcedLine` et `ForcedDestPath`, et adapter `attachResume` (`internal/scheduler/build.go`) : si `download.IsForced`, remplacer `Candidates` par la seule ligne forcée et renseigner `ForcedDestPath` depuis `download.DownloadPath` (normalisation partagée avec `resume_helper.buildBaseDestPath`, extension retirée seulement si c'est une extension média). Vérifier avec un test de `build_test.go` : un item résumé forcé n'a qu'un candidat, et un résumé non forcé garde le comportement actuel.
- [x] 2.2 Dans `mergeIncompleteDownloads`, sauter `findDownloadedLine`/`cancelSkippedIncompleteDownload` pour les `DownloadInfo` `IsForced`, films et épisodes, en conservant le contrôle « confirmé non monitoré → ignoré sans annuler ». Vérifier avec des tests de `build_test.go` : forcé + frère `downloaded` → inclus et non annulé ; non forcé + frère `downloaded` → toujours annulé ; forcé + non monitoré → ignoré, statut inchangé.
- [x] 2.3 Garantir un seul forcé par item et par run dans `mergeIncompleteDownloads` : garder le `DownloadInfo` le plus ancien (`CreatedAt`, puis `ID`), laisser les autres `pending`. Vérifier avec un test à deux forcés pour le même film (un seul retenu, l'autre inchangé), et un test pour un épisode.
- [x] 2.4 Vérifier qu'un stream synthétisé pour un film déjà téléchargé supprime le stream tier-2 correspondant (`dedupTier2MovieStreams`) lorsqu'un forcé existe. Vérifier : test de `build_test.go`.

## 3. Exécuteur : honorer l'occurrence et la destination forcées

- [x] 3.1 Dans `cmd/download.go`, adapter `downloadItem` : si `item.ForcedDestPath` est renseigné, l'utiliser à la place de `BaseDestDir + QualityTags(...)` et ne tenter que le candidat forcé (pas de repli sur d'autres occurrences en cas d'échec). Vérifier avec un test de `cmd/download_test.go` : un film avec un forcé 1080p et un 720p mieux classé télécharge le 1080p vers le chemin enregistré.
- [x] 3.2 Ajouter un test de bout en bout du scheduler + exécuteur : film avec un frère `downloaded`, forcé `pending` → le run télécharge le forcé, le `DownloadInfo` passe `completed`, la ligne `downloaded`, le frère est intact et le fichier est écrit sous le chemin suffixé. Vérifier : `go test ./cmd ./internal/scheduler`.

## 4. Spécifications et documentation

- [x] 4.1 Valider les deltas : `openspec validate fix-force-download-recovery --strict` ne signale aucune erreur ; les trois capacités concernées (`force-download-media-occurrence`, `media-download-scheduling`, `cancel-download-occurrence`) correspondent au comportement testé.
- [x] 4.2 Mettre à jour `README.md` (ligne décrivant `POST /api/v1/items/:id/force-download`) pour préciser que le transfert est exécuté au prochain run planifié et qu'un re-force réactive une occurrence annulée. Vérifier : la phrase est présente et cohérente avec la spec.

## 5. Vérification d'intégration

- [x] 5.1 Lancer `go build ./... && go test ./...` : tout passe, aucune régression sur les tests de duplication (`internal/api/duplicate_keep_one_test.go`) ni sur `duplicate-download-detection` (le filtre `is_forced = false` de `internal/database/duplicates.go` reste cohérent avec les forcés réactivés).
- [ ] 5.2 Après déploiement en prod : re-forcer Luca (item 220645) depuis l'UI, vérifier que `download_info` 1705 repasse `pending`, puis qu'au run suivant (≤ 2 h) le fichier `Luca (2021) [1080p][MULTI]` est présent à côté du 720p VF et que le statut est `completed` (lecture seule en base).

## Workflow follow-up

- Archiver le change une fois la revue et le déploiement validés.
