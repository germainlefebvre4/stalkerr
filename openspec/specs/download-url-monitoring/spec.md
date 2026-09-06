## Purpose

This capability enables monitoring of download URL status by storing the URL directly in `DownloadInfo`, tracking retry counts between each attempt, and allowing SQL queries on download state by URL without any JOIN.

## Requirements

### Requirement: URL stockée dans DownloadInfo
Le système SHALL stocker l'URL de téléchargement directement dans `DownloadInfo` lors de la création du record, afin de permettre des requêtes SQL sans JOIN sur `processed_lines`.

#### Scenario: Création d'un DownloadInfo avec URL
- **WHEN** un téléchargement est initié avec une URL non vide
- **THEN** le record `DownloadInfo` créé en base contient cette URL dans le champ `url`

#### Scenario: Record DownloadInfo existant
- **WHEN** un `DownloadInfo` existe déjà pour une `ProcessedLine` donnée
- **THEN** l'URL n'est pas modifiée (idempotent)

### Requirement: retry_count incrémenté une fois par tentative externe, sur tous les chemins
Le système SHALL incrémenter `retry_count` de 1 dans `DownloadInfo` à l'issue de chaque tentative externe de téléchargement (un appel complet à `Download()`, qu'il ait effectué un ou plusieurs essais HTTP internes) qui se termine par un échec — que cette tentative provienne du matching normal de contenu manquant ou du chemin de reprise (resume) des téléchargements incomplets, et indépendamment du fait que l'erreur rencontrée soit elle-même classée « retryable » pour les sous-tentatives internes. Une tentative qui se termine avec succès n'incrémente pas `retry_count`.

#### Scenario: Échec non-retryable compte comme une tentative pleine
- **WHEN** une tentative de téléchargement échoue avec une erreur non classée retryable (par exemple `unexpected EOF`), sans avoir déclenché de sous-tentative interne
- **THEN** `retry_count` est incrémenté de 1 en base à l'issue de cette tentative

#### Scenario: Plusieurs sous-tentatives internes ne comptent qu'une seule fois
- **WHEN** une tentative de téléchargement épuise ses sous-tentatives internes (par exemple 3 essais HTTP suite à des erreurs retryable) avant d'échouer définitivement
- **THEN** `retry_count` est incrémenté de 1 en base à l'issue de cette tentative, pas une fois par sous-tentative interne

#### Scenario: Une tentative issue du chemin de reprise compte de la même façon
- **WHEN** une tentative de téléchargement déclenchée par le chemin de reprise (resume) des téléchargements incomplets échoue
- **THEN** `retry_count` est incrémenté de la même façon que pour une tentative issue du matching normal de contenu manquant

#### Scenario: Un succès n'incrémente pas retry_count
- **WHEN** une tentative de téléchargement se termine avec succès
- **THEN** `retry_count` reste inchangé

#### Scenario: Erreur de tracking n'interrompt pas le download
- **WHEN** l'incrémentation de `retry_count` échoue (erreur DB)
- **THEN** le téléchargement continue normalement, sans que cette erreur de tracking ne bloque ou n'annule la tentative en cours

### Requirement: Queryabilité SQL de l'état par URL
Le système SHALL permettre de requêter l'état, le nombre de tentatives et le message d'erreur d'un téléchargement directement sur la table `download_info` par URL, sans JOIN.

#### Scenario: Requête état par URL
- **WHEN** une requête SQL `SELECT * FROM download_info WHERE url = '<url>'` est exécutée
- **THEN** elle retourne le statut, retry_count, error_message, started_at, completed_at du téléchargement

#### Scenario: Requête agrégée succès/échec
- **WHEN** une requête SQL `SELECT status, COUNT(*) FROM download_info GROUP BY status` est exécutée
- **THEN** elle retourne les comptes corrects par statut sans JOIN nécessaire

### Requirement: error_message effacé au redémarrage d'une tentative
Le système SHALL effacer `DownloadInfo.error_message` (le remettre à `null`) lorsqu'un `DownloadInfo` transitionne vers le statut `downloading` pour une nouvelle tentative de téléchargement, afin qu'un message d'erreur issu d'une tentative précédente ne survive pas à une tentative ultérieure sur le même enregistrement.

#### Scenario: Nouvelle tentative après un échec précédent
- **WHEN** un `DownloadInfo` dont `status` est `failed` et `error_message` est non nul reçoit une nouvelle tentative de téléchargement (transition vers `downloading`)
- **THEN** `error_message` est remis à `null` dès cette transition, avant même de connaître l'issue de la nouvelle tentative

#### Scenario: Tentative réussie après un échec précédent
- **WHEN** un `DownloadInfo` dont `status` était `failed` avec un `error_message` non nul termine avec succès une nouvelle tentative (transition vers `completed`)
- **THEN** le record en base a `status = 'completed'` et `error_message = null`

#### Scenario: Première tentative (pas d'erreur préalable)
- **WHEN** un `DownloadInfo` nouvellement créé transitionne vers `downloading` pour sa première tentative
- **THEN** `error_message` reste `null` (aucun changement de comportement lorsqu'il n'y a pas d'erreur préalable à effacer)
