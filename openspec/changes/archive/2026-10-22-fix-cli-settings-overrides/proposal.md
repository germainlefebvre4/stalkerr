# Proposal

## Why

Les commandes CLI (`download`, `process`, `m3u`, ...) résolvent leur configuration avec `settings.Effective()` **avant** `database.Initialize()`. Or `Effective()` lit les overrides via `database.Get()`, qui vaut `nil` à ce stade : la map d'overrides est alors vide, sans log ni erreur. Toute valeur surchargée depuis l'application est donc ignorée par les jobs CLI, qui retombent sur les valeurs fichier/env.

Cas observé : les clés API Radarr/Sonarr corrigées dans l'UI sont stockées en base et utilisées par le serveur, mais le CronJob `stalkerr-download` utilise encore les clés périmées du Secret et échoue avec `[EXTERNAL_SERVICE_ERROR] failed to get missing movies: unexpected status code 401`.

## What Changes

- Les commandes CLI connectent la base **avant** de résoudre la configuration effective, afin que les overrides stockés s'appliquent (clés API, URLs, notifications, sources M3U, niveaux de log, etc.).
- Les loggers sont initialisés une première fois depuis la config bootstrap (nécessaire à `database.Initialize()`), puis ré-initialisés avec la config effective (niveaux/format overridables).
- Commandes sans base aujourd'hui (`cleanup`, `config`, `m3u download`...) : résolution en meilleur effort, avec un avertissement explicite si la base est inaccessible, puis repli sur fichier/env.
- `settings.Effective()` journalise un `Warn` explicite lorsqu'il est appelé sans connexion à la base, au lieu d'ignorer silencieusement les overrides.
- Aucun changement de comportement du serveur API ni du chart Helm.

## Capabilities

### New Capabilities
<!-- Aucune -->

### Modified Capabilities
- `app-settings`: la résolution de la valeur effective doit s'appliquer à toute commande CLI (pas seulement au serveur), et l'impossibilité de lire les overrides doit être signalée.

## Impact

- Code : `cmd/download.go`, `cmd/process.go`, `cmd/resume_downloads.go`, `cmd/enrich_tvdb.go`, `cmd/db_prune.go`, `cmd/migrate.go`, `cmd/reset.go`, `cmd/m3u.go`, `cmd/cleanup.go`, `cmd/config_cmd.go`, `internal/settings/settings.go`.
- Comportement : les jobs CLI utiliseront désormais les valeurs saisies dans l'UI à la place de celles du Secret/config. Les déploiements qui ont des overrides divergents des valeurs fichier verront le comportement des jobs changer (c'est le but).
- Pas de changement de schéma de base ni d'API.
