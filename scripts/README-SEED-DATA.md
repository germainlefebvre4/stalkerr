# Test Data Seed Script

Ce script génère des données de test réalistes pour valider la fonctionnalité d'enrichissement des téléchargements.

## Données Générées

Le script crée **15 enregistrements de téléchargement** couvrant tous les cas d'usage :

### Films (9 items)

1. **La Cité de Dieu (2002)** - 1080p, MKV
   - ✅ Cas parfait : année présente, bonne résolution, format valide

2. **The Matrix (1999)** - 4K/2160p, MKV
   - ✅ Très haute qualité, année présente

3. **Avatar (2009)** - 1080p, MKV
   - ⚠️ **Problème** : Année manquante dans le path (`Avatar.BluRay.1080p.x264/`)

4. **Inception (2010)** - 720p, MP4
   - ⚠️ **Problème** : Année incorrecte dans le path (`Inception.2011.720p.BluRay/` au lieu de 2010)

5. **Pulp Fiction (1994)** - 480p, AVI
   - ⚠️ **Problème** : Basse qualité (480p)

6. **The Dark Knight (2008)** - Échec
   - ❌ **Statut** : `failed` avec erreur "HTTP 403 Forbidden - Link expired"
   - Téléchargement partiel : 1 GB / 4 GB

7. **Interstellar (2014)** - 1080p, MKV
   - ⏳ **Statut** : `downloading` (en cours, 2.5 GB / 6 GB - 41%)

8. **Forrest Gump (1994)** - FLV
   - ⚠️ **Problème** : Format peu courant (.flv)

9. **2001: A Space Odyssey (1968)** - 720p, MKV
   - ⚠️ **Problème** : Année ambiguë dans le path (`2001.A.Space.Odyssey.1968.720p.BluRay/`)
   - Le parser détectera "2001" au lieu de "1968"

### Séries TV (4 items)

10. **Breaking Bad S05E14** - 1080p, MKV
    - ✅ Cas parfait, série complète

11. **Game of Thrones S08E06** - 4K/2160p, MKV
    - ✅ Finale, très haute qualité

12. **The Office S02E01** - 720p
    - ❌ **Statut** : `failed` avec erreur "Network timeout"
    - Téléchargement partiel : 500 MB / 1 GB

13. **Stranger Things S04E09** - 720p, MKV
    - 🔄 **Statut** : `retrying` (2ème tentative, 897 MB / 1.5 GB)

### En Cours (2 items)

- **Interstellar** : downloading (41% complété)
- **Stranger Things S04E09** : retrying (55% complété, 2 tentatives)

## Scénarios de Test Couverts

| Scénario | Nombre | Exemples |
|----------|--------|----------|
| ✅ Téléchargements réussis | 9 | La Cité de Dieu, Matrix, Breaking Bad |
| ❌ Téléchargements échoués | 2 | The Dark Knight, The Office |
| ⏳ En cours / En retry | 2 | Interstellar, Stranger Things |
| ⚠️ Année manquante | 1 | Avatar |
| ⚠️ Année incorrecte | 2 | Inception, 2001 Space Odyssey |
| ⚠️ Basse qualité (≤480p) | 1 | Pulp Fiction |
| ⚠️ Format inhabituel | 1 | Forrest Gump (.flv) |
| 🎬 Films | 9 | - |
| 📺 Séries | 4 | - |
| 4K/2160p | 2 | Matrix, Game of Thrones |
| 1080p | 7 | La Cité de Dieu, Breaking Bad, etc. |
| 720p | 4 | Inception, 2001, Office, Stranger Things |
| 480p | 1 | Pulp Fiction |

## Utilisation

### Méthode 1 : Via Makefile (Recommandé)

```bash
make seed-test-data
```

Cette commande :
- Lit automatiquement la configuration depuis `config.yml`
- Se connecte à PostgreSQL
- Exécute le script SQL
- Affiche un résumé des données créées

### Méthode 2 : Directement avec psql

```bash
# Avec docker-compose
docker-compose exec postgres psql -U stalkeer -d stalkeer -f /scripts/seed-test-data.sql

# Avec PostgreSQL local
psql -h localhost -U stalkeer -d stalkeer -f scripts/seed-test-data.sql
```

## Nettoyage (Optionnel)

Si vous voulez nettoyer les données existantes avant de charger les données de test, décommentez les lignes au début du script :

```sql
-- Supprimer ces commentaires pour nettoyer :
DELETE FROM download_info WHERE id > 0;
DELETE FROM processed_lines WHERE id > 0;
DELETE FROM movies WHERE id > 0;
DELETE FROM tvshows WHERE id > 0;
```

⚠️ **Attention** : Cela supprimera TOUTES les données existantes !

## Vérification

Après avoir chargé les données, vérifiez avec :

```sql
-- Résumé des données
SELECT 
    (SELECT COUNT(*) FROM movies) as movies_count,
    (SELECT COUNT(*) FROM tvshows) as tvshows_count,
    (SELECT COUNT(*) FROM download_info) as downloads_count,
    (SELECT COUNT(*) FROM download_info WHERE status = 'completed') as completed,
    (SELECT COUNT(*) FROM download_info WHERE status = 'failed') as failed,
    (SELECT COUNT(*) FROM download_info WHERE status IN ('downloading', 'retrying')) as active;
```

Résultat attendu :
- **9 movies**
- **4 tvshows**
- **13 downloads** au total
- **9 complétés**
- **2 échoués**
- **2 actifs** (downloading/retrying)

## Test du Frontend

Avec ces données chargées, vous pouvez tester :

1. **Affichage enrichi** : Vérifier que les titres TMDB s'affichent correctement
2. **Filtres** :
   - Statut : `completed`, `failed`, `downloading`
   - Type : `movies`, `tvshows`
   - Problèmes : `missing_year`, `year_mismatch`, `low_quality`
3. **Indicateurs visuels** :
   - ✅ pour téléchargements réussis
   - ❌ pour échecs
   - ⚠️ pour problèmes détectés
4. **Parsing de fichiers** :
   - Extensions (.mkv, .mp4, .avi, .flv)
   - Résolutions (4K, 1080p, 720p, 480p)
   - Détection d'année dans les paths
5. **Messages d'erreur** : Vérifier l'affichage des erreurs pour les téléchargements échoués
6. **Progress bars** : Vérifier l'affichage de la progression pour les téléchargements en cours

## Notes Techniques

- Les données utilisent des ID TMDB réels pour les films/séries
- Les paths sont réalistes et suivent les conventions Plex/Jellyfin
- Les tailles de fichiers sont proportionnelles à la résolution
- Les timestamps sont échelonnés sur plusieurs jours
- Les erreurs incluent des messages réalistes (403, timeouts, etc.)
