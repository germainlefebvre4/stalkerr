## MODIFIED Requirements

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
