## REMOVED Requirements

### Requirement: Affichage de la taille du fichier distant dans le sidepanel
**Reason**: La donnée `remote_file_size` n'est plus sondée ni persistée (capacité `remote-file-size-backfill` retirée) ; son affichage n'a donc plus de source de données.
**Migration**: Le sidepanel n'affiche plus aucune information relative à la taille du fichier distant, y compris l'état "indisponible", pour aucun `content_type`.

#### Scenario: Le sidepanel affiche une taille de fichier distant connue
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item VOD
- **THEN** le panneau SHALL NOT afficher d'information de taille de fichier distant

#### Scenario: Le sidepanel affiche un état indisponible
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item VOD
- **THEN** le panneau SHALL NOT afficher d'état "indisponible" relatif à une taille de fichier distant

#### Scenario: Le sidepanel omet l'information pour une chaîne live
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item de `content_type = channels`
- **THEN** le panneau SHALL n'afficher aucun élément relatif à la taille du fichier distant (comportement inchangé, désormais vrai pour tous les `content_type`)
