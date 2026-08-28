import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { StatsResponse } from '../types';

interface StatsKPICardsProps {
  stats: StatsResponse | null;
  getDownloadSuccessRatio: () => string;
}

export function StatsKPICards({ stats, getDownloadSuccessRatio }: StatsKPICardsProps) {
  const { t } = useTranslation();
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <>
      <button
        type="button"
        className="kpi-toggle-row"
        aria-expanded={isExpanded}
        onClick={() => setIsExpanded(prev => !prev)}
      >
        <span>{t('kpi.toggle.label')}</span>
        <span aria-hidden="true">{isExpanded ? '▴' : '▾'}</span>
      </button>
      <section className={`kpi-grid${isExpanded ? '' : ' kpi-grid--collapsed'}`}>
        <div className="kpi-card">
          <span className="kpi-card-title">🗒️ {t('kpi.playlist.title')}</span>
          <span className="kpi-card-value">
            {stats ? stats.total_items.toLocaleString() : '...'}
          </span>
          <span className="kpi-card-subtitle">{t('kpi.playlist.subtitle')}</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-card-title">🎬 {t('kpi.movies.title')}</span>
          <span className="kpi-card-value">
            {stats && stats.by_content_type ? (stats.by_content_type.movies || 0).toLocaleString() : '...'}
          </span>
          <span className="kpi-card-subtitle">{t('kpi.movies.subtitle')}</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-card-title">📺 {t('kpi.tvshows.title')}</span>
          <span className="kpi-card-value">
            {stats && stats.by_content_type ? (stats.by_content_type.tvshows || 0).toLocaleString() : '...'}
          </span>
          <span className="kpi-card-subtitle">{t('kpi.tvshows.subtitle')}</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-card-title">📥 {t('kpi.successRate.title')}</span>
          <span className="kpi-card-value">
            {stats ? getDownloadSuccessRatio() : '...'}
          </span>
          <span className="kpi-card-subtitle">
            {stats && stats.by_state
              ? t('kpi.successRate.detail', { downloaded: stats.by_state.downloaded || 0, failed: stats.by_state.failed || 0 })
              : t('kpi.successRate.fallback')}
          </span>
        </div>
      </section>
    </>
  );
}
