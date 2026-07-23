## Context

L'IHM du tableau de bord (`App.tsx`) est un composant monolithique de plus de 900 lignes qui combine :
- Les types TypeScript pour les appels d'API.
- Des hooks de state nombreux (25+) entremêlés.
- Les requêtes d'API directes via `fetch`.
- Des effets secondaires de polling (`setInterval`) pour maintenir les données à jour.
- Le balisage HTML complet des quatre onglets et de deux boîtes de dialogue modales (Radix UI `Dialog`).
- Des styles inline volumineux.

Cette conception limite la testabilité, la lisibilité et l'évolutivité.

## Goals / Non-Goals

**Goals:**
- **Modularité :** Découper le monolithe en composants réutilisables, hooks de domaine et une couche de service API.
- **Séparation des responsabilités :** Isoler le state applicatif du rendu de présentation et de la communication HTTP.
- **Type Safety :** Centraliser et exporter les interfaces TypeScript pour qu'elles soient réutilisées proprement.
- **Encapsulation de l'état des Modales :** Transférer l'état transitoire des dialogues Radix (`Dialog.Content`) à l'intérieur des composants de dialogue eux-mêmes.
- **Zéro Régression :** Conserver exactement les mêmes comportements, styles (incluant l'esthétique du tableau de bord), fonctionnalités, animations et endpoints d'API.

**Non-Goals:**
- **Pas de réécriture graphique majeure :** Garder le design CSS d'origine intact (via les classes de `index.css`).
- **Pas de nouvelles fonctionnalités :** Ne pas ajouter de nouvelles fonctionnalités métiers au tableau de bord.
- **Pas de bibliothèque d'état externe :** Ne pas introduire Redux, Zustand ou React Query. La gestion de l'état s'appuiera uniquement sur des Custom Hooks fondés sur l'état natif de React (`useState`, `useEffect`, `useCallback`).

## Decisions

### 1. Centralisation des types dans `types.ts`
- **Décision :** Extraire toutes les interfaces (ex: `StatsResponse`, `PlaylistItem`, `DownloadInfo`) de `App.tsx` vers un fichier dédié `frontend/src/types.ts`.
- **Alternative :** Laisser les types dans chaque sous-composant. Rejeté car les types sont partagés entre les services, les hooks et les composants d'IHM.

### 2. Isolation de l'API dans `services/api.ts`
- **Décision :** Créer un client API structuré utilisant des fonctions de `fetch` standardisées et typées.
- **Alternative :** Continuer à faire des `fetch` directs avec des URL codées en dur dans les hooks ou composants. Rejeté pour éviter la duplication des URL et faciliter la maintenance.

### 3. Création de Custom Hooks de domaine pour chaque onglet et état partagé
- **Décision :** Introduire des hooks légers et réutilisables dans `frontend/src/hooks/` :
  - `useHealthAndStats` : Polling de l'état de l'API (`/health`) et des KPIs (`/api/v1/stats`).
  - `usePlaylist` : Gestion de l'état de pagination, filtrage par type et recherche textuelle de la playlist.
  - `useFilters` : Gestion de la liste des filtres, de la création et de la suppression.
  - `useLogs` : Historique des tâches de traitement et polling associé.
  - `useDownloads` : File de téléchargements active et polling associé.
  - `useToast` : Notifications transitoires à l'écran (succès ou erreur).
- **Alternative :** Garder tout le state dans `App.tsx` et le transmettre par props. Rejeté car cela provoquerait du "prop drilling" excessif et ne réduirait pas la taille de `App.tsx`.

### 4. Encapsulation des dialogues Radix UI
- **Décision :** Créer des composants autonomes pour les modales :
  - `MoveFolderDialog.tsx` : Encapsule les inputs de formulaires, les états de chargement/erreurs locaux et l'appel de déplacement.
  - `CreateFilterDialog.tsx` : Encapsule les formulaires d'ajout de regex et la validation.
- **Alternative :** Laisser les modales Radix dans `App.tsx` ou dans les listes d'onglets respectifs. Rejeté car les states transitoires de formulaires polluent inutilement le composant racine.

## Risks / Trade-offs

- **[Risk]** Perte ou désynchronisation d'état lors du découpage (ex: rafraîchir la liste de téléchargements après avoir fermé la modale de déplacement).
  - **Mitigation :** Les modales recevront des callbacks comme `onSuccess` pour notifier le composant parent ou le hook correspondant de déclencher un rafraîchissement des données (`refetch`).
- **[Risk]** Erreurs de compilation TypeScript avec l'utilisation de React 19.
  - **Mitigation :** Vérifier minutieusement le typage des props, des événements et s'assurer du respect des signatures standard React 19.
- **[Risk]** Altération accidentelle du style CSS lors du retrait des styles inline.
  - **Mitigation :** Exporter précautionneusement les styles CSS requis ou s'assurer que les styles inline restants ou déplacés dans les sous-composants conservent les valeurs originales au pixel près.
