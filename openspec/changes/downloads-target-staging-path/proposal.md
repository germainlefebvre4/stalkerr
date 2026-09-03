## Why

Le sidepanel de détail d'un téléchargement (`downloads-details-sidepanel`) affiche l'emplacement du fichier uniquement une fois le download `completed` (via `download_path`, renseigné seulement par `updateDownloadInfoCompleted`). Pendant un téléchargement en cours (`downloading`/`retrying`) ou après un échec (`failed`, ex. `"failed to write file: unexpected EOF"`), la section Fichier retombe sur l'URL source, sans jamais montrer où le fichier était censé atterrir ni où il était en train d'être écrit — information utile pour diagnostiquer un échec ou suivre un download actif.

`opts.BaseDestPath` (dossier + nom de fichier final, sans extension) est pourtant déjà connu avant le premier octet transféré, et l'extension est généralement déductible de l'URL seule (`detectFileExtension`). Le chemin temporaire de staging (`tempDownloadDir/download.tmp`) est lui aussi connu dès la création du dossier temporaire, avant le début du transfert HTTP.

## What Changes

- Ajout de deux nouveaux champs sur `DownloadInfo` : `target_path` (emplacement final visé, calculé dès le début de l'attempt) et `staging_path` (chemin du fichier temporaire de l'attempt en cours).
- `target_path` est écrit une fois, dès le passage à `downloading`, avant le transfert HTTP (utilise l'extension devinée depuis l'URL).
- `staging_path` est écrit une fois par appel `Download()`, juste après la création du répertoire temporaire, avant `retry.Do`.
- À la complétion réussie (`updateDownloadInfoCompleted`), `target_path` et `staging_path` sont vidés (mis à `nil`) — `download_path` reste la seule source de vérité une fois le download terminé.
- Ces deux champs sont **purement informatifs** : aucun code existant (`rename`, `moveMovieFolder`/`moveTVShowFolder`, `resume_helper.buildBaseDestPath`) ne les lit ni n'en dépend — zéro changement de comportement sur ces chemins, `download_path` garde exactement sa sémantique actuelle ("fichier réellement présent sur disque").
- `GET /api/v1/downloads` expose `target_path` et `staging_path` dans `DownloadEnrichedResponse` (omis quand nil, comme `download_path`).
- Le sidepanel Downloads affiche, dans la section Fichier, l'emplacement prévu (`target_path`) et/ou le fichier temporaire en cours (`staging_path`) quand `download_path` n'est pas encore disponible, avec un libellé distinct signalant qu'il s'agit d'un emplacement prévu/temporaire (pas encore un fichier terminé) et que `staging_path` peut ne plus exister sur disque après un échec (le répertoire temporaire est nettoyé au retour de `Download()`, succès ou échec).

## Capabilities

### Modified Capabilities
- `downloads-enrichment-api`: `GET /api/v1/downloads` expose deux nouveaux champs optionnels `target_path` et `staging_path` sur `DownloadEnrichedResponse`.
- `downloads-details-sidepanel`: la section Fichier du sidepanel affiche l'emplacement prévu et/ou le chemin temporaire pour un download en cours ou en échec, en plus du chemin final déjà affiché pour un download terminé.

## Impact

- **Backend** :
  - `internal/models/download.go` (nouveaux champs `TargetPath`, `StagingPath` sur `DownloadInfo`, migrés via `AutoMigrate`)
  - `internal/downloader/downloader.go` (calcul et persistance de `target_path` au passage à `downloading`, de `staging_path` à la création du répertoire temporaire, purge des deux à la complétion)
  - `internal/downloader/state_manager.go` (ou fonctions dédiées pour écrire ces champs sans dupliquer la logique de `UpdateState`)
  - `internal/api/types.go` (`DownloadEnrichedResponse`: `TargetPath`, `StagingPath`)
  - `internal/api/handlers_frontend.go` (`enrichDownloadInfo`: copie des deux champs)
- **Frontend** :
  - `frontend/src/types.ts` (`DownloadEnriched`: `target_path?`, `staging_path?`)
  - `frontend/src/components/DownloadsTab.tsx` (section Fichier: rendu conditionnel de `target_path`/`staging_path` avec libellés distincts quand `download_path` est absent)
  - `frontend/src/locales/{en,fr}/downloads.json` (nouvelles clés de libellé)
- **Aucun changement** sur `rename`, `moveMovieFolder`/`moveTVShowFolder`, `resume_helper.buildBaseDestPath` : ces chemins continuent de ne lire que `download_path`.
