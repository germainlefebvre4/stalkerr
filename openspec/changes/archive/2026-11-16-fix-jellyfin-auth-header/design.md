# Design

## Context

Le client Jellyfin (`internal/external/jellyfin/jellyfin.go`) construit toutes ses requêtes via un unique `newRequest`, qui pose `X-Emby-Token`. `SystemStatus`, `Sessions` et `postMediaUpdated` passent tous par là, donc un seul point de changement suffit. Voir proposal.md pour la motivation.

Vérifié en live depuis le pod stalkerr contre Jellyfin 12.4.0 (clé fournie hors artefacts) :

| Requête | `X-Emby-Token` | `Authorization: MediaBrowser Token="…"` |
|---|---|---|
| `GET /System/Info` | 401 | 200 |
| `GET /Sessions` | 401 | 200 |
| `POST /Library/Media/Updated` | non testé (effet de bord : rescan) | non testé |

## Goals / Non-Goals

**Goals:**
- Authentifier toutes les requêtes Jellyfin avec le schéma `MediaBrowser`.
- Couvrir le header envoyé par des tests, pour que la régression ne repasse pas inaperçue.

**Non-Goals:**
- Corriger l'envoi d'une clé vide par le front pour un champ sensible déjà enregistré. Conforme à la spec `integration-connectivity-test` actuelle, donc une autre change.
- Ajouter les paramètres facultatifs du schéma (`Client`, `Device`, `DeviceId`, `Version`).

## Decisions

**Pas de repli sur `X-Emby-Token`.** Le schéma `MediaBrowser` est le format documenté de longue date par Jellyfin et fonctionne sur les anciennes versions. Un repli (essayer un header, puis l'autre sur 401) doublerait les requêtes et masquerait de vraies clés invalides. Alternative écartée : envoyer les deux headers en même temps, qui laisserait le header legacy dans les logs et les proxys sans bénéfice.

**Seul `Token` est envoyé.** C'est ce qui a été validé en live (200). Les champs `Client`/`Device`/`DeviceId`/`Version` ne sont pas requis pour une clé API ; les ajouter sans besoin élargirait ce qu'il faut tester.

**Un seul point de changement : `newRequest`.** Évite de dupliquer la logique par endpoint et garantit que tout nouvel appel hérite du bon header.

**Clé vide : on envoie quand même la requête.** Garde le comportement actuel et le classement `unauthorized` de la spec `integration-connectivity-test`, sans cas particulier.

## Risks / Trade-offs

- [Une clé contenant un guillemet casserait la valeur du header] → Une clé API Jellyfin est une chaîne hexadécimale ; un guillemet ne se produit qu'avec une saisie invalide, qui serait de toute façon rejetée par Jellyfin. Aucun échappement n'est ajouté.
- [`POST /Library/Media/Updated` n'est pas vérifié en live] → Il partage `newRequest` avec les deux endpoints vérifiés, et un test unitaire contrôle le header envoyé. Une vérification manuelle reste possible pendant un téléchargement réel.
- [Jellyfin très ancien qui n'accepterait pas le schéma `MediaBrowser`] → Jugé improbable, le schéma existe depuis longtemps. Si un cas réel apparaît, il justifiera une change dédiée plutôt qu'un repli préventif.
