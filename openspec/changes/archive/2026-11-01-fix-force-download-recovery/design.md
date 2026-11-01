## Context

`POST /api/v1/items/:id/force-download` valide l'éligibilité, vérifie Radarr/Sonarr en direct, puis persiste un `DownloadInfo` `pending` + `is_forced=true` avec `download_path` (chemin de base suffixé résolution/langue/variante) et répond 202. Le transfert est exécuté par le prochain run du CronJob `download` (toutes les 2 h, Europe/Paris), via `mergeIncompleteDownloads` puis `downloadItem` (`cmd/download.go`).

L'analyse de la prod (item 220645, film Luca) a montré trois défauts en chaîne (chronologie en UTC) :

```
13:20 force 220645 -> DownloadInfo 1705 pending/forced
13:44 force 16052  -> DownloadInfo 1706 pending/forced
17:05 run: downloadItem classe par qualité -> télécharge 220644 (720p VF)   [défaut 1]
21:00 run: sibling "downloaded" -> 1705 et 1706 passés "cancelled"          [défaut 2]
22:06 re-force 220645 -> 202, mais status reste "cancelled"                 [défaut 3]
```

Contraintes de l'existant :
- `Downloader.Download` réutilise le `DownloadInfo` lié à la ligne (`getOrCreateDownloadInfo`) et son `download_path` n'est pas lu : la destination vient du `BaseDestPath` fourni par l'appelant.
- `resume_helper.buildBaseDestPath` (commande `resume-downloads`) honore déjà `download_path` ; seul le chemin du scheduler ne le fait pas.
- `dedupTier2MovieStreams` supprime un stream tier-2 quand un stream tier-1 existe pour le même film, y compris un stream synthétisé par `mergeIncompleteDownloads` : un forcé n'entre donc pas en concurrence avec un upgrade automatique.
- La spec `cancel-download-occurrence` interdit toute réactivation ciblée sauf reset complet ; le force explicite doit être déclaré comme exception.

## Goals / Non-Goals

**Goals:**
- Un force-download sur une occurrence annulée devient effectif au prochain run planifié.
- Un forcé est téléchargé tel quel (occurrence, destination suffixée) même si un frère est déjà `downloaded`.
- Aucune régression du fix anti-doublon pour les téléchargements non forcés.

**Non-Goals:**
- Déclencher le transfert immédiatement (latence de 2 h acceptée, architecture asynchrone inchangée).
- Afficher les `cancelled` dans l'onglet Downloads ou changer le filtre du listing.
- Modifier le contrat HTTP du handler (toujours 202 / mêmes codes d'erreur).
- Migrer les secrets en clair du ConfigMap.

## Decisions

### D1. Réactivation dans le handler, pas dans le scheduler
`persistForceDownloadPath`, dans la branche « `DownloadInfoID` déjà présent », remet en plus `status='pending'`, `retry_count=0`, `error_message=NULL`, `locked_at/locked_by=NULL`, et met la `ProcessedLine` à `processed` quand son état est `cancelled` (ou `failed`). Les deux mises à jour sont dans une même transaction, comme `cancelSkippedIncompleteDownload`.

*Alternative écartée* : laisser le scheduler réactiver un forcé annulé. Elle masque l'intention (l'utilisateur a re-forcé, le scheduler devrait deviner) et ne règle pas le 202 trompeur.

L'éligibilité reste inchangée (refus pour `downloaded`/`downloading`) : `cancelled` n'est pas refusé, il devient réactivable. Garder `StateDownloaded` comme seul motif de refus évite de toucher aux codes d'erreur existants.

