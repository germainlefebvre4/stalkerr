## REMOVED Requirements

### Requirement: Backfill automatique lors de chaque exécution de `process`
**Reason**: Le sondage automatique de taille de fichier distant génère des rafales de requêtes HTTP concurrentes vers le panel IPTV de l'utilisateur, déclenchant un throttling (429/503) qui dure au moins 12h côté fournisseur une fois activé. La valeur de cette donnée (affichage secondaire dans le sidepanel) ne justifie pas ce risque pour la stabilité de l'accès IPTV.
**Migration**: Aucune. `stalkeer process` ne tente plus de déterminer la taille des fichiers distants ; aucune action requise de l'utilisateur.

#### Scenario: Ligne héritée backfillée pendant un run
- **WHEN** `stalkeer process` s'exécute
- **THEN** le système SHALL NOT sonder la taille de fichier distant d'aucun `ProcessedLine`

### Requirement: Sondage par HEAD avec repli par GET par plage
**Reason**: Retiré avec l'ensemble de la capacité `remote-file-size-backfill`.
**Migration**: Aucune.

### Requirement: Exclusion des chaînes live du sondage
**Reason**: Retiré avec l'ensemble de la capacité `remote-file-size-backfill` : sans sondage, l'exclusion n'a plus d'objet.
**Migration**: Aucune.

### Requirement: Plafond configurable de lignes sondées par run
**Reason**: Retiré avec l'ensemble de la capacité `remote-file-size-backfill`. Les options de configuration `m3u.remote_file_size.per_run_cap` sont supprimées.
**Migration**: Toute configuration existante de `m3u.remote_file_size.per_run_cap` (fichier de config ou variable d'environnement) devient inerte et peut être retirée.

### Requirement: Sondage concurrent des lignes éligibles
**Reason**: Retiré avec l'ensemble de la capacité `remote-file-size-backfill`. C'est précisément ce sondage concurrent qui provoquait le throttling du fournisseur IPTV. Les options de configuration `m3u.remote_file_size.concurrency` sont supprimées.
**Migration**: Toute configuration existante de `m3u.remote_file_size.concurrency` devient inerte et peut être retirée.

### Requirement: Résilience face aux échecs individuels
**Reason**: Retiré avec l'ensemble de la capacité `remote-file-size-backfill`.
**Migration**: Aucune.

### Requirement: Délai d'attente borné par requête de sondage
**Reason**: Retiré avec l'ensemble de la capacité `remote-file-size-backfill`. Les options de configuration `m3u.remote_file_size.timeout_seconds` sont supprimées.
**Migration**: Toute configuration existante de `m3u.remote_file_size.timeout_seconds` devient inerte et peut être retirée.
