## Why

Sur la page Downloads, quand un épisode ou un film a été rangé sous un nom de dossier incorrect (coquille, mauvais titre), il n'existe aucun moyen de corriger ce nom depuis l'interface. La seule action existante, "Move", déplace le dossier complet vers une autre racine sans jamais permettre de changer son nom, et elle opère sur tout le film/toute la série à la fois — inadapté quand un seul épisode doit être corrigé sans perturber les autres épisodes déjà bien rangés dans le même dossier.

## What Changes

- Ajout d'une action "Renommer" sur chaque item de la page Downloads (films et séries), scopée au téléchargement cliqué (`DownloadInfo.ID`), distincte de l'action "Move" existante qui reste scopée au film/à la série entière.
- Le nouveau nom de dossier est saisi en texte libre (pré-rempli avec le nom actuel), indépendamment du matching TMDB : seul le chemin physique et `download_path` changent, les métadonnées TMDB de l'item ne sont pas modifiées.
- Seul le fichier de l'épisode (ou du film) ciblé est déplacé vers `{racine}/{nouveau nom}/Season NN/...` (série) ou `{racine}/{nouveau nom}/...` (film) ; les autres épisodes déjà présents dans le dossier saison/série d'origine ne sont ni déplacés ni modifiés en base.
- Un champ optionnel permet aussi de préciser une racine de destination différente, pour couvrir en une seule action le cas "déplacer et renommer" un item unique — sans dupliquer la logique de "Move" en masse.
- Si le dossier (saison et/ou série) d'origine devient vide après extraction de l'épisode, il est supprimé automatiquement.
- Si le chemin de destination calculé existe déjà (collision), l'opération est bloquée et retourne une erreur explicite plutôt que d'écraser un fichier existant.

## Capabilities

### New Capabilities
(aucune — l'action s'intègre dans les capacités existantes ci-dessous)

### Modified Capabilities
- `api-media-management`: nouvelle exigence pour un endpoint de renommage scopé par téléchargement (`DownloadInfo.ID`), avec déplacement d'un seul fichier, nettoyage des dossiers vidés, et blocage sur collision — en complément de l'exigence "Move Media Parent Folder" existante qui reste inchangée.
- `downloads-display-ui`: ajout d'une action "Renommer" par item (avec son propre dialog) à côté de l'action "Move" existante sur la page Downloads.

## Impact

- Backend : nouveau handler HTTP + route (`internal/api/handlers_frontend.go`, `internal/api/api.go`), réutilisation/adaptation de la logique de résolution de chemin déjà utilisée par `moveMovieFolder`/`moveTVShowFolder` (détection du dossier `Season NN`, sanitisation du nom) et de `MoveDir`/`moveFile` pour le déplacement effectif.
- Base de données : mise à jour de `download_info.download_path` pour la seule ligne concernée (aucune autre ligne `download_info`, `movies` ou `tvshows` n'est modifiée).
- Frontend : nouveau composant de dialog (sur le modèle de `MoveFolderDialog.tsx`), nouvelle entrée dans `frontend/src/services/api.ts`, ajout du bouton d'action dans `DownloadsTab.tsx`/`App.tsx`.
