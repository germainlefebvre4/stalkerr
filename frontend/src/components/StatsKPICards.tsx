import { StatsResponse } from '../types';

interface StatsKPICardsProps {
  stats: StatsResponse | null;
  getDownloadSuccessRatio: () => string;
}

export function StatsKPICards({ stats, getDownloadSuccessRatio }: StatsKPICardsProps) {
  return (
    <section className="kpi-grid">
      <div className="kpi-card">
        <span className="kpi-card-title">🗒️ Playlist M3U</span>
        <span className="kpi-card-value">
          {stats ? stats.total_items.toLocaleString() : '...'}
        </span>
        <span className="kpi-card-subtitle">Flux enregistrés en DB</span>
      </div>
      <div className="kpi-card">
        <span className="kpi-card-title">🎬 Films identifiés</span>
        <span className="kpi-card-value">
          {stats && stats.by_content_type ? (stats.by_content_type.movies || 0).toLocaleString() : '...'}
        </span>
        <span className="kpi-card-subtitle">TMDB enrichis</span>
      </div>
      <div className="kpi-card">
        <span className="kpi-card-title">📺 Séries suivies</span>
        <span className="kpi-card-value">
          {stats && stats.by_content_type ? (stats.by_content_type.tvshows || 0).toLocaleString() : '...'}
        </span>
        <span className="kpi-card-subtitle">Saisons et épisodes</span>
      </div>
      <div className="kpi-card">
        <span className="kpi-card-title">📥 Taux de Réussite</span>
        <span className="kpi-card-value">
          {stats ? getDownloadSuccessRatio() : '...'}
        </span>
        <span className="kpi-card-subtitle">
          {stats && stats.by_state ? `${stats.by_state.downloaded || 0} réussis / ${stats.by_state.failed || 0} échecs` : 'Téléchargements'}
        </span>
      </div>
    </section>
  );
}
