import * as Tabs from '@radix-ui/react-tabs';
import { useTranslation } from 'react-i18next';
import { ProcessingLog, StatsResponse, RadarrSonarrStats } from '../types';
import { formatDate, formatRunDuration } from '../utils/date';

interface HomeTabProps {
  latestLog: ProcessingLog | null;
  latestLogLoading: boolean;
  stats: StatsResponse | null;
  getDownloadSuccessRatio: () => string;
  radarrSonarrStats: RadarrSonarrStats | null;
  radarrSonarrStatsLoading: boolean;
  radarrSonarrStatsError: string | null;
  downloadsTotal: number;
  errorsTotal: number;
}

// A run's statistics are only meaningful once every count field is present:
// a pre-migration processing_logs row has them all absent (undefined), while
// a run that legitimately processed nothing has them present and zero.
function hasRunStatistics(log: ProcessingLog): boolean {
  return log.movies_count !== undefined && log.movies_count !== null;
}

export function HomeTab({
  latestLog, latestLogLoading,
  stats, getDownloadSuccessRatio,
  radarrSonarrStats, radarrSonarrStatsLoading, radarrSonarrStatsError,
  downloadsTotal, errorsTotal,
}: HomeTabProps) {
  const { t, i18n } = useTranslation('home');
  const { t: tCommon } = useTranslation('common');

  const statsAvailable = latestLog != null && hasRunStatistics(latestLog);
  const duration = latestLog ? formatRunDuration(latestLog.started_at, latestLog.completed_at) : null;

  return (
    <Tabs.Content value="home" className="card tab-panel">
      <div className="home-grid">
        <section className="home-card" aria-label={t('lastRun.title')}>
          <h3 className="home-card-title">{t('lastRun.title')}</h3>
          {latestLogLoading && latestLog === null ? (
            <div className="home-card-loading">{t('radarrSonarr.loading')}</div>
          ) : latestLog === null ? (
            <div className="home-empty-state">{t('lastRun.empty')}</div>
          ) : (
            <div className="home-fields">
              <div><strong>{formatDate(latestLog.started_at, i18n.language)}</strong> {t('lastRun.startedAt')}</div>
              <div>
                <strong>{t(`lastRun.status.${latestLog.status}`, { defaultValue: latestLog.status })}</strong>
                {' - '}
                {duration === null ? t('lastRun.inProgress') : `${t('lastRun.duration')}: ${duration}`}
              </div>
              {!statsAvailable ? (
                <div className="home-unavailable">{t('lastRun.unavailable')}</div>
              ) : (
                <>
                  <div><strong>{latestLog.movies_count}</strong> {t('lastRun.movies')}</div>
                  <div><strong>{latestLog.tv_shows_count}</strong> {t('lastRun.tvShows')}</div>
                  <div><strong>{latestLog.new_items_count}</strong> {t('lastRun.newItems')}</div>
                  <div><strong>{latestLog.tmdb_matched_count}</strong> {t('lastRun.tmdbMatched')}</div>
                  <div><strong>{latestLog.tmdb_unmatched_count}</strong> {t('lastRun.tmdbUnmatched')}</div>
                  <div>
                    {t('lastRun.groupTitles')}: {latestLog.group_titles && latestLog.group_titles.length > 0
                      ? latestLog.group_titles.join(', ')
                      : t('lastRun.noGroupTitles')}
                  </div>
                </>
              )}
            </div>
          )}
        </section>

        <section className="home-card" aria-label={t('catalog.title')}>
          <h3 className="home-card-title">{t('catalog.title')}</h3>
          <div className="home-fields">
            <div><strong>{stats ? stats.total_items.toLocaleString() : '...'}</strong> {t('catalog.totalItems')}</div>
            <div><strong>{stats?.by_content_type ? (stats.by_content_type.movies || 0).toLocaleString() : '...'}</strong> {t('catalog.movies')}</div>
            <div><strong>{stats?.by_content_type ? (stats.by_content_type.tvshows || 0).toLocaleString() : '...'}</strong> {t('catalog.tvShows')}</div>
            <div><strong>{stats ? getDownloadSuccessRatio() : '...'}</strong> {t('catalog.successRate')}</div>
          </div>
        </section>

        <section className="home-card" aria-label={t('radarrSonarr.title')}>
          <h3 className="home-card-title">{t('radarrSonarr.title')}</h3>
          <div className="home-fields">
            <div>
              <h4 className="home-card-subheading">{t('radarrSonarr.radarrHeading')}</h4>
              {radarrSonarrStatsError ? (
                <div className="home-error">{tCommon(`errors.${radarrSonarrStatsError}`, { defaultValue: tCommon('errors.generic') })}</div>
              ) : radarrSonarrStats?.radarr_error ? (
                <div className="home-error">{tCommon(`errors.${radarrSonarrStats.radarr_error}`, { defaultValue: tCommon('errors.generic') })}</div>
              ) : radarrSonarrStatsLoading && radarrSonarrStats === null ? (
                <div className="home-card-loading">{t('radarrSonarr.loading')}</div>
              ) : (
                <div>
                  <strong>{radarrSonarrStats?.radarr_monitored ?? '-'}</strong> {t('radarrSonarr.monitored')}, {' '}
                  <strong>{radarrSonarrStats?.radarr_matched ?? '-'}</strong> {t('radarrSonarr.matched')}
                </div>
              )}
            </div>
            <div>
              <h4 className="home-card-subheading">{t('radarrSonarr.sonarrHeading')}</h4>
              {radarrSonarrStatsError ? (
                <div className="home-error">{tCommon(`errors.${radarrSonarrStatsError}`, { defaultValue: tCommon('errors.generic') })}</div>
              ) : radarrSonarrStats?.sonarr_error ? (
                <div className="home-error">{tCommon(`errors.${radarrSonarrStats.sonarr_error}`, { defaultValue: tCommon('errors.generic') })}</div>
              ) : radarrSonarrStatsLoading && radarrSonarrStats === null ? (
                <div className="home-card-loading">{t('radarrSonarr.loading')}</div>
              ) : (
                <div><strong>{radarrSonarrStats?.sonarr_monitored ?? '-'}</strong> {t('radarrSonarr.monitored')}</div>
              )}
            </div>
          </div>
        </section>

        <section className="home-card" aria-label={t('downloadsErrors.title')}>
          <h3 className="home-card-title">{t('downloadsErrors.title')}</h3>
          <div className="home-fields">
            <div><strong>{downloadsTotal.toLocaleString()}</strong> {t('downloadsErrors.downloads')}</div>
            <div><strong>{errorsTotal.toLocaleString()}</strong> {t('downloadsErrors.errors')}</div>
          </div>
        </section>
      </div>
    </Tabs.Content>
  );
}
