## 1. Fondations techniques

- [x] 1.1 Extraire l'ensemble des types et interfaces de `App.tsx` vers un nouveau fichier `frontend/src/types.ts`.
- [x] 1.2 Implémenter le service d'abstraction API typé dans `frontend/src/services/api.ts` pour encapsuler tous les appels `fetch`.
- [x] 1.3 Créer le hook personnalisé de notifications `frontend/src/hooks/useToast.ts`.

## 2. Boîtes de dialogue autonomes (Radix UI)

- [x] 2.1 Implémenter le composant `frontend/src/components/CreateFilterDialog.tsx` avec validation de formulaire, gestion d'erreurs et callback `onSuccess`.
- [x] 2.2 Implémenter le composant `frontend/src/components/MoveFolderDialog.tsx` pour encapsuler le choix de chemin, les états de déplacement et le callback `onSuccess`.

## 3. Custom Hooks de domaine d'état

- [x] 3.1 Créer le hook `frontend/src/hooks/useHealthAndStats.ts` pour gérer le statut d'API et les KPIs de statistiques avec polling.
- [x] 3.2 Créer le hook `frontend/src/hooks/usePlaylist.ts` pour le chargement paginé, le filtrage et la recherche textuelle de la playlist.
- [x] 3.3 Créer le hook `frontend/src/hooks/useFilters.ts` pour la récupération et la manipulation des filtres d'inclusion/exclusion.
- [x] 3.4 Créer le hook `frontend/src/hooks/useLogs.ts` pour récupérer l'historique des tâches de traitement avec polling.
- [x] 3.5 Créer le hook `frontend/src/hooks/useDownloads.ts` pour suivre les fichiers en cours de téléchargement avec polling.

## 4. Composants de présentation et Assemblage

- [x] 4.1 Extraire l'en-tête et les KPI cards dans des composants dédiés `frontend/src/components/FloatingHeader.tsx` et `frontend/src/components/StatsKPICards.tsx`.
- [x] 4.2 Découper les contenus d'onglets dans des sous-répertoires dédiés : `PlaylistTab`, `FiltersTab`, `LogsTab`, `DownloadsTab`.
- [x] 4.3 Nettoyer et réécrire le composant principal `frontend/src/App.tsx` pour orchestrer les hooks et les composants avec moins de 100 lignes de code.

## 5. Validation et Tests

- [x] 5.1 Lancer le linter et la vérification des types TypeScript (`npm run lint` et `npm run build` dans le sous-répertoire `frontend`) pour s'assurer de l'absence d'erreurs.
- [x] 5.2 Valider visuellement et fonctionnellement le fonctionnement du tableau de bord (onglets, filtres, modales, toasts) en s'assurant qu'aucune régression n'a été introduite.
