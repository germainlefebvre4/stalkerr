import { useState } from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import * as Progress from '@radix-ui/react-progress';
import { useTranslation } from 'react-i18next';
import { TFunction } from 'i18next';
import { Clock, Library, Download } from 'lucide-react';
import { ProcessingLog, StatsResponse, RadarrSonarrStats } from '../types';
import { formatDate, formatRunDuration } from '../utils/date';
import { useIsMobile } from '../hooks/useMediaQuery';
import radarrIcon from '../assets/icons/radarr.svg';
import sonarrIcon from '../assets/icons/sonarr.svg';

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

const GROUP_TITLES_VISIBLE_COUNT = 3;

// A run's statistics are only meaningful once every count field is present:
// a pre-migration processing_logs row has them all absent (undefined), while
// a run that legitimately processed nothing has them present and zero.
function hasRunStatistics(log: ProcessingLog): boolean {
  return log.movies_count !== undefined && log.movies_count !== null;
}

function renderStatusIndicator(isMobile: boolean, badgeClass: string, label: string) {
  if (isMobile) {
    return <span className={`status-dot ${badgeClass}`} role="img" title={label} aria-label={label} />;
  }
  return <span className={`badge ${badgeClass}`}>{label}</span>;
}

function GroupTitlesDisclosure({ groupTitles, t }: { groupTitles: string[] | null | undefined; t: TFunction }) {
  const [expanded, setExpanded] = useState(false);

  if (!groupTitles || groupTitles.length === 0) {
    return <div>{t('lastRun.groupTitles')}: {t('lastRun.noGroupTitles')}</div>;
  }

  const hasOverflow = groupTitles.length > GROUP_TITLES_VISIBLE_COUNT;
  const visibleTitles = expanded || !hasOverflow ? groupTitles : groupTitles.slice(0, GROUP_TITLES_VISIBLE_COUNT);

  return (
    <div>
      <div>{t('lastRun.groupTitles')}: {visibleTitles.join(', ')}</div>
      {hasOverflow && !expanded && (
        <button type="button" className="home-disclosure-btn" onClick={() => setExpanded(true)}>
          {t('lastRun.moreGroupTitles', { count: groupTitles.length - GROUP_TITLES_VISIBLE_COUNT })}
        </button>
      )}
    </div>
  );
}

function parsePercent(value: string): number {
  const parsed = parseInt(value, 10);
  return Number.isNaN(parsed) ? 0 : parsed;
}

