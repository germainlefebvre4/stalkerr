## ADDED Requirements

### Requirement: Affichage de la taille du fichier distant dans le sidepanel
Pour une entrée de `content_type` `movies`, `tvshows`, ou `uncategorized` dont `remote_file_size` est renseigné, le panneau latéral SHALL afficher la taille du fichier distant sous une forme lisible (unités Mo/Go selon la magnitude). Lorsque l'entrée a été vérifiée (`remote_file_size_checked_at` renseigné) mais que `remote_file_size` reste `null`, le panneau SHALL afficher un état "indisponible" plutôt qu'un champ vide. Pour les entrées de `content_type = channels`, le panneau SHALL omettre entièrement cette information.

#### Scenario: Le sidepanel affiche une taille de fichier distant connue
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item VOD dont `remote_file_size` est renseigné
- **THEN** le panneau SHALL afficher la taille formatée en Mo ou Go

#### Scenario: Le sidepanel affiche un état indisponible
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item VOD dont `remote_file_size_checked_at` est renseigné mais dont `remote_file_size` est `null`
- **THEN** le panneau SHALL afficher un état "indisponible" sans erreur ni champ vide disgracieux

#### Scenario: Le sidepanel omet l'information pour une chaîne live
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item de `content_type = channels`
- **THEN** le panneau SHALL n'afficher aucun élément relatif à la taille du fichier distant
