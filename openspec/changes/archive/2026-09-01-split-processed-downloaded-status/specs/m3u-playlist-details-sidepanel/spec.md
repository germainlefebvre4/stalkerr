## MODIFIED Requirements

### Requirement: Playlist Item Details Sidepanel
L'interface utilisateur de la playlist M3U SHALL permettre au clic sur n'importe quelle ligne de la table d'ouvrir un panneau latéral (drawer) fluide basé sur Radix UI `Dialog` affichant l'intégralité des informations enrichies du média, l'état du pipeline, ainsi que les métadonnées brutes de provenance et d'ingestion utiles à l'écran (nom original, catégorie, numéro de ligne M3U, taille du fichier distant, contenu brut d'ingestion, et URL du flux). Le panneau SHALL ne pas afficher le `line_hash` (hash unique interne de déduplication) : ce champ reste disponible via l'API mais n'a pas d'utilité pour l'utilisateur et SHALL être omis de l'affichage.

L'état du pipeline SHALL être affiché sous la forme de deux badges textuels indépendants et complets, l'un pour le statut de traitement (`pending`/`processed`) et l'autre pour le statut de téléchargement (`non téléchargé`/`downloading`/`organizing`/`downloaded`/`failed`), plutôt qu'un unique badge combiné. Lorsqu'aucun téléchargement n'a encore été tenté pour l'item, le panneau SHALL afficher explicitement un badge de statut de téléchargement `[ non téléchargé ]` de style neutre/gris, plutôt que de masquer ce champ.

Lorsque l'item associé (`Movie` ou `TVShow`) dispose d'un `poster_path` et/ou d'un `overview`, le panneau SHALL afficher une miniature du poster (via le CDN d'images TMDB) et le synopsis. Lorsque l'item dispose d'un `tmdb_id`, `imdb_id`, et/ou `tvdb_id`, le panneau SHALL afficher des liens cliquables ouvrant respectivement la fiche TMDB, IMDB, et TheTVDB correspondante (fiche film ou série selon le `content_type` de l'item) dans un nouvel onglet. Lorsque ces champs sont absents (item pas encore enrichi ou pas encore backfillé), le panneau SHALL omettre proprement le poster, le synopsis, et/ou les liens correspondants, sans erreur ni espace vide disgracieux.

Les dates affichées dans ce panneau (date d'import, date de forçage manuel) SHALL être formatées en `DD/MM/YYYY` (jour et mois sur 2 chiffres), de la même manière que la colonne "Créé le" du tableau principal, indépendamment de la locale du navigateur.

#### Scenario: Click row in playlist table to display detailed sidepanel
- **WHEN** l'utilisateur clique sur une ligne de la playlist dans le tableau
- **THEN** le frontend SHALL afficher un panneau latéral glissant contenant les métadonnées de média TMDB (titre, année, genres, etc.), les badges de statut de traitement et de téléchargement, le numéro de ligne M3U d'origine, un bloc préformaté avec le contenu brut d'ingestion copiable, et l'URL du flux accompagnée d'un bouton de copie.

#### Scenario: Sidepanel does not display the item's unique hash
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item de la playlist
- **THEN** le frontend SHALL ne rendre aucun élément affichant `line_hash` (ni libellé, ni valeur, ni bouton de copie associé), quelle que soit la longueur de ce hash

#### Scenario: Display import date with zero-padded day and month in the sidepanel
- **WHEN** le panneau latéral affiche la date d'import (`created_at`) ou la date de forçage manuel (`override_at`) d'un item
- **THEN** le frontend SHALL afficher cette date au format `DD/MM/YYYY` avec jour et mois sur 2 chiffres.

#### Scenario: Sidepanel displays poster, synopsis, and metadata links for an enriched item
- **WHEN** l'utilisateur clique sur une ligne dont l'item associé possède `poster_path`, `overview`, `tmdb_id`, `imdb_id`, et `tvdb_id`
- **THEN** le panneau latéral SHALL afficher la miniature du poster, le texte du synopsis, et trois liens cliquables menant respectivement à la fiche TMDB, IMDB, et TheTVDB de ce film ou de cette série

#### Scenario: Sidepanel gracefully omits missing metadata
- **WHEN** l'utilisateur clique sur une ligne dont l'item associé n'a pas encore de `poster_path`, d'`overview`, ou d'identifiants externes (`imdb_id`/`tvdb_id`) renseignés
- **THEN** le panneau latéral SHALL afficher les informations disponibles sans erreur, sans afficher de poster/synopsis/liens vides ou cassés pour les champs manquants

#### Scenario: Sidepanel shows independent processing and download badges after a failed forced download
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item qui a été traité avec succès et dont une tentative de téléchargement forcé a échoué
- **THEN** le panneau SHALL afficher un badge de traitement `[ processed ]` et, séparément, un badge de téléchargement `[ failed ]`, sans que l'un masque ou remplace l'autre

#### Scenario: Sidepanel shows explicit not-downloaded badge
- **WHEN** l'utilisateur ouvre le panneau latéral d'un item traité pour lequel aucun téléchargement n'a encore été tenté
- **THEN** le panneau SHALL afficher un badge de téléchargement explicite `[ non téléchargé ]` de style neutre/gris, distinct des styles "downloaded" (vert) et "failed" (rouge)
