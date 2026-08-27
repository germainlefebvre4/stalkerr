## MODIFIED Requirements

### Requirement: Plafond configurable de lignes sondées par run
Le backfill SHALL limiter le nombre de `ProcessedLine` sondées au cours d'une même exécution de `process` à une valeur configurable (avec une valeur par défaut relevée pour dépasser le rythme d'ingestion typique d'une exécution `process`), afin qu'un backlog important se résorbe sur plusieurs exécutions plutôt que d'allonger indéfiniment un seul run.

#### Scenario: Backlog supérieur au plafond
- **WHEN** le nombre de lignes éligibles au backfill dépasse le plafond configuré pour ce run
- **THEN** seules un nombre de lignes égal au plafond SHALL être sondées durant ce run, les lignes restantes étant traitées lors d'exécutions ultérieures

#### Scenario: Backlog inférieur au plafond
- **WHEN** le nombre de lignes éligibles au backfill est inférieur au plafond configuré
- **THEN** toutes les lignes éligibles SHALL être sondées durant ce run

## ADDED Requirements

### Requirement: Sondage concurrent des lignes éligibles
Le backfill SHALL sonder plusieurs `ProcessedLine` éligibles en parallèle, jusqu'à un niveau de concurrence configurable (avec une valeur par défaut raisonnable), au lieu de les traiter une par une de manière strictement séquentielle, afin d'augmenter le débit de sondage par run sans nécessiter un plafond disproportionné.

#### Scenario: Sondage de plusieurs lignes en parallèle
- **WHEN** le backfill traite un lot de lignes éligibles dont la taille dépasse 1
- **THEN** le système SHALL émettre les sondages HEAD/GET par plage de plusieurs lignes simultanément, dans la limite du niveau de concurrence configuré

#### Scenario: Le niveau de concurrence n'est jamais dépassé
- **WHEN** le nombre de lignes restant à sonder dépasse le niveau de concurrence configuré
- **THEN** le système SHALL NOT avoir plus de requêtes de sondage en vol simultanément que ce niveau de concurrence

#### Scenario: Résilience individuelle préservée sous concurrence
- **WHEN** le sondage d'une ligne échoue alors que d'autres sondages sont en cours en parallèle
- **THEN** cet échec SHALL être journalisé et SHALL NOT interrompre ni faire échouer les sondages concurrents des autres lignes
