## 1. Retrait du pipeline de sondage (backend)

- [x] 1.1 Retirer l'appel à `BackfillRemoteFileSize` et les champs `Statistics.FileSizeBackfilled` / `FileSizeBackfillErrors` dans `internal/processor/processor.go`, et vérifier que `go build ./...` passe
- [x] 1.2 Supprimer `internal/processor/remote_file_size.go` et `internal/processor/remote_file_size_test.go`, et vérifier que `go vet ./...` ne signale plus de référence orpheline
- [x] 1.3 Retirer les champs `RemoteFileSize` / `RemoteFileSizeCheckedAt` (et leur tag d'index composite) de `internal/models/processed_line.go`, et vérifier que le projet compile toujours

## 2. Retrait de la configuration

- [x] 2.1 Retirer `RemoteFileSizeConfig`, le champ `RemoteFileSize` de `M3UConfig`, les `viper.BindEnv` et `viper.SetDefault` associés (`m3u.remote_file_size.*`) dans `internal/config/config.go`
- [x] 2.2 Retirer la section `remote_file_size` de `config.yml.example`
- [x] 2.3 Mettre à jour `internal/config/config_test.go` pour retirer les assertions sur `m3u.remote_file_size.*`, et vérifier que `go test ./internal/config/...` passe

## 3. Retrait de l'exposition API

- [x] 3.1 Retirer le champ `RemoteFileSize` de `ItemResponse` dans `internal/api/dto.go` et son mapping dans `internal/api/handlers.go`
- [x] 3.2 Mettre à jour `internal/api/handlers_frontend_test.go` pour retirer les assertions sur `remote_file_size`, et vérifier que `go test ./internal/api/...` passe

## 4. Retrait de l'affichage frontend

- [x] 4.1 Retirer le champ `remote_file_size` de `frontend/src/types.ts`
- [x] 4.2 Retirer le bloc d'affichage de la taille de fichier distant et le helper `formatRemoteFileSize` dans `frontend/src/components/PlaylistTab.tsx`
- [x] 4.3 Retirer les clés `remoteFileSize` / `remoteFileSizeUnavailable` de `frontend/src/locales/en/playlist.json` et `frontend/src/locales/fr/playlist.json`
- [x] 4.4 Vérifier que `npm run build` (ou équivalent du projet) passe côté `frontend/` sans référence résiduelle à `remote_file_size`

## 5. Vérification globale

- [x] 5.1 Exécuter la suite de tests complète (`go test ./...`) et confirmer l'absence de référence résiduelle à `RemoteFileSize`/`remote_file_size` dans le code source (`grep -r` sur `internal/` et `frontend/src/`, hors `openspec/`)
- [x] 5.2 Lancer `stalkeer process` sur un jeu de données de test et confirmer dans les logs qu'aucune requête HTTP de sondage de taille de fichier n'est émise et que le run se termine sans les étapes `Remote File Size Backfill`
