## 1. Backend Implementation & Testing

- [x] 1.1 Mettre à jour `internal/api/handlers.go` pour extraire les paramètres `tvg_name` et `tmdb_enriched` de la requête HTTP
- [x] 1.2 Appliquer les filtres correspondants à la requête GORM dans la méthode `listItems`
- [x] 1.3 Écrire des tests unitaires complets dans `internal/api/handlers_frontend_test.go` pour valider le filtrage par `tvg_name` et `tmdb_enriched`

## 2. Frontend API & Hook Integration

- [x] 2.1 Mettre à jour la signature de `getPlaylist` dans `frontend/src/services/api.ts` pour accepter `searchName` et `tmdbEnriched`
- [x] 2.2 Adapter la construction de l'URL dans `getPlaylist` pour passer ces nouveaux paramètres au format de requête backend
- [x] 2.3 Ajouter les états `playlistSearchName` et `playlistTMDBFilter` dans le hook `frontend/src/hooks/usePlaylist.ts`
- [x] 2.4 Mettre à jour la fonction `fetchPlaylist` et le hook `useEffect` associés pour envoyer les nouveaux états à l'API et réinitialiser à la page 1 lors de la modification des filtres

## 3. Frontend UI Implementation

- [x] 3.1 Concevoir le conteneur de filtres avancés en grille responsive CSS dans `frontend/src/components/PlaylistTab.tsx`
- [x] 3.2 Intégrer les champs de saisie pour "Nom du Média (VOD)" et "Groupe / Catégorie" (en liant les anciens et nouveaux états du hook)
- [x] 3.3 Intégrer le sélecteur dropdown pour le filtrage d'enrichissement TMDB ("Tous", "Oui", "Non")
- [x] 3.4 Ajouter l'état "organizing" dans le sélecteur dropdown des statuts du pipeline ("En cours d'organisation")
- [x] 3.5 Adapter le rendu graphique des badges dans `PlaylistTab.tsx` pour prendre en compte le statut `organizing` en utilisant le style `.badge-progress`
