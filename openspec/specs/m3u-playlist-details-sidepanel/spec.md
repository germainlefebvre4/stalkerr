# m3u-playlist-details-sidepanel Specification

## Purpose
TBD - created by archiving change m3u-playlist-sidepanel. Update Purpose after archive.
## Requirements
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

### Requirement: Force-download action in the occurrence sidepanel
The playlist sidepanel (drawer) SHALL offer a "Forcer le téléchargement" action for the single occurrence currently displayed, letting the user request a forced download of exactly that occurrence without affecting any of its sibling occurrences.

This action SHALL be presented together with the "Associate" action (see Requirement: Associate action in the occurrence sidepanel) in a single action bar positioned near the top of the sidepanel, directly below the panel header, rather than in a separate section further down the panel. Any eligibility hint, queued acknowledgement, or error message related to this action SHALL be displayed directly beneath this action bar.

The action SHALL be unavailable (hidden or disabled, with an explanatory reason) when the displayed occurrence is not matched to a movie or TV show, or when the occurrence's own state already reflects a download that is in progress or already completed.

After the action is triggered, the sidepanel SHALL reflect the outcome without requiring the user to wait for the underlying file transfer: a distinguishable error message when the media could not be confirmed to exist in Radarr/Sonarr (or when that check failed), and a distinct "queued"/in-progress acknowledgement when the request was accepted, after which the sidepanel remains usable and closable while the transfer continues in the background.

#### Scenario: Action available for an eligible occurrence
- **WHEN** the user opens the sidepanel for an occurrence matched to a movie or TV show, not yet downloaded and not currently downloading
- **THEN** the sidepanel SHALL display an enabled "Forcer le téléchargement" action

#### Scenario: Action unavailable for an unmatched occurrence
- **WHEN** the user opens the sidepanel for an occurrence with no associated movie or TV show
- **THEN** the sidepanel SHALL NOT offer an active "Forcer le téléchargement" action for that occurrence

#### Scenario: Action unavailable for an already-downloaded occurrence
- **WHEN** the user opens the sidepanel for an occurrence whose own state is already "downloaded", or for which a download is already in progress
- **THEN** the sidepanel SHALL NOT offer an active "Forcer le téléchargement" action for that occurrence

#### Scenario: Successful trigger shows a queued acknowledgement
- **WHEN** the user clicks "Forcer le téléchargement" and the request is accepted (media confirmed to exist, eligibility checks passed)
- **THEN** the sidepanel SHALL display a distinct queued/in-progress acknowledgement, and SHALL remain usable and closable while the download continues in the background

#### Scenario: Failed trigger shows an explicit error
- **WHEN** the user clicks "Forcer le téléchargement" and the request is refused (media not found in Radarr/Sonarr, the existence check failed, or the occurrence turned out ineligible)
- **THEN** the sidepanel SHALL display an explicit, distinguishable error message and SHALL NOT show a queued/in-progress acknowledgement

#### Scenario: Force Download and Associate share one action bar
- **WHEN** the user opens the sidepanel for any occurrence
- **THEN** the "Associate" and "Forcer le téléchargement" actions SHALL both render side by side in a single action bar near the top of the sidepanel, above the TMDB metadata, pipeline state, and provenance sections

### Requirement: Associate action in the occurrence sidepanel
The playlist sidepanel (drawer) SHALL offer an "Associate" action for the occurrence currently displayed, opening the same manual TMDB association dialog used by the Playlist table's "Associate"/"Correct" action, pre-targeted at that occurrence. This action SHALL be available regardless of whether the occurrence is already matched to a movie/TV show, already downloaded, or mid-download, since correcting or confirming a TMDB match is independent of the occurrence's download state.

#### Scenario: Associate action opens the manual override dialog for the displayed occurrence
- **WHEN** the user clicks "Associate" in the sidepanel for a given occurrence
- **THEN** the manual override dialog SHALL open targeting that occurrence, identical to the dialog opened via the Playlist table's row-level "Associate" button

#### Scenario: Associate action available for an already-matched occurrence
- **WHEN** the sidepanel displays an occurrence already matched to a movie or TV show
- **THEN** the "Associate" action SHALL still be enabled, allowing the user to correct or reassociate it

#### Scenario: Associate action available for an occurrence mid-download or already downloaded
- **WHEN** the sidepanel displays an occurrence that is downloading or already downloaded
- **THEN** the "Associate" action SHALL remain enabled, unlike the "Forcer le téléchargement" action which becomes disabled in this state

#### Scenario: Associate action available from both entry points
- **WHEN** the sidepanel is opened either from the Playlist tab or from the Radarr/Sonarr tab
- **THEN** the "Associate" action SHALL be available and functional in both contexts

