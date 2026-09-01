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
}

export interface ProcessingLog {
  id: number;
  action: string;
  item_count: number;
  status: 'success' | 'failed' | 'in_progress';
  started_at: string;
  completed_at?: string;
  error_message?: string;
}

export interface DownloadInfo {
  id: number;
  url: string;
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'failed' | 'retrying';
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
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'failed' | 'retrying';
  download_path?: string;
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
}

export interface SonarrSeriesListItem {
  sonarr_id: number;
  title: string;
  year: number;
  matched_count: number;
  monitored_count: number;
}

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

export interface TMDBSearchResult {
  id: number;
  title: string;
  poster_path?: string;
  release_date?: string;
  overview?: string;
}
