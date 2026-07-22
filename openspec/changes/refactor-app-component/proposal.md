## Why

L'IHM du tableau de bord (`App.tsx`) est devenue volumineuse (plus de 900 lignes) et monolithique. Elle concentre toutes les définitions de types, l'ensemble du state de l'application (plus de 25 states), toute la logique d'appels API directe avec polling, et le balisage des interfaces utilisateur et des fenêtres modales. Cette surcharge nuit à la lisibilité, complique la maintenance et rend les tests unitaires particulièrement difficiles à écrire. Ce changement vise à appliquer les meilleures pratiques de React et de Radix UI pour restructurer la base de code du frontend.

## What Changes

- **Extraction des définitions de types** vers un fichier global `types.ts` pour une meilleure réutilisation.
- **Création d'une couche d'API centralisée** (`services/api.ts`) pour isoler les requêtes HTTP directes.
- **Modularisation de l'état** à travers des **custom hooks** spécifiques aux domaines applicatifs (`usePlaylist`, `useFilters`, `useLogs`, `useDownloads`, `useToast`, `useHealthAndStats`).
- **Découpage des boîtes de dialogue de Radix UI** (`MoveFolderDialog.tsx`, `CreateFilterDialog.tsx`) afin de rendre autonome et d'encapsuler leur état local complexe (inputs de formulaire, états d'envoi, gestion des erreurs).
- **Découpage de l'IHM principale** en composants fonctionnels dédiés (`FloatingHeader.tsx`, `StatsKPICards.tsx`, composants par onglet `PlaylistTab`, `FiltersTab`, etc.).
- **Nettoyage des styles inline** pour capitaliser sur les classes et variables globales définies dans `index.css` et `variables.css`.

## Capabilities

### New Capabilities
<!-- Aucune nouvelle fonctionnalité métier n'est introduite, il s'agit d'une refactorisation technique d'IHM existante -->

### Modified Capabilities
<!-- Les exigences fonctionnelles du tableau de bord ne sont pas altérées par cette refactorisation d'architecture -->

## Impact

- `frontend/src/App.tsx` : Nettoyé pour devenir un simple conteneur d'orchestration (estimé à moins de 100 lignes).
- `frontend/src/types.ts` : Nouveau fichier regroupant les interfaces TypeScript.
- `frontend/src/services/api.ts` : Nouvelle couche d'abstraction des requêtes API.
- `frontend/src/hooks/` : Nouveau répertoire contenant les custom hooks d'état et d'effets de domaine.
- `frontend/src/components/` : Nouveau répertoire pour les composants UI modulaires et fenêtres modales découplées.
