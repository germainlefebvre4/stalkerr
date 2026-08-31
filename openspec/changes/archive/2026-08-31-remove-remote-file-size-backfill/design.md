## Context

Voir `proposal.md` - Why pour la motivation. Le projet gère son schéma via `gorm.AutoMigrate` (`database.Initialize()`, appelé par `stalkeer migrate` et au démarrage) : il n'existe pas de mécanisme de migration explicite (pas de fichiers de migration `up`/`down`) dans ce codebase.

## Goals / Non-Goals

**Goals:**
- Retirer entièrement le sondage de taille de fichier distant du chemin d'exécution de `process`, pour éliminer le risque de throttling IPTV qu'il introduit.
- Retirer le champ correspondant de tous les points où il est exposé (API, frontend) pour ne pas laisser de référence à une donnée qui ne sera plus jamais renseignée.

**Non-Goals:**
- Ne pas ajouter de migration DB pour supprimer les colonnes `remote_file_size` / `remote_file_size_checked_at` : hors de portée, voir Décision ci-dessous.
- Ne pas remplacer la fonctionnalité par une alternative (ex: taille calculée autrement, affichage différé) : le retrait est net, sans solution de repli produit.

## Decisions

### Laisser les colonnes DB orphelines plutôt que migrer

Le projet n'a pas d'outillage de migration descendante (`gorm.AutoMigrate` ajoute uniquement des colonnes). Deux options :

1. **Retenue** : retirer les champs du struct Go `ProcessedLine` et laisser les colonnes `remote_file_size` / `remote_file_size_checked_at` (et leur index composite) en base, inertes.
2. Écarté : ajouter un script SQL de migration manuelle (`ALTER TABLE ... DROP COLUMN`) exécuté hors du chemin `AutoMigrate` habituel.

L'option 2 introduirait un mécanisme de migration ad hoc qui n'existe nulle part ailleurs dans le projet, pour un gain limité (quelques colonnes nullable inutilisées, pas de PII, pas de volume significatif). L'option 1 est cohérente avec les conventions existantes du projet et ne bloque pas le retrait fonctionnel.

### Supprimer le fichier entier plutôt que de le vider de son contenu

`internal/processor/remote_file_size.go` et son fichier de test sont supprimés dans leur intégralité plutôt que vidés progressivement, puisqu'aucune partie de cette logique (probe HEAD/GET, semaphore de concurrence, politique de cooldown) n'est réutilisée ailleurs dans le codebase.

## Risks / Trade-offs

- [Un client externe de l'API lisait encore `remote_file_size` de `ItemResponse`] → Champ optionnel (`omitempty`) dès l'origine et déjà `null`/absent pour l'immense majorité du catalogue (99%+ jamais sondé avant le retrait) ; l'impact d'une disparition complète du champ est marginal et déjà dans un état proche de la normale pour les consommateurs actuels.
- [Colonnes DB orphelines laissées en place] → Sans effet fonctionnel (non lues, non écrites) ; documenté dans `proposal.md` et le delta spec `remote-file-size-storage` pour que ce ne soit pas redécouvert par surprise plus tard.

## Migration Plan

Déploiement standard (pas de downtime, pas d'étape de migration DB à exécuter) : le prochain déploiement du binaire retire le code de sondage et les champs exposés ; les colonnes existantes restent en base sans effet. Pas de rollback spécifique requis au-delà d'un retour à la version précédente du binaire si nécessaire (les colonnes n'ayant pas été supprimées, un rollback ne perd aucune donnée).
