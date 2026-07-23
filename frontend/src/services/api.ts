import { 
  PaginatedResponse, 
  PlaylistItem, 
  ProcessingLog, 
  DownloadEnriched, 
  ConfigPaths, 
  StatsResponse, 
  FilterConfig 
} from '../types';

export const api = {
  async getHealth(): Promise<{ status: string }> {
    const res = await fetch('/health');
    if (!res.ok) throw new Error('Failed to fetch health');
    return res.json();
  },

  async getStats(): Promise<StatsResponse> {
    const res = await fetch('/api/v1/stats');
    if (!res.ok) throw new Error('Failed to fetch stats');
    return res.json();
  },

  async getConfigPaths(): Promise<ConfigPaths> {
    const res = await fetch('/api/v1/config/paths');
    if (!res.ok) throw new Error('Failed to fetch config paths');
    return res.json();
  },

  async getPlaylist(
    page: number, 
    limit: number = 15, 
    contentType?: 'all' | 'movies' | 'tvshows', 
    stateFilter?: string, 
    search?: string
  ): Promise<PaginatedResponse<PlaylistItem>> {
    let url = `/api/v1/items?limit=${limit}&offset=${(page - 1) * limit}`;
    if (contentType && contentType !== 'all') {
      url += `&content_type=${contentType}`;
    }
    if (stateFilter && stateFilter !== 'all') {
      url += `&state=${stateFilter}`;
    }
    if (search) {
      url += `&group_title=${encodeURIComponent(search)}`;
    }
    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to fetch playlist');
    return res.json();
  },

  async getLogs(limit: number = 20): Promise<PaginatedResponse<ProcessingLog>> {
    const res = await fetch(`/api/v1/processing-logs?limit=${limit}`);
    if (!res.ok) throw new Error('Failed to fetch processing logs');
    return res.json();
  },

  async getDownloads(
    limit: number = 20, 
    status?: string, 
    type?: string, 
    problem?: string
  ): Promise<PaginatedResponse<DownloadEnriched>> {
    let url = `/api/v1/downloads?limit=${limit}`;
    if (status) url += `&status=${status}`;
    if (type) url += `&type=${type}`;
    if (problem) url += `&problem=${problem}`;
    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to fetch downloads');
    return res.json();
  },

  async getFilters(): Promise<{ filters: FilterConfig[] }> {
    const res = await fetch('/api/v1/filters');
    if (!res.ok) throw new Error('Failed to fetch filters');
    return res.json();
  },

  async resetPipeline(id: number, contentType: string): Promise<any> {
    const endpoint = contentType === 'movies'
      ? `/api/v1/movies/${id}/reset`
      : `/api/v1/tvshows/${id}/reset`;

    const res = await fetch(endpoint, { method: 'POST' });
    if (!res.ok) {
      throw new Error('La réinitialisation a échoué');
    }
    return res.json();
  },

  async searchTMDB(query: string, type: 'movie' | 'tvshow', year?: string): Promise<any[]> {
    const res = await fetch(`/api/v1/tmdb/search?query=${encodeURIComponent(query)}&type=${type}&year=${year || ''}`);
    if (!res.ok) {
      throw new Error('La recherche a échoué');
    }
    return res.json();
  },

  async forceOverride(
    id: number, 
    payload: { tmdb_id: number; type: 'movie' | 'tvshow'; season: number | null; episode: number | null }
  ): Promise<any> {
    const res = await fetch(`/api/v1/items/${id}/override`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      throw new Error("L'association forcée a échoué");
    }
    return res.json();
  },

  async moveFolder(id: number, type: 'movie' | 'tvshow', destDir: string): Promise<any> {
    const endpoint = type === 'movie' 
      ? `/api/v1/movies/${id}/move`
      : `/api/v1/tvshows/${id}/move`;

    const res = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ destination_parent_dir: destDir }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Le déplacement a échoué');
    }
    return res.json();
  },

  async createFilter(payload: { 
    name: string; 
    attribute: string; 
    include_patterns?: string; 
    exclude_patterns?: string; 
  }): Promise<any> {
    const res = await fetch('/api/v1/filters', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'La création du filtre a échoué');
    }
    return res.json();
  },

  async deleteFilter(id: number): Promise<any> {
    const res = await fetch(`/api/v1/filters/${id}`, { method: 'DELETE' });
    if (!res.ok) {
      throw new Error('La suppression du filtre a échoué');
    }
    return res.json();
  }
};
