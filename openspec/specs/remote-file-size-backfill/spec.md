# remote-file-size-backfill Specification

## Purpose

Déterminer et persister automatiquement la taille des fichiers médias distants des entrées de playlist VOD pendant le pipeline `process`, sans commande dédiée ni déclenchement manuel, tout en protégeant le pipeline d'ingestion des fournisseurs IPTV peu fiables.

## Requirements

### Requirement: Backfill automatique lors de chaque exécution de `process`
Chaque exécution de `stalkeer process` SHALL, dans le cadre de ce même run, rechercher les `ProcessedLine` dont `content_type` est `movies`, `tvshows` ou `uncategorized`, dont `line_url` est non-nul, et dont `remote_file_size_checked_at` est `NULL`, puis tenter de déterminer la taille du fichier distant de chacune. Ce backfill SHALL s'exécuter automatiquement et silencieusement, sans flag CLI requis pour le déclencher.

#### Scenario: Ligne héritée backfillée pendant un run
- **WHEN** `stalkeer process` s'exécute et qu'un `ProcessedLine` éligible existe avec `remote_file_size_checked_at` à `NULL`
- **THEN** le système SHALL sonder la taille du fichier distant de cette ligne pendant le run et persister le résultat

#### Scenario: Aucune ligne à vérifier
- **WHEN** `stalkeer process` s'exécute et qu'aucun `ProcessedLine` éligible n'a `remote_file_size_checked_at` à `NULL`
- **THEN** le système SHALL se limiter à la requête de sélection, n'émettre aucune requête réseau de sondage, et n'ajouter aucun délai significatif au run

### Requirement: Sondage par HEAD avec repli par GET par plage
Pour chaque ligne éligible, le système SHALL envoyer une requête HTTP `HEAD` vers `line_url` pour lire l'en-tête `Content-Length`. Si le serveur distant ne répond pas correctement au `HEAD` ou ne renvoie pas de `Content-Length` exploitable, le système SHALL retenter avec une requête `GET` avec l'en-tête `Range: bytes=0-0`, et lire la taille totale depuis l'en-tête `Content-Range` de la réponse.

#### Scenario: Le HEAD suffit
- **WHEN** la requête `HEAD` vers `line_url` répond avec un `Content-Length` valide
- **THEN** cette valeur SHALL être persistée comme `remote_file_size`, sans requête `GET` supplémentaire

#### Scenario: Repli sur GET par plage
- **WHEN** la requête `HEAD` échoue, retourne un statut d'erreur, ou ne fournit pas de `Content-Length` exploitable
- **THEN** le système SHALL effectuer une requête `GET` avec `Range: bytes=0-0` et extraire la taille totale du fichier depuis l'en-tête `Content-Range` de la réponse

#### Scenario: Aucune des deux méthodes n'aboutit
- **WHEN** ni le `HEAD` ni le `GET` par plage ne permettent d'obtenir une taille exploitable
- **THEN** le système SHALL marquer la ligne comme vérifiée (`remote_file_size_checked_at` renseigné) avec `remote_file_size` laissé à `NULL`, sans lever d'erreur bloquante

### Requirement: Exclusion des chaînes live du sondage
Les `ProcessedLine` de `content_type = channels` SHALL être exclues du sondage de taille de fichier distant : elles ne SHALL jamais être sélectionnées par la requête de backfill, ni sondées.

#### Scenario: Une chaîne live n'est jamais sondée
- **WHEN** le backfill de `stalkeer process` s'exécute
- **THEN** aucun `ProcessedLine` de `content_type = channels` SHALL faire l'objet d'une requête `HEAD` ou `GET` de sondage de taille

### Requirement: Plafond configurable de lignes sondées par run
Le backfill SHALL limiter le nombre de `ProcessedLine` sondées au cours d'une même exécution de `process` à une valeur configurable (avec une valeur par défaut raisonnable), afin qu'un backlog important se résorbe sur plusieurs exécutions plutôt que d'allonger indéfiniment un seul run.

#### Scenario: Backlog supérieur au plafond
- **WHEN** le nombre de lignes éligibles au backfill dépasse le plafond configuré pour ce run
- **THEN** seules un nombre de lignes égal au plafond SHALL être sondées durant ce run, les lignes restantes étant traitées lors d'exécutions ultérieures

#### Scenario: Backlog inférieur au plafond
- **WHEN** le nombre de lignes éligibles au backfill est inférieur au plafond configuré
- **THEN** toutes les lignes éligibles SHALL être sondées durant ce run

### Requirement: Résilience face aux échecs individuels
Un échec ou un dépassement de délai lors du sondage d'une ligne SHALL être journalisé, marquer cette ligne comme vérifiée (`remote_file_size_checked_at` renseigné, `remote_file_size` laissé à `NULL`), et SHALL NOT interrompre le sondage des lignes suivantes ni faire échouer la commande `process` dans son ensemble.

#### Scenario: Une ligne échoue, les suivantes continuent
- **WHEN** le sondage d'une ligne du backlog échoue (erreur réseau, timeout, réponse invalide)
- **THEN** l'échec SHALL être journalisé, cette ligne SHALL être marquée comme vérifiée sans taille, et le backfill SHALL continuer avec les lignes suivantes
- **THEN** la commande `process` SHALL se terminer normalement, sans échec provoqué par cette erreur de sondage

### Requirement: Délai d'attente borné par requête de sondage
Chaque requête `HEAD` ou `GET` par plage émise pour le sondage SHALL appliquer un délai d'attente borné, afin qu'un serveur distant qui ne répond pas ne puisse pas bloquer indéfiniment le backfill ou le run `process`.

#### Scenario: Serveur distant qui ne répond pas
- **WHEN** un serveur IPTV ne répond pas à une requête de sondage dans le délai imparti
- **THEN** la requête SHALL être abandonnée après ce délai et traitée comme un échec de sondage pour cette ligne
