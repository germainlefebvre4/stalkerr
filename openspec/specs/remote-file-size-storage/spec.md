# remote-file-size-storage Specification

## Purpose

Persister la taille en octets du fichier média distant référencé par `line_url` sur une entrée de playlist, une fois qu'elle a été sondée, afin que l'API et l'interface puissent l'afficher sans nouvelle requête réseau.

## Requirements

### Requirement: Champs de taille de fichier distant sur ProcessedLine
Le modèle GORM `ProcessedLine` SHALL inclure deux colonnes nullable supplémentaires : `remote_file_size` (entier 64 bits, taille en octets) et `remote_file_size_checked_at` (timestamp), afin de distinguer une ligne jamais vérifiée d'une ligne vérifiée sans résultat exploitable.

#### Scenario: Le schéma inclut les nouvelles colonnes après migration
- **WHEN** l'application exécute son auto-migration de base de données
- **THEN** la table `processed_lines` SHALL contenir les colonnes `remote_file_size` et `remote_file_size_checked_at`, nullable, sans perte de données sur les lignes existantes

### Requirement: Retry après délai de repos pour les lignes vérifiées sans résultat
Une `ProcessedLine` dont `remote_file_size_checked_at` est renseigné mais dont `remote_file_size` est resté `NULL` (sondage échoué) SHALL redevenir éligible au sondage une fois qu'un délai de repos configurable (cooldown) s'est écoulé depuis `remote_file_size_checked_at`. Une `ProcessedLine` dont `remote_file_size` a été obtenu avec succès SHALL rester exclue du sondage : elle SHALL NOT être re-sondée, même après l'écoulement du cooldown.

#### Scenario: Ligne échouée, cooldown écoulé
- **WHEN** une `ProcessedLine` a `remote_file_size_checked_at` renseigné, `remote_file_size` à `NULL`, et que le délai de repos configuré s'est écoulé depuis `remote_file_size_checked_at`
- **THEN** cette ligne SHALL redevenir éligible à la requête d'éligibilité du backfill et SHALL pouvoir être re-sondée lors d'une exécution ultérieure

#### Scenario: Ligne échouée, cooldown non écoulé
- **WHEN** une `ProcessedLine` a `remote_file_size_checked_at` renseigné, `remote_file_size` à `NULL`, et que le délai de repos configuré ne s'est pas encore écoulé depuis `remote_file_size_checked_at`
- **THEN** cette ligne SHALL NOT être sélectionnée par la requête d'éligibilité du backfill

#### Scenario: Ligne vérifiée avec succès n'est jamais re-sondée
- **WHEN** une `ProcessedLine` a `remote_file_size` renseigné avec une valeur exploitable
- **THEN** cette ligne SHALL NOT être re-sondée, quel que soit le temps écoulé depuis `remote_file_size_checked_at`

### Requirement: L'API expose la taille du fichier distant
`ItemResponse` SHALL inclure le champ `remote_file_size` (entier 64 bits nullable). Il SHALL être `null`/omis lorsque la valeur n'a pas été renseignée, que la ligne n'ait jamais été vérifiée ou qu'elle ait été vérifiée sans résultat exploitable.

#### Scenario: Récupération d'un item dont la taille est connue
- **WHEN** un client récupère un item dont le `ProcessedLine` associé a `remote_file_size` renseigné
- **THEN** la réponse SHALL inclure `remote_file_size` avec la valeur en octets stockée

#### Scenario: Récupération d'un item non vérifié ou indisponible
- **WHEN** un client récupère un item dont le `ProcessedLine` associé a `remote_file_size` à `null` (vérifié ou non)
- **THEN** la réponse SHALL omettre ou retourner `null` pour `remote_file_size`, sans erreur
