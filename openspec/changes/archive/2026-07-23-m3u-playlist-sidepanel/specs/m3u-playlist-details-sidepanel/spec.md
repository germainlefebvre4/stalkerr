## ADDED Requirements

### Requirement: Track Ingestion Line Number Capture
Le parser M3U SHALL capturer et enregistrer le numéro de la ligne d'origine de chaque entrée (la ligne `#EXTINF` correspondante) lors de la lecture du fichier, et le stocker sous le nom de `line_number` dans la base de données.

#### Scenario: Ingest M3U entries and capture correct line numbers
- **WHEN** le parser analyse un fichier M3U et traite un groupe EXTINF/URL
- **THEN** le parser SHALL associer le numéro de ligne de début d'entrée (la ligne `#EXTINF`) à l'objet `ProcessedLine` persistant.

### Requirement: Playlist Item API Ingestion Details
L'API de récupération des entrées (`/api/v1/items`) SHALL retourner toutes les informations d'ingestion brute associées à chaque entrée de la playlist M3U, à savoir : `line_content`, `line_url`, `line_hash`, et le nouveau `line_number`.

#### Scenario: Fetch playlist items API list
- **WHEN** le client appelle l'API d'obtention de la playlist `/api/v1/items`
- **THEN** la réponse JSON SHALL inclure pour chaque entrée les champs `line_content`, `line_url`, `line_hash`, et `line_number`.

### Requirement: Playlist Item Details Sidepanel
L'interface utilisateur de la playlist M3U SHALL permettre au clic sur n'importe quelle ligne de la table d'ouvrir un panneau latéral (drawer) fluide basé sur Radix UI `Dialog` affichant l'intégralité des informations enrichies du média, l'état du pipeline, ainsi que toutes les métadonnées brutes de provenance et d'ingestion.

#### Scenario: Click row in playlist table to display detailed sidepanel
- **WHEN** l'utilisateur clique sur une ligne de la playlist dans le tableau
- **THEN** le frontend SHALL afficher un panneau latéral glissant contenant les métadonnées de média TMDB (titre, année, genres, etc.), le statut coloré, le numéro de ligne M3U d'origine, un bloc préformaté avec le contenu brut d'ingestion copiable, et l'URL du flux accompagnée d'un bouton de copie.
