# Stalkerr — Résumé de l'application

> Dépôt : https://github.com/germainlefebvre4/stalkerr

## Description

Stalkerr est un outil (CLI + API REST + dashboard web) écrit en Go qui parse des playlists **M3U** (listes de chaînes/films/séries au format IPTV), en extrait les films et séries, les enrichit avec des métadonnées **TMDB**, puis identifie ceux qui sont **absents de Radarr / Sonarr** pour déclencher leur téléchargement via des liens directs. Les résultats (playlist traitée, statuts de téléchargement, logs) sont stockés en base **PostgreSQL** et exposés via une API REST et un dashboard web React.

En résumé, Stalkerr comble le fossé entre une source IPTV (playlist M3U) et une médiathèque gérée par Radarr/Sonarr : il détecte ce qui manque dans la bibliothèque et automatise sa récupération.

## Architecture

Application Go modulaire organisée en CLI (Cobra) + API REST (Gin) + frontend React, avec PostgreSQL comme base de données.

```
stalkerr/
├── cmd/            # Points d'entrée CLI (Cobra) : process, dryrun, server, migrate,
│                   #   radarr, sonarr, resume-downloads, m3u-download, cleanup, etc.
├── internal/
│   ├── api/         # Handlers REST (Gin)
│   ├── config/      # Configuration (Viper : config.yml + variables d'env STALKEER_*)
│   ├── database/    # Connexion PostgreSQL, migrations, maintenance (GORM)
│   ├── parser/      # Parsing des fichiers M3U
│   ├── fileparser/  # Parsing de noms de fichiers/titres
│   ├── classifier/  # Classification (film / série / chaîne / non catégorisé)
│   ├── filter/       # Filtrage des entrées selon des patterns configurables
│   ├── matcher/      # Rapprochement playlist ↔ bibliothèque Radarr/Sonarr
│   ├── scheduler/    # Ordonnancement / priorisation des téléchargements
│   ├── downloader/   # Téléchargement des fichiers, gestion espace disque, chemins
│   ├── m3udownloader/# Téléchargement + archivage/rotation des playlists M3U distantes
│   ├── external/     # Clients API externes : Radarr, Sonarr, TMDB
│   ├── processor/    # Orchestration du traitement (enrichissement, backfill)
│   ├── dryrun/       # Mode d'analyse sans écriture en base
│   ├── models/       # Modèles de données (media, download, log, filter, mapping…)
│   ├── circuitbreaker/, retry/, apperrors/ # Résilience et gestion d'erreurs
│   ├── logger/       # Logging structuré (JSON/texte)
│   └── shutdown/     # Arrêt propre (graceful shutdown)
├── frontend/         # Dashboard web (React 19 + TypeScript + Radix UI + Vite + i18next)
├── charts/stalkerr/  # Chart Helm (déploiement Kubernetes)
├── docker-compose.yml, Dockerfile  # Déploiement Docker
└── docs/             # Documentation détaillée (déploiement, dev, DB, resume-downloads…)
```