### D2. Exemption `is_forced` dans `mergeIncompleteDownloads`
Pour un `DownloadInfo` incomplet avec `IsForced=true`, on saute le bloc « `findDownloadedLine` ≠ nil → `cancelSkippedIncompleteDownload` ». Le contrôle « monitoré confirmé non monitoré → ignoré, non annulé » est conservé (cohérent avec la spec existante et avec la vérification faite à l'acceptation du force).

*Alternative écartée* : supprimer l'annulation pour tous. Elle réintroduirait les re-téléchargements de doublons corrigés par `e37fc5f`.

### D3. Le forcé impose son candidat et sa destination à l'`Item`
`Item` gagne deux champs : `ForcedLine *models.ProcessedLine` et `ForcedDestPath string`. `attachResume`, quand `download.IsForced`, ne préfixe plus la ligne aux candidats (elle y est déjà souvent, d'où le bug) mais **remplace** `Candidates` par `[]ProcessedLine{*line}` et renseigne `ForcedDestPath` depuis `download.DownloadPath`. `downloadItem` utilise `ForcedDestPath` si non vide à la place de `BaseDestDir + QualityTags(...)`, et n'enchaîne pas sur d'autres candidats (un seul candidat, donc l'échec rend `itemFailed`).

La normalisation du chemin (retrait d'une éventuelle extension) réutilise la logique de `resume_helper.buildBaseDestPath` plutôt que d'en dupliquer une troisième variante. Risque connu : `filepath.Ext` sur un titre contenant un point (« Mr. Robot (2015) [1080p] ») ; on ne retire l'extension que lorsqu'elle est une extension média connue.

### D4. Un forcé par item et par run
Le scheduler construit au plus un `Item` par film/épisode. Quand plusieurs `DownloadInfo` forcés ciblent le même item, `mergeIncompleteDownloads` garde celui dont `CreatedAt` (puis `ID`) est le plus ancien ; les autres restent `pending`, non annulés, et sont pris aux runs suivants. Le choix est explicite plutôt que dépendant de `ORDER BY updated_at` de `GetIncompleteDownloads`, qui change à chaque mise à jour.

### D5. Stream synthétisé pour un film déjà téléchargé
Si le film n'a pas de stream tier-1 (fichier présent, donc non « missing » dans Radarr), le chemin existant de synthèse crée le stream ; il est complété par D3. `dedupTier2MovieStreams` retire le stream d'upgrade éventuel. Les résultats (rescan Radarr/Sonarr, notification Jellyfin) suivent le chemin normal d'un item téléchargé.

### D6. Spec `cancel-download-occurrence`
L'exception est ajoutée à l'exigence « réversible uniquement par reset complet » (MODIFIED), en nommant précisément le force-download de la même occurrence. L'annulation reste exclue de la sélection automatique : seule la réactivation par le handler la change.

## Risks / Trade-offs

- **Suffixe de destination figé** : si la langue ou la résolution de la ligne change entre le force et le run, le fichier garde le nom persisté. → Accepté : c'est le comportement voulu par la spec (chemin persisté avant transfert).
- **Deux fichiers dans le même dossier film** (ex. 720p VF existant + 1080p MULTI forcé) : prévu par la spec ; Radarr/Jellyfin gèrent plusieurs versions via les suffixes. → Aucun mitigation requise, à vérifier lors des tests manuels.
- **`DownloadInfo` annulés en prod (1705, 1706)** : ils ne se réactivent pas seuls. → Re-forcer depuis l'UI après déploiement (D1) ; aucune migration SQL.
- **Un forcé jamais repris** si le film reste confirmé non monitoré dans Radarr : il reste `pending` sans erreur visible. → Hors périmètre ; comportement déjà décrit par la spec existante.
- **Pas de test existant** pour « forcé + frère `downloaded` » dans le scheduler : couverture ajoutée dans les tâches, sinon la régression reviendrait silencieusement.

## Migration Plan

1. Déployer le correctif (image + chart habituels), sans migration de schéma.
2. Re-forcer Luca (item 220645) depuis l'UI ; vérifier que `download_info` 1705 repasse `pending`.
3. Au run planifié suivant (≤ 2 h), vérifier le fichier `Luca (2021) [1080p][MULTI]` à côté du 720p VF et le statut `completed`.
4. Rollback : redéployer l'image précédente ; les enregistrements réactivés restent valides (simple `pending`).

## Open Questions

- Faut-il, à terme, exposer les `cancelled` dans l'onglet Downloads pour rendre ce type d'échec visible ? (Hors périmètre ici.)
