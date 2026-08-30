## ADDED Requirements

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
