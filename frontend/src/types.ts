export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
  total_pages: number;
}

export interface MovieResponse {
  id: number;
  tmdb_id: number;
  tmdb_title: string;
  tmdb_year: number;
  genres?: string;
  duration?: number;
  poster_path?: string;
  overview?: string;
  imdb_id?: string;
  tvdb_id?: number;
}

export interface TVShowResponse {
  id: number;
  tmdb_id: number;
  tmdb_title: string;
  tmdb_year: number;
  genres?: string;
  season?: number;
  episode?: number;
  poster_path?: string;
  overview?: string;
  imdb_id?: string;
  tvdb_id?: number;
}

export interface PlaylistItem {
  id: number;
  tvg_name: string;
  group_title: string;
  content_type: 'movies' | 'tvshows' | 'channels' | 'uncategorized';
  state: string;
  movie?: MovieResponse;
  tvshow?: TVShowResponse;
  override_by?: string;
  override_at?: string;
  line_content: string;
  line_url?: string;
  line_hash: string;
  line_number: number;
  processing_log_id?: number;
  created_at: string;
  downloaded_at: string | null;
}

export interface MediaGroupItem {
  type: 'movie' | 'tvshow' | 'unmatched_movies' | 'unmatched_tvshows';
  movie_id?: number;
  tmdb_id?: number;
  title?: string;
  year?: number;
  season_start?: number;
  season_end?: number;
  latest_activity: string;
  latest_processing_log_id?: number;
}

export interface ProcessingLog {
  id: number;
  action: string;
  item_count: number;
  status: 'success' | 'failed' | 'in_progress';
  started_at: string;
  completed_at?: string;
  error_message?: string;
  movies_count?: number;
  tv_shows_count?: number;
  new_items_count?: number;
  tmdb_matched_count?: number;
  tmdb_unmatched_count?: number;
  group_titles?: string[] | null;
}

export interface DownloadInfo {
  id: number;
  url: string;
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'failed' | 'retrying' | 'cancelled';
  download_path?: string;
  file_size?: number;
  bytes_downloaded?: number;
  total_bytes?: number;
  retry_count: number;
  error_message?: string;
  updated_at: string;
}

export interface DownloadEnriched {
  id: number;
  url: string;
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'failed' | 'retrying' | 'cancelled';
  download_path?: string;
  target_path?: string;
  staging_path?: string;
  file_size?: number;
  bytes_downloaded?: number;
  total_bytes?: number;
  retry_count: number;
  error_message?: string;
  updated_at: string;
  completed_at?: string;

  content?: {
    type: 'movies' | 'tvshows' | 'channels' | 'uncategorized';
    title: string;
    year?: number;
    resolution?: string;
    season?: number;
    episode?: number;
    genres?: string;
    duration?: number;
  };
  
  file_info?: {
    extension: string;
    folder_name: string;
    file_name: string;
    has_year_in_path: boolean;
    year_mismatch: boolean;
    detected_year?: number;
    detected_resolution?: string;
    is_valid_format: boolean;
  };

  rename_folder_name?: string;
}

export interface RenameDownloadResponse {
  status: string;
  new_path: string;
}

export type ResyncPathOutcome = 'corrected' | 'already_up_to_date' | 'not_managed_by_radarr_sonarr';

export interface ResyncPathResponse {
  status: ResyncPathOutcome;
  old_path?: string;
  new_path?: string;
}

export interface CancelDownloadResponse {
  status: string;
  download_id: number;
}

export interface ConfigPaths {
  movies_path: string;
  tvshows_path: string;
}

export interface StatsResponse {
  total_items: number;
  by_content_type: Record<string, number>;
  by_state: Record<string, number>;
}

export interface FilterConfig {
  id: number;
  name: string;
  attribute: string;
  include_patterns?: string;
  exclude_patterns?: string;
  is_runtime: boolean;
}

export interface FilterOriginEntry {
  attribute: string;
  include_patterns: string[];
  exclude_patterns: string[];
}

// What the shared FilterTestDrawer/FilterTestPanelBody are currently testing:
// either one attribute's patterns (the create dialog's in-progress values, or
// a card's origin/override patterns) or both attributes together ("Tester
// l'ensemble"). The drawer is open exactly when this is non-null; `label` is
// the contextual title the owning component computes for its own context.
export type FilterTestTarget =
  | { mode: 'single'; attribute: 'group_title' | 'tvg_name'; includePatterns: string; excludePatterns: string; label: string }
  | { mode: 'combined'; groupTitleInclude: string; groupTitleExclude: string; tvgNameInclude: string; tvgNameExclude: string; label: string };

// Dry-run testing of a (not necessarily saved) attribute/pattern combination
// against a source's latest downloaded archive. See filter-dry-run-test.
// `attributes` lists which of "group_title"/"tvg_name" are in scope (1 or 2
// entries); an attribute not listed imposes no filtering. `search_attribute`
// is required when both attributes are supplied and `search` is set.
export interface FilterDryRunRequest {
  source_name: string;
  attributes: string[];
  group_title_include_patterns?: string;
  group_title_exclude_patterns?: string;
  tvg_name_include_patterns?: string;
  tvg_name_exclude_patterns?: string;
  search?: string;
  search_attribute?: string;
}

