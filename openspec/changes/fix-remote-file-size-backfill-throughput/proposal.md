## Why

Le backfill de taille de fichier distant tourne chaque nuit mais ne traite que 200 lignes par run (`per_run_cap`), alors que l'ingestion M3U ajoute en moyenne ~703 nouvelles lignes éligibles par nuit. Le backlog croît donc d'environ 500 lignes par nuit au lieu de se résorber : sur les ~190 000 lignes éligibles actuelles, 99.3% n'ont jamais été sondées, et une ligne créée il y a plus de 9 semaines n'a toujours pas été atteinte. Résultat visible : la taille reste affichée "Unavailable" dans le side panel de la playlist M3U pour la quasi-totalité du catalogue, indéfiniment.

Une cause secondaire, marginale mais réelle (0.23% des lignes éligibles, 442 lignes), aggrave le problème : une ligne dont le sondage échoue une seule fois (timeout réseau, hoquet du serveur IPTV) est marquée comme vérifiée pour toujours et n'est plus jamais retentée, même si le serveur redevient disponible.

## What Changes

- Augmenter le débit du backfill pour dépasser durablement le rythme d'ingestion : sondage des lignes en parallèle (au lieu d'une boucle séquentielle un par un) avec une concurrence configurable, et relèvement de la valeur par défaut de `per_run_cap`.
- Ajouter une politique de retry avec délai de repos (cooldown) configurable pour les lignes déjà vérifiées mais sans taille exploitable (`remote_file_size IS NULL` malgré `remote_file_size_checked_at` renseigné) : ces lignes redeviennent éligibles au sondage une fois le cooldown écoulé, au lieu d'être exclues définitivement. **BREAKING** (comportement) : remplace la garantie actuelle "une seule tentative par ligne, jamais retentée" par une garantie de retry périodique borné dans le temps.
- Aucun changement d'API ni de schéma de données (les colonnes `remote_file_size` / `remote_file_size_checked_at` existantes suffisent ; le cooldown se calcule à partir de `remote_file_size_checked_at`).

## Capabilities

### New Capabilities
(aucune)

### Modified Capabilities
- `remote-file-size-backfill`: le plafond de lignes sondées par run passe d'un traitement strictement séquentiel à un sondage concurrent (parallélisé), avec une valeur par défaut de plafond relevée pour dépasser le rythme d'ingestion observé.
- `remote-file-size-storage`: remplace l'exigence "une seule tentative de vérification par ligne, jamais retentée" par une politique de retry après un délai de repos (cooldown) configurable pour les lignes vérifiées sans résultat exploitable.

## Impact

- Code affecté : `internal/processor/remote_file_size.go` (requête d'éligibilité, boucle de sondage, parallélisation), `internal/config/config.go` (nouvelle option de concurrence, nouvelle option de cooldown, révision du défaut de `per_run_cap`).
- Config : `config.yml` / `config.yml.example` — ajout de `remote_file_size.concurrency` et `remote_file_size.retry_cooldown_hours` (ou équivalent), révision du défaut `remote_file_size.per_run_cap`.
- Pas d'impact sur l'API (`ItemResponse`) ni sur le frontend (`PlaylistTab.tsx`) : le champ `remote_file_size` et le libellé "Unavailable" restent inchangés, seule la vitesse à laquelle les lignes passent de "jamais vérifié"/"échoué" à "taille connue" change.
- Tests existants à réviser : `internal/processor/remote_file_size_test.go` (notamment `TestBackfillRemoteFileSize_BothProbesFail`, qui vérifie aujourd'hui l'absence de retry).
