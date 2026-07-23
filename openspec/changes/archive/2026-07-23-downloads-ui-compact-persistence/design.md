## Context

Le tableau de bord actuel utilise Radix UI et React 19. Pour améliorer l'expérience utilisateur, nous devons persister l'onglet sélectionné entre les rechargements de page et offrir une interface plus compacte et lisible pour la file de téléchargements, en séparant visuellement le chemin d'accès (encart gris) des métadonnées et indicateurs techniques (dehors de l'encart).

## Goals / Non-Goals

**Goals:**
- Persister l'état de l'onglet actif avec `localStorage`.
- Rendre la carte de téléchargement minimaliste en réduisant sa hauteur et en resserrant les espacements.
- Conserver le chemin physique (Dossier, Fichier) dans l'encart monospace `.file-info` en réduisant son padding.
- Extraire les valeurs techniques (`Format`, `Résolution`, `Taille`, `Durée`) pour les afficher inline hors de l'encart.
- Afficher les statuts de validation (Année, Format, Basse qualité) sous forme de badges structurés colorés (`badge`).

**Non-Goals:**
- Modifier l'API backend ou ajouter de nouveaux endpoints de données.
- Modifier le style des autres onglets de l'IHM.

## Decisions

### 1. Stockage de l'onglet dans le `localStorage`
- **Choix :** Utiliser `localStorage` synchrone dans l'initialiseur du `useState` d' `App.tsx` et un `useEffect` pour persister la valeur.
- **Alternative :** Stocker l'onglet dans l'URL (hash ou query string). C'est plus verbeux et moins propre pour une simple persistance de session utilisateur sans partage de liens nécessaire.

### 2. Restructuration de `DownloadsTab.tsx`
- **Choix :** Réduire le padding global de `.download-card` à `1rem` et utiliser des conteneurs Flexbox imbriqués pour diviser clairement le chemin physique, les données techniques inline, et les chips de validation.
- **Structure :**
  - Bloc `.file-info` resserré (padding `0.5rem 0.75rem`, police monospace).
  - Ligne d'informations techniques inline avec des puces `•` de séparation.
  - Ligne de chips de validation alignée horizontalement à droite des données techniques ou en wrap s'il n'y a pas assez de place.

### 3. Nouvel indicateur "Basse qualité"
- **Choix :** Calculer `isLowQuality` côté frontend en vérifiant si `detected_resolution` est égal à `"480p"` ou `"360p"`. Si oui, afficher un badge `badge-pending` (jaune/orange) avec le texte `⚠️ Basse qualité (<resolution>)`.

### 4. Affichage de la Durée
- **Choix :** Afficher `duration` (en minutes) s'il est présent dans l'objet `item.content` sous la forme `🕒 <durée> min` à côté des caractéristiques techniques.

## Risks / Trade-offs

- **[Risk]** Incohérences de layout sur petit écran (mobile).  
  → **Mitigation :** Utiliser des propriétés flexbox fluides (`flexWrap: 'wrap'`) et des gap raisonnables pour s'assurer que les chips s'empilent proprement sur mobile.
