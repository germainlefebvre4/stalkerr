## Why

La page Downloads affiche chaque téléchargement comme une carte dense (titre, chemin, badges techniques, badges de validation, genres, barre de progression, erreur) qui reste visible en permanence pour chaque ligne. Avec beaucoup de téléchargements, la liste devient longue à parcourir et noie l'information utile au premier coup d'œil (titre + statut) sous le détail technique. La page Playlist a déjà résolu ce problème avec une ligne résumée (table desktop / carte mobile) et un panneau latéral (sidepanel) ouvert au clic pour le détail complet. On applique le même pattern à Downloads pour la cohérence de l'IHM et la lisibilité de la liste.

## What Changes

- Remplacement du rendu en carte détaillée de chaque téléchargement par une ligne résumée : titre (+année), statut (badge), indicateur de progression/taille compact — desktop en table, mobile en carte, comme `PlaylistItemsTable`.
- **BREAKING** (UI) : les actions "Move" et "Renommer", auparavant accessibles directement sur la carte, ne sont plus visibles dans la ligne résumée — elles se trouvent désormais uniquement dans le sidepanel de détail.
- Ajout d'un panneau latéral (sidepanel), ouvert au clic sur une ligne, basé sur Radix UI `Dialog` (réutilisant le pattern déjà en place sur Playlist), affichant : statut & progression, informations fichier (dossier/nom/chemin), spécifications techniques (format, résolution, durée, date de complétion), badges de validation (année, format, basse qualité), message d'erreur (uniquement si `status === 'failed'`), et les actions Move/Renommer.
- Le sidepanel reste synchronisé avec le polling 5s existant (`useDownloads`) pendant qu'il est ouvert : l'item sélectionné est dérivé par id à chaque rafraîchissement plutôt que figé au moment du clic, pour qu'un téléchargement actif affiche sa progression en temps réel. Si l'item disparaît du tableau filtré (ex. changement de filtre), le panneau se ferme automatiquement.
- La barre de filtres (statut / type / problème) reste inchangée.

## Capabilities

### New Capabilities
- `downloads-details-sidepanel`: Panneau latéral de détail d'un téléchargement, ouvert au clic sur une ligne résumée, avec synchronisation live sur le polling existant et les actions Move/Renommer.

### Modified Capabilities
- `downloads-display-ui`: La liste des téléchargements passe d'un rendu en carte détaillée à une ligne résumée cliquable (table desktop / carte mobile) ; les actions Move/Renommer et la bannière d'erreur inline quittent la ligne résumée pour le sidepanel.

## Impact

- **Frontend** :
  - `frontend/src/components/DownloadsTab.tsx` (réécriture du rendu de liste, ajout de l'état de sélection et du `Dialog` sidepanel)
  - Nouveau composant de table/carte résumée pour Downloads (miroir de `PlaylistItemsTable.tsx`)
  - `frontend/src/components/DownloadsTab.test.tsx` (le test de la bannière d'erreur doit désormais ouvrir le sidepanel avant d'asserter)
  - `frontend/src/locales/{en,fr}/downloads.json` (nouvelles clés pour le sidepanel)
  - `frontend/src/index.css` (réutilisation des classes existantes `custom-table`, `mobile-list-card`, `drawer-content`/`drawer-overlay`, `clickable-row` ; ajouts mineurs si nécessaire)
- **Aucun changement backend** : tous les champs nécessaires (`content`, `file_info`, `status`, `error_message`, etc.) sont déjà exposés par `GET /api/v1/downloads`.
