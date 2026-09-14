import { SettingsGroupFieldSpec } from './components/SettingsGroupCard';

// Field-group definitions shared between the Intégrations/Notifications/
// Avancé tab sections (which render them) and ConfigurationPage (which
// needs their keys to compute per-tab override counts). Kept in their own
// module, rather than exported alongside a component, so those component
// files stay fast-refresh-friendly (react-refresh/only-export-components).

export const RADARR_FIELDS: SettingsGroupFieldSpec[] = [
  { key: 'radarr.enabled', label: 'radarr.enabled', type: 'boolean' },
  { key: 'radarr.url', label: 'radarr.url' },
  { key: 'radarr.api_key', label: 'radarr.apiKey' },
  { key: 'radarr.sync_interval', label: 'radarr.syncInterval', type: 'number' },
  { key: 'radarr.quality_profile_id', label: 'radarr.qualityProfileId', type: 'number' },
];

export const SONARR_FIELDS: SettingsGroupFieldSpec[] = [
  { key: 'sonarr.enabled', label: 'sonarr.enabled', type: 'boolean' },
  { key: 'sonarr.url', label: 'sonarr.url' },
  { key: 'sonarr.api_key', label: 'sonarr.apiKey' },
  { key: 'sonarr.sync_interval', label: 'sonarr.syncInterval', type: 'number' },
  { key: 'sonarr.quality_profile_id', label: 'sonarr.qualityProfileId', type: 'number' },
];

export const TMDB_FIELDS: SettingsGroupFieldSpec[] = [
  { key: 'tmdb.enabled', label: 'tmdb.enabled', type: 'boolean' },
  { key: 'tmdb.api_key', label: 'tmdb.apiKey' },
  { key: 'tmdb.language', label: 'tmdb.language' },
  { key: 'tmdb.requests_per_second', label: 'tmdb.requestsPerSecond', type: 'number' },
];

export const JELLYFIN_FIELDS: SettingsGroupFieldSpec[] = [
  { key: 'jellyfin.enabled', label: 'jellyfin.enabled', type: 'boolean' },
  { key: 'jellyfin.url', label: 'jellyfin.url' },
  { key: 'jellyfin.api_key', label: 'jellyfin.apiKey' },
];

export const NOTIFICATIONS_FIELDS: SettingsGroupFieldSpec[] = [
  { key: 'notifications.enabled', label: 'enabled', type: 'boolean' },
  { key: 'notifications.ntfy.enabled', label: 'ntfyEnabled', type: 'boolean' },
  { key: 'notifications.ntfy.server_url', label: 'ntfyServerUrl' },
  { key: 'notifications.ntfy.topic', label: 'ntfyTopic' },
  { key: 'notifications.ntfy.auth_token', label: 'ntfyAuthToken' },
];

export const DOWNLOADS_FIELDS: SettingsGroupFieldSpec[] = [
  { key: 'downloads.movies_path', label: 'moviesPath' },
  { key: 'downloads.tvshows_path', label: 'tvshowsPath' },
  { key: 'downloads.temp_dir', label: 'tempDir' },
  { key: 'downloads.max_parallel', label: 'maxParallel', type: 'number' },
  { key: 'downloads.timeout', label: 'timeout', type: 'number' },
  { key: 'downloads.retry_attempts', label: 'retryAttempts', type: 'number' },
  { key: 'downloads.resume_enabled', label: 'resumeEnabled', type: 'boolean' },
  { key: 'downloads.progress_interval_mb', label: 'progressIntervalMb', type: 'number' },
  { key: 'downloads.progress_interval_seconds', label: 'progressIntervalSeconds', type: 'number' },
  { key: 'downloads.lock_timeout_minutes', label: 'lockTimeoutMinutes', type: 'number' },
  { key: 'downloads.max_retry_attempts', label: 'maxRetryAttempts', type: 'number' },
  { key: 'downloads.min_file_size_mb', label: 'minFileSizeMb', type: 'number' },
  { key: 'downloads.force_tier_probability', label: 'forceTierProbability', type: 'number' },
];

export const LOGGING_FIELDS: SettingsGroupFieldSpec[] = [
  { key: 'logging.app.level', label: 'appLevel' },
  { key: 'logging.database.level', label: 'databaseLevel' },
  { key: 'logging.format', label: 'format' },
];

export const M3U_FIELDS: SettingsGroupFieldSpec[] = [
  { key: 'm3u.update_interval', label: 'updateInterval', type: 'number' },
];

export const INTEGRATIONS_SETTINGS_KEYS: string[] = [
  ...RADARR_FIELDS, ...SONARR_FIELDS, ...TMDB_FIELDS, ...JELLYFIN_FIELDS,
].map(f => f.key);

export const NOTIFICATIONS_SETTINGS_KEYS: string[] = NOTIFICATIONS_FIELDS.map(f => f.key);

export const ADVANCED_SETTINGS_KEYS: string[] = [
  ...DOWNLOADS_FIELDS, ...LOGGING_FIELDS, ...M3U_FIELDS,
].map(f => f.key);