export function HomeTab({
  latestLog, latestLogLoading,
  stats, getDownloadSuccessRatio,
  radarrSonarrStats, radarrSonarrStatsLoading, radarrSonarrStatsError,
  downloadsTotal, errorsTotal,
}: HomeTabProps) {
  const { t, i18n } = useTranslation('home');
  const { t: tCommon } = useTranslation('common');
  const isMobile = useIsMobile();

  const statsAvailable = latestLog != null && hasRunStatistics(latestLog);
  const duration = latestLog ? formatRunDuration(latestLog.started_at, latestLog.completed_at) : null;

  const lastRunBadgeClass = latestLog?.status === 'success' ? 'badge-success'
    : latestLog?.status === 'failed' ? 'badge-failed'
    : 'badge-progress';
  const lastRunStatusLabel = latestLog ? t(`lastRun.status.${latestLog.status}`, { defaultValue: latestLog.status }) : '';

  const radarrHasError = !!radarrSonarrStatsError || !!radarrSonarrStats?.radarr_error;
  const radarrLoading = radarrSonarrStatsLoading && radarrSonarrStats === null && !radarrSonarrStatsError;
  const sonarrHasError = !!radarrSonarrStatsError || !!radarrSonarrStats?.sonarr_error;
  const sonarrLoading = radarrSonarrStatsLoading && radarrSonarrStats === null && !radarrSonarrStatsError;

  const radarrMonitored = radarrSonarrStats?.radarr_monitored ?? null;
  const radarrMatched = radarrSonarrStats?.radarr_matched ?? null;
  const radarrMatchedRatio = radarrMonitored && radarrMonitored > 0 && radarrMatched !== null
    ? Math.round((radarrMatched / radarrMonitored) * 100)
    : 0;

  const sonarrMonitored = radarrSonarrStats?.sonarr_monitored ?? null;
  const sonarrMatched = radarrSonarrStats?.sonarr_matched ?? null;
  const sonarrMatchedRatio = sonarrMonitored && sonarrMonitored > 0 && sonarrMatched !== null
    ? Math.round((sonarrMatched / sonarrMonitored) * 100)
    : 0;

  const catalogSuccessPercent = stats ? parsePercent(getDownloadSuccessRatio()) : 0;

  return (
    <Tabs.Content value="home" className="card tab-panel">
      <div className="home-grid">
        <section className="home-card" aria-label={t('lastRun.title')}>
          <header className="home-card-header">
            <div className="home-card-heading">
              <Clock size={18} className="home-card-icon" aria-hidden="true" />
              <h3 className="home-card-title">{t('lastRun.title')}</h3>
            </div>
            {latestLog && renderStatusIndicator(isMobile, lastRunBadgeClass, lastRunStatusLabel)}
          </header>
          {latestLogLoading && latestLog === null ? (
            <div className="home-card-loading">{t('radarrSonarr.loading')}</div>
          ) : latestLog === null ? (
            <div className="home-empty-state">{t('lastRun.empty')}</div>
          ) : (
            <div className="home-fields">
              <div>
                <strong>{formatDate(latestLog.started_at, i18n.language)}</strong> {t('lastRun.startedAt')}
                {' · '}
                {duration === null ? t('lastRun.inProgress') : `${t('lastRun.duration')}: ${duration}`}
              </div>
              {!statsAvailable ? (
                <div className="home-unavailable">{t('lastRun.unavailable')}</div>
              ) : (
                <>
                  <div className="home-secondary-grid">
                    <div><strong>{latestLog.movies_count}</strong> {t('lastRun.movies')}</div>
                    <div><strong>{latestLog.tv_shows_count}</strong> {t('lastRun.tvShows')}</div>
                    <div><strong>{latestLog.new_items_count}</strong> {t('lastRun.newItems')}</div>
                    <div><strong>{latestLog.tmdb_matched_count}</strong> {t('lastRun.tmdbMatched')}</div>
                    <div><strong>{latestLog.tmdb_unmatched_count}</strong> {t('lastRun.tmdbUnmatched')}</div>
                  </div>
                  <GroupTitlesDisclosure groupTitles={latestLog.group_titles} t={t} />
                </>
              )}
            </div>
          )}
        </section>

        <section className="home-card" aria-label={t('catalog.title')}>
          <header className="home-card-header">
            <div className="home-card-heading">
              <Library size={18} className="home-card-icon" aria-hidden="true" />
              <h3 className="home-card-title">{t('catalog.title')}</h3>
            </div>
          </header>
          <div className="home-card-hero">{stats ? stats.total_items.toLocaleString() : '...'}</div>
          <div className="home-card-hero-label">{t('catalog.totalItems')}</div>
          <div className="home-fields">
            <div className="home-secondary-grid">
              <div><strong>{stats?.by_content_type ? (stats.by_content_type.movies || 0).toLocaleString() : '...'}</strong> {t('catalog.movies')}</div>
              <div><strong>{stats?.by_content_type ? (stats.by_content_type.tvshows || 0).toLocaleString() : '...'}</strong> {t('catalog.tvShows')}</div>
            </div>
            <div className="home-progress-row">
              <div className="home-progress-label">
                <span>{t('catalog.successRate')}</span>
                <span>{stats ? getDownloadSuccessRatio() : '...'}</span>
              </div>
              <Progress.Root value={catalogSuccessPercent} className="progress-root">
                <Progress.Indicator className="progress-indicator" style={{ width: `${catalogSuccessPercent}%` }} />
              </Progress.Root>
            </div>
          </div>
        </section>

        <section className="home-card" aria-label={t('radarrSonarr.title')}>
          <header className="home-card-header">
            <div className="home-card-heading">
              <h3 className="home-card-title">{t('radarrSonarr.title')}</h3>
            </div>
          </header>
          <div className="home-fields">
            <div className="home-subsection">
              <header className="home-card-header">
                <div className="home-card-heading">
                  <img src={radarrIcon} alt="" className="home-card-brand-icon" />
                  <h4 className="home-card-subheading">{t('radarrSonarr.radarrHeading')}</h4>
                </div>
                {!radarrLoading && renderStatusIndicator(
                  isMobile,
                  radarrHasError ? 'badge-failed' : 'badge-success',
                  radarrHasError ? t('lastRun.status.failed') : t('lastRun.status.success')
                )}
              </header>
              {radarrHasError ? (
                <div className="home-error">
                  {tCommon(`errors.${radarrSonarrStatsError || radarrSonarrStats?.radarr_error}`, { defaultValue: tCommon('errors.generic') })}
                </div>
              ) : radarrLoading ? (
                <div className="home-card-loading">{t('radarrSonarr.loading')}</div>
              ) : (
                <>
                  <div>
                    <strong>{radarrMonitored ?? '-'}</strong> {t('radarrSonarr.monitored')}, {' '}
                    <strong>{radarrMatched ?? '-'}</strong> {t('radarrSonarr.matched')}
                  </div>
                  <Progress.Root value={radarrMatchedRatio} className="progress-root">
                    <Progress.Indicator className="progress-indicator" style={{ width: `${radarrMatchedRatio}%` }} />
                  </Progress.Root>
                </>
              )}
            </div>
            <div className="home-subsection">
              <header className="home-card-header">
                <div className="home-card-heading">
                  <img src={sonarrIcon} alt="" className="home-card-brand-icon" />
                  <h4 className="home-card-subheading">{t('radarrSonarr.sonarrHeading')}</h4>
                </div>
                {!sonarrLoading && renderStatusIndicator(
                  isMobile,
                  sonarrHasError ? 'badge-failed' : 'badge-success',
                  sonarrHasError ? t('lastRun.status.failed') : t('lastRun.status.success')
                )}
              </header>
              {sonarrHasError ? (
                <div className="home-error">
                  {tCommon(`errors.${radarrSonarrStatsError || radarrSonarrStats?.sonarr_error}`, { defaultValue: tCommon('errors.generic') })}
                </div>
              ) : sonarrLoading ? (
                <div className="home-card-loading">{t('radarrSonarr.loading')}</div>
              ) : (
                <>
                  <div>
                    <strong>{sonarrMonitored ?? '-'}</strong> {t('radarrSonarr.monitored')}, {' '}
                    <strong>{sonarrMatched ?? '-'}</strong> {t('radarrSonarr.matched')}
                  </div>
                  <Progress.Root value={sonarrMatchedRatio} className="progress-root">
                    <Progress.Indicator className="progress-indicator" style={{ width: `${sonarrMatchedRatio}%` }} />
                  </Progress.Root>
                </>
              )}
            </div>
          </div>
        </section>

        <section className="home-card" aria-label={t('downloadsErrors.title')}>
          <header className="home-card-header">
            <div className="home-card-heading">
              <Download size={18} className="home-card-icon" aria-hidden="true" />
              <h3 className="home-card-title">{t('downloadsErrors.title')}</h3>
            </div>
          </header>
          <div className="home-card-hero">{downloadsTotal.toLocaleString()}</div>
          <div className="home-card-hero-label">{t('downloadsErrors.downloads')}</div>
          <div className="home-fields">
            <div>
              <span className={`badge ${errorsTotal > 0 ? 'badge-failed' : 'badge-neutral'}`}>
                {errorsTotal.toLocaleString()} {t('downloadsErrors.errors')}
              </span>
            </div>
          </div>
        </section>
      </div>
    </Tabs.Content>
  );
}
