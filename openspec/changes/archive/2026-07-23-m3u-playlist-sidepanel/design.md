## Context

Les utilisateurs du tableau de bord de Stalkeer ont besoin de comprendre précisément comment les pistes de leur playlist M3U ont été ingérées et associées aux films ou séries. Actuellement, la table de la playlist présente les colonnes principales mais ne permet pas de visualiser les métadonnées brutes issues du fichier d'origine (comme le contenu exact de la ligne `#EXTINF`, l'URL du flux d'origine et le numéro de la ligne dans le fichier). 

Nous concevons un système d'exploration interactif sous la forme d'un panneau latéral (sidepanel) accessible au clic sur n'importe quelle ligne de la playlist.

## Goals / Non-Goals

**Goals:**
- **Capturer le numéro de ligne d'origine** lors du parsing de la playlist M3U et le persister en base de données.
- **Exposer l'intégralité des métadonnées d'ingestion brute** via les réponses de l'API `/api/v1/items`.
- **Créer une interface utilisateur interactive de type tiroir (Drawer)** fluide, accessible et élégante, affichant les détails enrichis du média et les données brutes d'origine.
- **Ajouter des raccourcis pratiques** (copie en un clic) pour l'URL du flux et la ligne d'ingestion brute.

**Non-Goals:**
- Recalculer ou backfiller rétroactivement les numéros de ligne des pistes déjà importées (les lignes existantes vaudront `0` par défaut).
- Permettre l'édition directe du contenu brut d'ingestion depuis l'IHM.

## Decisions

### Décision 1 : Extension du modèle de base de données `ProcessedLine`
Nous ajoutons un champ `LineNumber` au modèle Go `ProcessedLine`. 
- **Alternative considérée** : Conserver uniquement `LineContent` et parser l'info dynamiquement à la volée.
- **Raison du choix** : L'indexation par numéro de ligne à l'ingestion est immédiate, précise, ne coûte qu'un entier en base de données, et permet un tri ou un filtrage ultra-rapide si nécessaire.

### Décision 2 : Inclusion des données brutes directement dans l'API de liste `/api/v1/items`
L'API `/api/v1/items` transmettra directement les informations d'ingestion brute au sein du DTO `ItemResponse`.
- **Alternative considérée** : Faire un appel supplémentaire de type `GET /api/v1/items/:id` lors du clic sur une ligne.
- **Raison du choix** : La taille de `LineContent` et des autres attributs bruts d'une entrée est négligeable (quelques octets). Les inclure directement évite un aller-retour réseau (roundtrip) supplémentaire vers le backend à l'ouverture du sidepanel, offrant une expérience utilisateur instantanée et extrêmement réactive.

### Décision 3 : Utilisation de Radix UI `Dialog` pour le Sidepanel (Drawer)
Nous réutilisons le composant `Dialog` de `@radix-ui/react-dialog` (déjà présent dans les dépendances du projet) pour l'implémentation du tiroir latéral, associé à des animations CSS personnalisées de glissement horizontal.
- **Alternative considérée** : Un composant avec état CSS `visible` fait maison.
- **Raison du choix** : Radix UI gère parfaitement l'accessibilité : fermeture par touche Échap, piège de focus (focus trap) pour la navigation clavier, et gestion sémantique des lecteurs d'écran, tout en nous laissant le contrôle total sur le design du tiroir via notre fichier de styles CSS existant.

## Risks / Trade-offs

- **[Risque] Conflit de clics sur les lignes de la table** → Le clic sur une ligne pour ouvrir le sidepanel pourrait accidentellement déclencher les boutons d'action existants ("Associer" ou "Réinitialiser") présents en bout de ligne.
  - *Atténuation* : Utilisation systématique de `e.stopPropagation()` sur les gestionnaires de clic des boutons d'action et du badge d'état.
  
- **[Risque] Affichage de lignes d'ingestion vierges pour les anciennes entrées** → Les anciennes entrées n'auront pas de `line_number` (valeur par défaut `0`).
  - *Atténuation* : Cacher ou adapter l'affichage du numéro de ligne dans le sidepanel si sa valeur est inférieure ou égale à `0` (ex: afficher "Inconnue (Ancien import)").
