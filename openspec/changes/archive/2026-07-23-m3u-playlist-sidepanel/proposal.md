## Why

Lors de la gestion de la playlist M3U, les utilisateurs ont besoin de retracer précisément l'origine d'une entrée : d'où elle provient, quelle est sa ligne correspondante dans le fichier M3U brut, son URL de streaming d'origine, etc. Actuellement, ces informations techniques essentielles (comme le numéro de ligne dans le fichier, le contenu brut `#EXTINF` et l'URL) ne sont pas stockées ou ne sont pas exposées sur l'interface, ce qui rend difficile le débogage des mauvaises correspondances de médias.

## What Changes

- **Ingestion & Base de Données** :
  - Modification du parser M3U pour capturer et associer le numéro de la ligne d'origine de chaque entrée.
  - Ajout d'une colonne `line_number` à la table `processed_lines` en base de données.
- **API Backend** :
  - Exposition des champs techniques de l'entrée (`line_content`, `line_url`, `line_hash` et `line_number`) dans le DTO de retour de l'API `/api/v1/items`.
- **Interface Utilisateur (IHM)** :
  - Ajout d'une interaction par clic sur n'importe quelle ligne de la table Playlist M3U.
  - Implémentation d'un panneau latéral (sidepanel/drawer) fluide et accessible basé sur Radix UI `Dialog` affichant en détail :
    - Les métadonnées enrichies du média (Titre, Année, Genres, Durée, ID TMDB, ID TVDB).
    - Les métadonnées du pipeline (Statut actuel, dates de création et mise à jour, informations d'override).
    - Les données brutes d'origine de la ligne M3U (nom d'origine, groupe/catégorie, numéro de ligne d'origine dans le fichier, hash unique).
    - Un bloc de code scrollable affichant l'extrait de ligne M3U brut complet (`#EXTINF` + URL) avec un bouton de copie rapide.
    - L'URL brute du flux de streaming avec un bouton de copie rapide.

## Capabilities

### New Capabilities
- `m3u-playlist-details-sidepanel`: Fournit un panneau de détails interactif pour chaque entrée de la playlist M3U, exposant les métadonnées médias enrichies et les données d'ingestion brutes (numéro de ligne d'origine, hash, bloc `#EXTINF` et URL).

### Modified Capabilities
_Aucune modification de spécification sur les fonctionnalités existantes._

## Impact

- **Database** : La table PostgreSQL `processed_lines` verra l'ajout d'une colonne `line_number` (type integer, default 0, non null).
- **Go Backend** : 
  - `internal/models/processed_line.go`
  - `internal/parser/parser.go`
  - `internal/api/dto.go`
  - `internal/api/handlers.go`
- **React Frontend** :
  - `frontend/src/types.ts`
  - `frontend/src/components/PlaylistTab.tsx`
  - `frontend/src/index.css`
