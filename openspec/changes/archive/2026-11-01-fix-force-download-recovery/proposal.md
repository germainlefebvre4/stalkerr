# Proposal

## Why

En production, forcer le téléchargement d'une occurrence (ex. « Luca (FHD MULTI) », `ProcessedLine` 220645) répond `202 queued` mais aucun transfert n'a lieu et rien n'apparaît dans l'onglet Downloads. L'analyse de la base de prod montre trois défauts qui s'enchaînent et qui, ensemble, vident la fonctionnalité `force-download-media-occurrence` de son sens quand le film a déjà un fichier téléchargé (précisément son cas d'usage : l'upgrade de qualité) :

1. **L'exécuteur ignore l'occurrence forcée.** `downloadItem` (`cmd/download.go`) parcourt `item.Candidates` par ordre de qualité (720p avant 1080p) sans tenir compte de `ResumeInfo` ni de `download_path`. Un run a ainsi téléchargé la variante 720p VF à la place des 1080p forcées.
2. **Le fix anti-doublon annule les forcés.** `mergeIncompleteDownloads` (`internal/scheduler/build.go`) passe en `cancelled` tout `DownloadInfo` incomplet dont le film/épisode a déjà une autre occurrence `downloaded`. `is_forced` n'est jamais lu.
3. **Re-forcer une occurrence annulée est un no-op silencieux.** Quand `DownloadInfoID` existe déjà, `persistForceDownloadPath` ne met à jour que `download_path` et `is_forced` : le statut reste `cancelled`, la ligne reste `cancelled`, le handler répond 202, et le listing par défaut masque les `cancelled`.

## What Changes

- Un re-force sur une occurrence dont le `DownloadInfo` est `cancelled` (ou dont la ligne est `cancelled`) la **réactive** : `DownloadInfo` repasse en `pending` (`retry_count` remis à 0, erreur et verrou vidés), la `ProcessedLine` repasse en `processed`.
- `mergeIncompleteDownloads` **n'annule plus** les `DownloadInfo` `is_forced` lorsqu'un frère est déjà `downloaded`, et ne les ignore plus sur ce motif. Le contrôle « monitoré » reste appliqué.
- Pour un item portant un `DownloadInfo` forcé, le scheduler **restreint les candidats à la seule ligne forcée** et utilise `download_path` comme destination (suffixes résolution/langue/variante persistés par le handler), au lieu du classement par qualité.
- Quand plusieurs occurrences forcées existent pour le même film/épisode, **une seule est traitée par run** (la plus ancienne), les autres aux runs suivants.
- La spec `cancel-download-occurrence` autorise une exception explicite à « réversible uniquement par un reset complet » : un force-download délibéré.
- Aucun changement de l'architecture asynchrone : le transfert reste exécuté par le prochain run planifié du `download` (jusqu'à 2 h de latence acceptée).

## Capabilities

### New Capabilities
<!-- Aucune nouvelle capacité. -->

### Modified Capabilities
- `force-download-media-occurrence`: réactivation d'une occurrence annulée lors d'un re-force ; l'exécution du transfert honore exactement l'occurrence forcée (candidat unique, destination persistée) ; une seule occurrence forcée par film/épisode et par run.
- `media-download-scheduling`: la reprise des téléchargements incomplets exempte les téléchargements forcés de l'annulation « autre occurrence déjà téléchargée ».
- `cancel-download-occurrence`: la réversibilité d'une annulation admet une exception, le force-download explicite de cette même occurrence.

## Impact

- Code : `internal/api/force_download.go` (réactivation, éligibilité), `internal/scheduler/build.go` (`mergeIncompleteDownloads`, `attachResume`), `internal/scheduler/types.go` / `cmd/download.go` (`downloadItem`, sélection du candidat et de la destination), tests associés (`internal/api/force_download_test.go`, tests scheduler et `cmd/download_test.go`).
- API : aucun changement de contrat ; `POST /api/v1/items/:id/force-download` répond toujours 202, mais il devient effectif pour une occurrence annulée.
- Données : aucune migration de schéma. Les `DownloadInfo` déjà annulés en prod (ex. 1705, 1706) se récupèrent en re-forçant depuis l'UI après déploiement.
- Hors périmètre : déclenchement immédiat du transfert (option écartée), affichage des `cancelled` dans l'onglet Downloads, migration des secrets du ConfigMap.
