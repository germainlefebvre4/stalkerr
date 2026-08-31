## REMOVED Requirements

### Requirement: Champs de taille de fichier distant sur ProcessedLine
**Reason**: Retiré avec la capacité `remote-file-size-backfill` : plus aucune donnée n'est sondée, la persistance dédiée n'a plus d'objet.
**Migration**: Les champs Go `RemoteFileSize` / `RemoteFileSizeCheckedAt` sont retirés du modèle `ProcessedLine`. Les colonnes `remote_file_size` / `remote_file_size_checked_at` et leur index composite restent en base (le projet utilise `gorm.AutoMigrate`, qui ne supprime jamais de colonnes) mais deviennent orphelines et ne sont plus lues ni écrites par l'application.

#### Scenario: Le schéma inclut les nouvelles colonnes après migration
- **WHEN** l'application exécute son auto-migration de base de données
- **THEN** le système SHALL NOT lire ni écrire les colonnes `remote_file_size` et `remote_file_size_checked_at`, qu'elles existent ou non en base

### Requirement: Retry après délai de repos pour les lignes vérifiées sans résultat
**Reason**: Retiré avec l'ensemble de la capacité : sans sondage, aucune politique de retry n'a plus d'objet. Les options de configuration `m3u.remote_file_size.retry_cooldown_hours` sont supprimées.
**Migration**: Toute configuration existante de `m3u.remote_file_size.retry_cooldown_hours` devient inerte et peut être retirée.

### Requirement: L'API expose la taille du fichier distant
**Reason**: Retiré avec l'ensemble de la capacité : la donnée n'étant plus sondée, elle ne peut plus être exposée de façon significative.
**Migration**: Le champ `remote_file_size` est retiré de `ItemResponse`. Les clients de l'API qui lisaient ce champ doivent cesser d'en dépendre ; il ne sera plus jamais présent dans la réponse.

#### Scenario: Récupération d'un item
- **WHEN** un client récupère un item via l'API
- **THEN** la réponse SHALL NOT inclure de champ `remote_file_size`
