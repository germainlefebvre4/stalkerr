## MODIFIED Requirements

### Requirement: Playlist Item Details Sidepanel
L'interface utilisateur de la playlist M3U SHALL permettre au clic sur n'importe quelle ligne de la table d'ouvrir un panneau latéral (drawer) fluide basé sur Radix UI `Dialog` affichant l'intégralité des informations enrichies du média, l'état du pipeline, ainsi que toutes les métadonnées brutes de provenance et d'ingestion.

Les dates affichées dans ce panneau (date d'import, date de forçage manuel) SHALL être formatées en `DD/MM/YYYY` (jour et mois sur 2 chiffres), de la même manière que la colonne "Créé le" du tableau principal, indépendamment de la locale du navigateur.

#### Scenario: Click row in playlist table to display detailed sidepanel
- **WHEN** l'utilisateur clique sur une ligne de la playlist dans le tableau
- **THEN** le frontend SHALL afficher un panneau latéral glissant contenant les métadonnées de média TMDB (titre, année, genres, etc.), le statut coloré, le numéro de ligne M3U d'origine, un bloc préformaté avec le contenu brut d'ingestion copiable, et l'URL du flux accompagnée d'un bouton de copie.

#### Scenario: Display import date with zero-padded day and month in the sidepanel
- **WHEN** le panneau latéral affiche la date d'import (`created_at`) ou la date de forçage manuel (`override_at`) d'un item
- **THEN** le frontend SHALL afficher cette date au format `DD/MM/YYYY` avec jour et mois sur 2 chiffres.
