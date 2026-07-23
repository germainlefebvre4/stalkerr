## Why

Actuellement, l'onglet actif du tableau de bord est réinitialisé à la valeur par défaut ("Playlist M3U") à chaque rafraîchissement de la page ou d'un onglet, ce qui nuit à l'expérience utilisateur et ralentit le flux de travail. De plus, la page des téléchargements présente des cartes volumineuses où les données techniques et les statuts de validation sont regroupés de manière indistincte dans un encart compact, réduisant considérablement la lisibilité et la densité d'informations.

## What Changes

- **Persistance de l'état d'onglet actif** : Intégration d'un mécanisme de sauvegarde de l'onglet actif dans le `localStorage` du navigateur pour le restaurer après rafraîchissement.
- **Refonte minimaliste et compacte de la liste de téléchargements** :
  - Restructuration des cartes de téléchargements pour en réduire la hauteur globale (marges et paddings resserrés).
  - Isolation du chemin d'accès (dossier et fichier) comme seul contenu de l'encart gris `file-info`.
  - Extraction visuelle des caractéristiques techniques (`Format`, `Résolution`, `Taille`, `Durée`) pour les afficher inline sous forme d'une ligne claire et épurée sous l'encart de chemin d'accès.
  - Transformation des indicateurs de validation en badges visuels structurés ("chips") disposés à droite ou en dessous (validation de l'année, validation du format de fichier, et indicateur de basse qualité si résolution < 720p).
- **Intégration de nouveaux indicateurs** :
  - Affichage de la durée du média si disponible dans la ligne technique.
  - Détection automatique de basse qualité (480p ou 360p) et affichage d'un badge d'avertissement jaune.

## Capabilities

### New Capabilities
<!-- Aucune nouvelle fonctionnalité métier globale, il s'agit d'améliorations de l'IHM existante -->

### Modified Capabilities
- `frontend-ihm-dashboard`: Ajouter la persistance de l'état de l'onglet actif dans le cycle de vie de l'application.
- `downloads-display-ui`: Restructurer l'IHM des téléchargements pour un affichage plus compact, l'extraction visuelle des métadonnées, et l'affichage de chips structurés de validation ainsi que l'indicateur de basse qualité et de durée.

## Impact

- **Affected folders/files**:
  - `frontend/src/App.tsx` (persistence de l'onglet)
  - `frontend/src/components/DownloadsTab.tsx` (restructuration IHM)
- **APIs and Endpoints consumed**: Aucun changement sur les APIs consommées.