**Stack technique principale** :
- **Backend** : Go 1.25, [Gin](https://gin-gonic.com/) (API HTTP), [GORM](https://gorm.io/) (ORM PostgreSQL/SQLite), [Cobra](https://github.com/spf13/cobra) (CLI), [Viper](https://github.com/spf13/viper) (config)
- **Frontend** : React 19, TypeScript, Vite, Radix UI, i18next (multilingue)
- **Base de données** : PostgreSQL
- **Déploiement** : Docker / Docker Compose, Helm chart Kubernetes, Nginx (frontend + reverse proxy)

## Fonctionnalités clés

- **Parsing M3U** : lecture des playlists M3U et extraction des entrées films/séries/chaînes.
- **Téléchargement de playlists M3U distantes** : récupération automatique d'une playlist depuis une URL, avec archivage horodaté et rotation configurable des anciennes archives.
- **Enrichissement TMDB** : ajout de métadonnées (titre, année, genres, affiches) en plusieurs langues (en-US, fr-FR, es-ES…).
- **Classification & filtrage** : catégorisation automatique du contenu (film / série / chaîne) et filtrage selon des règles configurables.
- **Détection des éléments manquants** : comparaison de la playlist enrichie avec le contenu déjà présent dans Radarr et Sonarr.
- **Téléchargement automatisé** : récupération des films/épisodes manquants via liens directs, avec gestion de la concurrence (`--parallel`), des tentatives (`--max-retries`) et de la reprise (`resume-downloads`) après interruption/crash.
- **API REST** : endpoints pour consulter les lignes traitées, films, séries et statistiques (`/api/v1/lines`, `/movies`, `/tvshows`, `/stats`, `/health`).
- **Dashboard Web (IHM)** : interface React pour explorer les playlists, suivre la progression des téléchargements, consulter les logs de traitement et exécuter des déplacements de fichiers au niveau dossier.
- **Mode dry-run** : analyse d'une playlist sans modification de la base, utile pour prévisualiser un traitement.
- **Résilience** : circuit breaker, retries, gestion d'erreurs typées, arrêt propre (graceful shutdown).
- **Observabilité** : logs structurés (JSON/texte) avec niveaux indépendants pour l'application et la base de données.

## Fonctionnement

1. **Récupération de la playlist** : `stalkeer m3u-download` télécharge la playlist M3U depuis une URL configurée (ou fournie via `--url`), la valide, l'enregistre atomiquement et en conserve une archive horodatée (avec rotation automatique). Plusieurs sources M3U peuvent être configurées (`m3u.sources`) pour agréger plusieurs abonnements IPTV en un seul run ; l'échec d'une source n'empêche pas les autres d'être tentées (voir [M3U-DOWNLOAD.md](M3U-DOWNLOAD.md#multiple-sources)).
2. **Traitement (`process`)** : la playlist de chaque source configurée est parsée ligne par ligne, chaque entrée est classifiée (film/série/chaîne), filtrée selon les règles configurées, puis enrichie via l'API TMDB (sauf `--skip-tmdb`). Les entrées sont stockées en base PostgreSQL par lots (`--batch-size`) et taguées avec le nom de leur source d'origine (`source_name`), avec détection des doublons scopée par source (sauf `--force` pour reforcer le retraitement).
3. **Analyse rapide (`dryrun`)** : permet de rejouer l'étape de traitement sans écrire en base, pour prévisualiser les résultats (répartition films/séries, taux de correspondance TMDB…).
4. **Rapprochement avec Radarr/Sonarr** : les commandes `radarr` et `sonarr` comparent les entrées de la playlist enrichie à la bibliothèque existante dans Radarr/Sonarr et identifient les films/épisodes manquants (matcher + scheduler).
5. **Téléchargement** : les éléments manquants sont téléchargés via liens directs, avec parallélisme configurable ; les téléchargements incomplets ou en échec (crash, coupure réseau) peuvent être repris via `resume-downloads`, avec filtrage par service (`radarr`/`sonarr`) et nettoyage des verrous obsolètes.
6. **Exposition & supervision** : le serveur API (`stalkeer server`) expose les données traitées et les statistiques via REST ; le dashboard web (React) consomme cette API pour offrir une vue temps réel des playlists, téléchargements et logs. En développement, le frontend (Vite, port 5173) proxifie les appels `/api/v1/*` vers l'API Go (port 8080) pour éviter les problèmes CORS.
7. **Maintenance** : les commandes `migrate`, `db_prune`, `cleanup` et `reset` permettent respectivement d'appliquer les migrations de schéma, de purger les anciennes données, de nettoyer les fichiers temporaires/téléchargements incomplets, et de réinitialiser l'état de l'application.

---
*Résumé généré à partir de l'exploration du dépôt local (README.md, structure `cmd/`, `internal/`, `frontend/`) — le dépôt GitHub public [germainlefebvre4/stalkerr](https://github.com/germainlefebvre4/stalkerr) correspond à ce projet.*
