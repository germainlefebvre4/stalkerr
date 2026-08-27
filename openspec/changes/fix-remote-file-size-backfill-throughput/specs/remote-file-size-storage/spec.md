## REMOVED Requirements

### Requirement: Une seule tentative de vérification par ligne
**Reason**: Le rythme d'ingestion (~703 lignes éligibles/nuit) dépasse le plafond de sondage par run ; combinée à cette règle, une ligne dont le sondage échouait une seule fois (timeout transitoire, hoquet réseau du fournisseur IPTV) restait bloquée indéfiniment avec une taille "Unavailable", même si le serveur redevenait disponible juste après.
**Migration**: Remplacée par "Retry après délai de repos pour les lignes vérifiées sans résultat" (voir ADDED Requirements). Aucune action requise côté consommateurs de l'API : le champ `remote_file_size` reste `null` en attendant un sondage réussi, comme avant.

## ADDED Requirements

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
