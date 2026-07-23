## Why

Actuellement, l'exploration et la gestion de la playlist M3U dans l'IHM sont limitées par un manque d'options de filtrage et de recherche. L'unique barre de recherche, intitulée de manière trompeuse "Rechercher par Groupe / VOD", ne permet de chercher que par nom de groupe (`group_title`) et non par nom de média (VOD / `tvg_name`). De plus, il n'est pas possible de filtrer les lignes selon leur statut d'enrichissement TMDB (oui/non), ni de visualiser ou filtrer sur l'état "Organizing" du pipeline. L'ajout de ces filtres avancés permettra aux utilisateurs d'identifier rapidement les contenus non enrichis pour effectuer des corrections manuelles, de suivre précisément les flux en cours d'organisation, et de rechercher directement des contenus par leur titre.

## What Changes

- **Nouvel input de recherche par Nom du Média (VOD)** (`tvg_name`) sur l'IHM de la playlist M3U, alimentant un filtre SQL `tvg_name ILIKE %...%` dans l'API backend.
- **Clarification et découplage du filtre par Groupe / Catégorie** (`group_title`) sous forme d'un champ de saisie dédié.
- **Nouveau filtre sélecteur d'enrichissement TMDB** permettant d'afficher :
  - "Tous" les items (comportement par défaut)
  - "Oui (Enrichis)" (items avec `movie_id IS NOT NULL` ou `tvshow_id IS NOT NULL`)
  - "Non (Non Enrichis)" (items sans métadonnées TMDB associées)
- **Mise à jour du sélecteur d'État Pipeline** pour intégrer l'état "En cours d'organisation" (`organizing`), et mise à jour du composant d'affichage des badges pour gérer cet état graphiquement.
- **Évolution de l'API backend `GET /api/v1/items`** pour accepter les nouveaux query parameters `tvg_name` et `tmdb_enriched` et enrichir la requête SQL GORM correspondante.
- **Couverture de tests unitaires du backend** pour garantir la validité des requêtes SQL de filtrage.

## Capabilities

### New Capabilities

*(Aucune nouvelle capacité globale, nous faisons évoluer l'existant)*

### Modified Capabilities

- `frontend-ihm-dashboard`: Évolution des spécifications de la vue "Playlist m3u" pour inclure la recherche par Nom du Média, le filtre d'enrichissement TMDB (oui/non) et le statut de pipeline manquant ("Organizing").

## Impact

- **Backend API (`internal/api/`)** : modification de `listItems` dans `handlers.go` et écriture de tests de filtrage dans `handlers_frontend_test.go`.
- **Frontend API Service (`frontend/src/services/api.ts`)** : mise à jour de la fonction `getPlaylist` pour propager les query params `tvg_name` et `tmdb_enriched`.
- **Frontend Hook (`frontend/src/hooks/usePlaylist.ts`)** : gestion des états d'IHM pour la recherche de média, recherche de groupe, filtre TMDB et filtre d'état pipeline.
- **Frontend UI (`frontend/src/components/PlaylistTab.tsx`)** : intégration des nouveaux champs d'IHM dans un bandeau de filtres modernisé et responsive, et intégration graphique de l'état `organizing`.