export interface FilterDryRunValueCount {
  value: string;
  count: number;
}

// The aggregate-summary shape (no content search requested) for a
// single-attribute request. When no_archive is true, every other field is
// zero-valued.
export interface FilterDryRunSummaryResponse {
  no_archive: boolean;
  total_lines: number;
  matched_count: number;
  excluded_count: number;
  top_matched: FilterDryRunValueCount[];
  top_excluded: FilterDryRunValueCount[];
}

// The aggregate-summary shape for a combined (both attributes) request: a
// cause-ventilated breakdown plus each attribute's own top values. When
// no_archive is true, every other field is zero-valued.
export interface FilterDryRunCombinedSummaryResponse {
  no_archive: boolean;
  total_lines: number;
  kept_count: number;
  excluded_by_group_title_only: number;
  excluded_by_tvg_name_only: number;
  excluded_by_both: number;
  group_title_top_matched: FilterDryRunValueCount[];
  group_title_top_excluded: FilterDryRunValueCount[];
  tvg_name_top_matched: FilterDryRunValueCount[];
  tvg_name_top_excluded: FilterDryRunValueCount[];
}

// `matched` is populated in single-attribute mode; `verdict` is populated in
// combined mode ("kept" | "excluded_by_group_title" | "excluded_by_tvg_name"
// | "excluded_by_both").
export interface FilterDryRunResultLine {
  group_title: string;
  tvg_name: string;
  matched: boolean;
  verdict?: string;
}

// The content-search shape (a non-empty `search` was requested), shared by
// single-attribute and combined modes.
export interface FilterDryRunSearchResponse {
  no_archive: boolean;
  results: FilterDryRunResultLine[];
  truncated: boolean;
}

export interface OccurrenceResponse {
  id: number;
  resolution?: string;
  state: string;
}

export interface RadarrMovieListItem {
  radarr_id: number;
  title: string;
  year: number;
  has_file: boolean;
  matched: boolean;
  movie_id?: number;
  occurrence_count: number;
}

export interface SonarrSeriesListItem {
  sonarr_id: number;
  title: string;
  year: number;
  matched_count: number;
  monitored_count: number;
  occurrence_count: number;
}

export type MatchStatusFilter = '' | 'matched' | 'no_match';

export interface RadarrMovieMatchesResponse {
  matched: boolean;
  movie?: MovieResponse;
  occurrences: OccurrenceResponse[];
}

export interface SonarrSeriesEpisodeItem {
  season: number;
  episode: number;
  matched: boolean;
  occurrences: OccurrenceResponse[];
}

export interface SonarrSeriesEpisodesResponse {
  episodes: SonarrSeriesEpisodeItem[];
}

export interface RadarrSonarrStats {
  radarr_monitored: number | null;
  radarr_matched: number | null;
  radarr_error?: string;
  sonarr_monitored: number | null;
  sonarr_matched: number | null;
  sonarr_error?: string;
}

export type ServiceStatusState = 'ok' | 'ko' | 'not_configured';

export interface ServiceStatus {
  status: ServiceStatusState;
  reason?: string;
}

export interface DiskUsageEntry {
  paths: string[];
  available?: number;
  free?: number;
  total?: number;
  used_pct?: number;
  unavailable?: boolean;
  reason?: string;
}

export interface SystemStatusResponse {
  database: ServiceStatus;
  radarr: ServiceStatus;
  sonarr: ServiceStatus;
  tmdb: ServiceStatus;
  disk: DiskUsageEntry[];
  version: string;
  commit: string;
  date: string;
}

export type IntegrationTestService = 'radarr' | 'sonarr' | 'jellyfin';

/** On-demand connectivity-test result: reuses ServiceStatus's status/reason shape (the endpoint only ever returns 'ok' or 'ko', never 'not_configured'). */
export type IntegrationTestResult = ServiceStatus;

export interface TMDBSearchResult {
  id: number;
  title: string;
  poster_path?: string;
  release_date?: string;
  overview?: string;
}

// Effective origin of a settings/M3U-source value: "interface" when a
// stored override applies, "config" when the file/env/default value does.
export type SettingsOrigin = 'interface' | 'config';

export interface SettingsField {
  key: string;
  value?: string | number | boolean;
  is_set?: boolean; // present only for sensitive fields, in place of value
  sensitive: boolean;
  origin: SettingsOrigin;
  restart_required: boolean;
}

export interface M3uSource {
  name: string;
  file_path: string;
  enabled: boolean;
  url: string;
  archive_dir: string;
  retention_count: number;
  max_file_size_mb: number;
  timeout_seconds: number;
  retry_attempts: number;
  auth_username: string;
  has_auth_password: boolean;
  is_runtime: boolean;
}

// auth_password omitted entirely (not just empty) means "keep the current
// effective password"; an explicit empty string clears it.
export interface M3uSourceInput {
  file_path: string;
  enabled: boolean;
  url: string;
  archive_dir: string;
  retention_count: number;
  max_file_size_mb: number;
  timeout_seconds: number;
  retry_attempts: number;
  auth_username: string;
  auth_password?: string;
}
