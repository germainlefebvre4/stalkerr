## Context

L'entité `ProcessedLine` stocke les informations des entrées de playlist M3U et les associe à des métadonnées enrichies `Movie` ou `TVShow` de TMDB, tout en enregistrant leur progression via l'attribut `State`. 
Actuellement, l'API backend `GET /api/v1/items` ne permet de filtrer que par type de contenu (`content_type`), par état du pipeline (`state`), et par groupe (`group_title` via le paramètre de recherche global).
De plus, l'IHM n'offre pas d'input séparé pour rechercher spécifiquement par nom de média (VOD / `tvg_name`), ni de moyen de filtrer les flux enrichis ou non, ce qui rend la correction d'association difficile. Enfin, l'état `organizing` du pipeline est absent de l'IHM.

## Goals / Non-Goals

**Goals:**
- **Filtrage SQL performant** : Étendre l'API backend pour traiter les nouveaux critères de filtrage de manière optimale en base de données.
- **Découplage de la recherche** : Permettre de chercher à la fois par titre de média et par groupe/catégorie de manière indépendante.
- **Filtrage TMDB** : Identifier instantanément les médias enrichis (Oui) ou non enrichis (Non).
- **Intégration d'IHM fluide** : Présenter les nouveaux contrôles dans un panneau de filtres responsive et moderne, aligné avec la charte graphique de Stalkeer.
- **Gestion de l'état `organizing`** : Rendre l'état d'organisation sélectionnable et visuellement distinct dans l'IHM.

**Non-Goals:**
- Modification du schéma de base de données (aucune migration requise, les colonnes et index nécessaires existent déjà).
- Modification de la recherche avancée `/api/v1/items/search` (qui est globale et non paginée/filtrée de la même façon).

## Decisions

### Décision 1 : Extension de l'API `GET /api/v1/items`
Nous ajoutons de nouveaux query parameters optionnels plutôt que de créer un nouvel endpoint. C'est rétrocompatible et préserve le mécanisme de pagination existant.
- `tvg_name` : Recherche textuelle partielle (`ILIKE %...%`) sur la colonne `tvg_name`.
- `tmdb_enriched` : Filtrage selon l'existence d'une association.
  - Valeur `"yes"` : `(movie_id IS NOT NULL OR tvshow_id IS NOT NULL)`
  - Valeur `"no"` : `(movie_id IS NULL AND tvshow_id IS NULL)`
  - Autre valeur/vide : Aucun filtrage appliqué.

*Alternatives considérées* :
- Utiliser un endpoint `POST /api/v1/items/filter` : rejeté car un appel `GET` est plus sémantique et facilite la mise en cache ou l'historisation de l'URL si besoin.

### Décision 2 : Disposition de l'IHM en Grille Responsive
Afin de ne pas surcharger l'IHM sur mobile ou petits écrans, les filtres avancés seront présentés sous forme de grille CSS flexible (`display: grid` ou `display: flex` avec `flex-wrap` et `gap`).
- Un bloc supérieur contenant les boutons de type de contenu.
- Un bloc inférieur contenant 4 colonnes de largeur égale (s'empilant sur mobile) :
  1. Recherche VOD (Input texte)
  2. Recherche Groupe (Input texte)
  3. Enrichissement (Sélecteur "Tous / Oui / Non")
  4. État Pipeline (Sélecteur de statut complet)

### Décision 3 : Prise en charge complète de l'état `organizing`
L'état de pipeline `organizing` sera ajouté dans le menu déroulant des statuts. Pour son rendu visuel :
- Il sera traduit par `"En cours d'organisation"`.
- Son badge utilisera la classe CSS `.badge-progress` (fond bleu translucide) pour signaler l'activité de traitement.

## Risks / Trade-offs

### Risque : Dégradation des performances des requêtes GORM avec de nombreux filtres combinés
- **Impact** : Ralentissement du temps de réponse de la playlist.
- **Atténuation** : La table `processed_lines` possède déjà des index composites et individuels sur les colonnes clés :
  - `idx_processed_lines_m3u` sur `(tvg_name, group_title)`
  - `idx_processed_lines_content` sur `(content_type, state)`
  - Index de clés étrangères sur `movie_id` et `tvshow_id`.
  Les requêtes resteront donc extrêmement rapides même avec des dizaines de milliers d'entrées.
