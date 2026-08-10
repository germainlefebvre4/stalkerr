## Why

Le sidepanel de détails d'un item de playlist affiche déjà les métadonnées TMDB de base (titre, année, genres) mais aucun visuel (poster) ni synopsis, et surtout aucun lien vers la fiche TMDB (ni IMDB/TheTVDB). Or ces données sont déjà récupérées par le backend à chaque appel TMDB (`GetMovieDetails`/`GetTVShowDetails` + `GetExternalIDs`) au moment de l'enrichissement automatique (`process`) et de l'override manuel — elles sont simplement jetées avant persistance. Les items déjà importés avant ce changement resteraient sans ces informations sans un mécanisme de rattrapage.

## What Changes

- Étendre les modèles `Movie`/`TVShow` avec `poster_path`, `overview`, `imdb_id` (nouveaux) et exposer `tvdb_id` (déjà stocké mais jamais renvoyé par l'API).
- Persister ces champs dans les deux flux qui appellent déjà les endpoints TMDB détaillés : l'enrichissement automatique à l'ingestion (`internal/processor`) et l'override manuel (`POST /api/v1/items/:id/override`) — sans appel TMDB supplémentaire.
- Exposer ces champs dans `MovieResponse`/`TVShowResponse`.
- Ajouter un backfill automatique et silencieux, exécuté à chaque lancement de la commande `stalkeer process`, qui complète `poster_path`/`overview`/`imdb_id` (et `tvdb_id` si absent) pour les `Movie`/`TVShow` déjà associés à un `tmdb_id` mais n'ayant pas encore ces champs. Pas de flag pour le désactiver ; requête à vide négligeable quand tout est déjà renseigné.
- Mettre à jour le sidepanel de la page Playlist M3U (`PlaylistTab.tsx`) pour afficher le poster (miniature), le synopsis, et des liens vers les fiches TMDB / IMDB / TheTVDB (movie ou tvshow selon le type), avec repli propre si les champs sont absents (item pas encore backfillé).

## Capabilities

### New Capabilities
- `tmdb-metadata-storage`: persistance des champs TMDB riches (poster, synopsis, identifiants externes) sur `Movie`/`TVShow` lors de l'ingestion automatique et de l'override manuel, et exposition via l'API.
- `tmdb-metadata-backfill`: routine automatique de rattrapage, exécutée à chaque `stalkeer process`, qui complète les champs riches manquants sur les enregistrements existants.

### Modified Capabilities
- `m3u-playlist-details-sidepanel`: le panneau latéral affiche désormais le poster, le synopsis, et des liens vers TMDB/IMDB/TheTVDB pour l'item sélectionné.

## Impact

- **Backend**: `internal/models/media.go` (nouveaux champs + migration GORM AutoMigrate), `internal/processor/processor.go` (extension de `attrs` lors de l'enrichissement), `internal/api/handlers_frontend.go` (extension de `attrs` lors de l'override), `internal/api/dto.go` (nouveaux champs sur `MovieResponse`/`TVShowResponse`), `cmd/process.go` / `internal/processor` (nouvelle routine de backfill invoquée au sein de `Process()`).
- **Frontend**: `frontend/src/types.ts` (champs supplémentaires), `frontend/src/components/PlaylistTab.tsx` (affichage poster/synopsis/liens dans le sidepanel).
- **Aucun appel TMDB additionnel** : les données sont déjà récupérées par les endpoints existants (`GetMovieDetails`, `GetTVShowDetails`, `GetMovieExternalIDs`, `GetTVShowExternalIDs`) ; seule leur persistance change. Le backfill réutilise le rate limiter TMDB existant (4 req/s par défaut).
