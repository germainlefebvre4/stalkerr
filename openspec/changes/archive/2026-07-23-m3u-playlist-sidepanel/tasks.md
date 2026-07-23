## 1. Extension de la Base de Données et Modèles Go

- [x] 1.1 Ajouter le champ `LineNumber` au struct `ProcessedLine` dans `internal/models/processed_line.go`
- [x] 1.2 Ajouter le champ `LineNumber` au struct `M3UEntry` dans `internal/parser/parser.go`
- [x] 1.3 Mettre à jour la méthode `parseExtinf` dans `internal/parser/parser.go` pour stocker le numéro de ligne dans l'entrée `M3UEntry`
- [x] 1.4 Mettre à jour `createProcessedLine` dans `internal/parser/parser.go` pour affecter le numéro de ligne à l'objet `ProcessedLine` retourné

## 2. API Backend et Tests unitaires Go

- [x] 2.1 Ajouter les champs bruts `LineContent`, `LineURL`, `LineHash` et `LineNumber` au struct `ItemResponse` dans `internal/api/dto.go`
- [x] 2.2 Mettre à jour la fonction `toItemResponse` dans `internal/api/handlers.go` pour mapper ces champs depuis l'objet `ProcessedLine`
- [x] 2.3 Mettre à jour et ajouter des tests unitaires dans `internal/parser/parser_test.go` pour vérifier le bon parsing et l'enregistrement du numéro de ligne
- [x] 2.4 Lancer la suite de tests Go afin de vérifier la compilation et s'assurer qu'il n'y a aucune régression

## 3. Interface Utilisateur (React Frontend)

- [x] 3.1 Déclarer les champs bruts `line_content`, `line_url`, `line_hash` et `line_number` sur l'interface `PlaylistItem` dans `frontend/src/types.ts`
- [x] 3.2 Ajouter les styles CSS d'animations de glissement (drawer sidepanel) et de bloc de code copiable dans `frontend/src/index.css`
- [x] 3.3 Gérer l'état local du média sélectionné et ajouter un gestionnaire de clic sur les lignes du tableau dans `frontend/src/components/PlaylistTab.tsx` (avec arrêt de la propagation du clic pour les boutons d'action existants)
- [x] 3.4 Implémenter le rendu du sidepanel interactif de détails basé sur Radix UI `Dialog` dans `frontend/src/components/PlaylistTab.tsx` en affichant l'ensemble des données enrichies et brutes (avec raccourcis de copie)
- [x] 3.5 Lancer le linter et le build de production de la partie frontend (`npm run lint` et `npm run build`) pour s'assurer qu'aucun avertissement ni erreur de typage n'est présent
