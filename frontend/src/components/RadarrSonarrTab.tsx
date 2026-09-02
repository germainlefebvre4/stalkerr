import React from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { RadarrMovieListItem, SonarrSeriesListItem, RadarrMovieMatchesResponse, SonarrSeriesEpisodesResponse, SonarrSeriesEpisodeItem, OccurrenceResponse, PlaylistItem, RadarrSonarrStats, MatchStatusFilter } from '../types';
import { useIsMobile } from '../hooks/useMediaQuery';
import { useRadarrSonarrView } from '../hooks/useRadarrSonarrView';
import { getProcessingStatus, getDownloadStatus, getProcessingStatusBadgeClass, getDownloadStatusBadgeClass } from '../utils/pipelineState';
import { Pagination } from './Pagination';
import { MediaOccurrenceDrawer, MediaOccurrenceDrawerBody } from './MediaOccurrenceDrawer';
import { api, ApiError } from '../services/api';
import radarrIcon from '../assets/icons/radarr.svg';
import sonarrIcon from '../assets/icons/sonarr.svg';

interface RadarrSonarrTabProps {
  filmsItems: RadarrMovieListItem[];
  filmsLoading: boolean;
  filmsError: string | null;
  filmsTotal: number;
  filmsPage: number;
  setFilmsPage: React.Dispatch<React.SetStateAction<number>>;
  filmsLimit: number;
  fetchFilms: () => void;
  filmsSearch: string;
  setFilmsSearch: (value: string) => void;
  filmsFilter: MatchStatusFilter;
  setFilmsFilter: (value: MatchStatusFilter) => void;

  seriesItems: SonarrSeriesListItem[];
  seriesLoading: boolean;
  seriesError: string | null;
  seriesTotal: number;
  seriesPage: number;
  setSeriesPage: React.Dispatch<React.SetStateAction<number>>;
  seriesLimit: number;
  fetchSeries: () => void;
  refreshSeries: () => void;
  seriesSearch: string;
  setSeriesSearch: (value: string) => void;
  seriesFilter: MatchStatusFilter;
  setSeriesFilter: (value: MatchStatusFilter) => void;

  stats: RadarrSonarrStats | null;
  statsLoading: boolean;
  statsError: string | null;
  fetchStats: () => void;
}

const SEARCH_DEBOUNCE_MS = 300;

function seriesBadgeClass(matched: number, monitored: number): string {
  if (monitored === 0 || matched === 0) return 'badge-pending';
  if (matched === monitored) return 'badge-success';
  return 'badge-progress';
}

function padNumber(n: number): string {
  return String(n).padStart(2, '0');
}

export function RadarrSonarrTab({
  filmsItems, filmsLoading, filmsError, filmsTotal, filmsPage, setFilmsPage, filmsLimit, fetchFilms,
  filmsSearch, setFilmsSearch, filmsFilter, setFilmsFilter,
  seriesItems, seriesLoading, seriesError, seriesTotal, seriesPage, setSeriesPage, seriesLimit, fetchSeries, refreshSeries,
  seriesSearch, setSeriesSearch, seriesFilter, setSeriesFilter,
  stats, statsLoading, statsError, fetchStats,
}: RadarrSonarrTabProps) {
  const { t } = useTranslation('radarrSonarr');
  const { t: tCommon } = useTranslation('common');
  const { t: tPlaylist } = useTranslation('playlist');
  const isMobile = useIsMobile();

  const { activeSubTab, setActiveSubTab } = useRadarrSonarrView();

  const [filmsSearchInput, setFilmsSearchInput] = React.useState(filmsSearch);
  const [seriesSearchInput, setSeriesSearchInput] = React.useState(seriesSearch);

  React.useEffect(() => {
    const handle = setTimeout(() => setFilmsSearch(filmsSearchInput), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(handle);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filmsSearchInput]);

  React.useEffect(() => {
    const handle = setTimeout(() => setSeriesSearch(seriesSearchInput), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(handle);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [seriesSearchInput]);

  const [selectedMovie, setSelectedMovie] = React.useState<RadarrMovieListItem | null>(null);
  const [selectedSeries, setSelectedSeries] = React.useState<SonarrSeriesListItem | null>(null);
  const [movieDetail, setMovieDetail] = React.useState<RadarrMovieMatchesResponse | null>(null);
  const [seriesDetail, setSeriesDetail] = React.useState<SonarrSeriesEpisodesResponse | null>(null);
  const [detailLoading, setDetailLoading] = React.useState(false);
  const [detailError, setDetailError] = React.useState<string | null>(null);
  const [expandedSeason, setExpandedSeason] = React.useState<number | null>(null);
  const [expandedEpisode, setExpandedEpisode] = React.useState<number | null>(null);
  const [detailItem, setDetailItem] = React.useState<PlaylistItem | null>(null);
  const [drawerView, setDrawerView] = React.useState<'occurrences' | 'detail'>('occurrences');

  React.useEffect(() => {
    if (!selectedMovie) return;
    void Promise.resolve().then(() => {
      setDetailLoading(true);
      setDetailError(null);
      setMovieDetail(null);
      api.getRadarrMovieMatches(selectedMovie.radarr_id)
        .then(setMovieDetail)
        .catch(err => setDetailError(err instanceof ApiError ? err.code : 'generic'))
        .finally(() => setDetailLoading(false));
    });
  }, [selectedMovie]);

  React.useEffect(() => {
    if (!selectedSeries) return;
    void Promise.resolve().then(() => {
      setDetailLoading(true);
      setDetailError(null);
      setSeriesDetail(null);
      api.getSonarrSeriesEpisodes(selectedSeries.sonarr_id)
        .then(setSeriesDetail)
        .catch(err => setDetailError(err instanceof ApiError ? err.code : 'generic'))
        .finally(() => setDetailLoading(false));
    });
  }, [selectedSeries]);

  const openMovie = (item: RadarrMovieListItem) => {
    setSelectedSeries(null);
    setSelectedMovie(item);
    setExpandedSeason(null);
    setExpandedEpisode(null);
    setDetailItem(null);
    setDrawerView('occurrences');
  };
  const openSeries = (item: SonarrSeriesListItem) => {
    setSelectedMovie(null);
    setSelectedSeries(item);
    setExpandedSeason(null);
    setExpandedEpisode(null);
    setDetailItem(null);
    setDrawerView('occurrences');
  };
  const closeDrawer = () => {
    setSelectedMovie(null);
    setSelectedSeries(null);
    setMovieDetail(null);
    setSeriesDetail(null);
    setExpandedSeason(null);
    setExpandedEpisode(null);
    setDetailItem(null);
    setDrawerView('occurrences');
  };

  const seasonGroups = React.useMemo(() => {
    if (!seriesDetail) return [];
    const bySeason = new Map<number, SonarrSeriesEpisodeItem[]>();
    for (const ep of seriesDetail.episodes) {
      const list = bySeason.get(ep.season);
      if (list) list.push(ep);
      else bySeason.set(ep.season, [ep]);
    }
    return Array.from(bySeason.entries())
      .sort(([a], [b]) => a - b)
      .map(([season, episodes]) => {
        const sortedEpisodes = [...episodes].sort((a, b) => a.episode - b.episode);
        const matchedCount = sortedEpisodes.filter(ep => ep.matched).length;
        return { season, matchedCount, totalCount: sortedEpisodes.length, episodes: sortedEpisodes };
      });
  }, [seriesDetail]);

  const handleOccurrenceClick = (occurrenceId: number) => {
    api.getItem(occurrenceId)
      .then(resolved => {
        setDetailItem(resolved);
        if (isMobile) setDrawerView('detail');
      })
      .catch(err => console.error('Failed to load occurrence detail', err));
  };

  const handleBackToOccurrences = () => {
    setDrawerView('occurrences');
    setDetailItem(null);
  };

  const renderOccurrences = (occurrences: OccurrenceResponse[]) => {
    if (isMobile) {
      return (
        <div>
          {occurrences.map(occ => (
            <div key={occ.id} className="mobile-list-card" onClick={() => handleOccurrenceClick(occ.id)}>
              <div className="mobile-list-card-main">
                <span className="mobile-list-card-title">{occ.resolution || t('drawer.unknownResolution')}</span>
              </div>
              <div style={{ display: 'flex', gap: '0.4rem', flexShrink: 0 }}>
                <span className={`badge ${getProcessingStatusBadgeClass(getProcessingStatus(occ.state))}`}>{getProcessingStatus(occ.state)}</span>
                <span className={`badge ${getDownloadStatusBadgeClass(getDownloadStatus(occ.state))}`}>{getDownloadStatus(occ.state) === 'not_downloaded' ? tPlaylist('pipelineStatus.notDownloaded') : getDownloadStatus(occ.state)}</span>
              </div>
            </div>
          ))}
        </div>
      );
    }
    return (
      <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
        <table className="custom-table">
          <thead>
            <tr>
              <th>{t('drawer.occurrenceResolution')}</th>
              <th>{tPlaylist('drawer.processingStatus')}</th>
              <th>{tPlaylist('drawer.downloadStatus')}</th>
            </tr>
          </thead>
          <tbody>
            {occurrences.map(occ => (
              <tr key={occ.id} className="clickable-row" onClick={() => handleOccurrenceClick(occ.id)}>
                <td>{occ.resolution || t('drawer.unknownResolution')}</td>
                <td><span className={`badge ${getProcessingStatusBadgeClass(getProcessingStatus(occ.state))}`}>{getProcessingStatus(occ.state)}</span></td>
                <td><span className={`badge ${getDownloadStatusBadgeClass(getDownloadStatus(occ.state))}`}>{getDownloadStatus(occ.state) === 'not_downloaded' ? tPlaylist('pipelineStatus.notDownloaded') : getDownloadStatus(occ.state)}</span></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  const filmsEmptyMessage = filmsSearch ? t('films.noResults') : t('films.empty');
  const seriesEmptyMessage = seriesSearch ? t('series.noResults') : t('series.empty');

  return (
    <Tabs.Content value="radarr-sonarr" className="card tab-panel">
      <Tabs.Root value={activeSubTab} onValueChange={(value) => setActiveSubTab(value as 'resume' | 'radarr' | 'sonarr')}>
        <Tabs.List className="segmented-tabs-list radarr-sonarr-subtabs">
          <Tabs.Trigger value="resume" className="segmented-tabs-trigger">{t('subTabs.resume')}</Tabs.Trigger>
          <Tabs.Trigger value="radarr" className="segmented-tabs-trigger"><img src={radarrIcon} alt="" className="tab-icon" />{t('subTabs.radarr')}</Tabs.Trigger>
          <Tabs.Trigger value="sonarr" className="segmented-tabs-trigger"><img src={sonarrIcon} alt="" className="tab-icon" />{t('subTabs.sonarr')}</Tabs.Trigger>
        </Tabs.List>

        {/* Sous-onglet Résumé */}
        <Tabs.Content value="resume" className="tab-panel">
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.25rem' }}>
            <div className="table-flush" style={{ padding: '1.5rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 800, color: 'var(--primary-slate)', marginBottom: '1rem' }}>{t('resume.radarrHeading')}</h3>
              {statsError ? (
                <div style={{ textAlign: 'center', color: 'var(--status-failed-text)' }}>
                  <div>{tCommon(`errors.${statsError}`, { defaultValue: tCommon('errors.generic') })}</div>
                  <button onClick={fetchStats} className="btn-secondary" style={{ marginTop: '0.75rem' }}>{t('retry')}</button>
                </div>
              ) : stats?.radarr_error ? (
                <div style={{ color: 'var(--status-failed-text)' }}>
                  {tCommon(`errors.${stats.radarr_error}`, { defaultValue: tCommon('errors.generic') })}
                </div>
              ) : statsLoading && stats === null ? (
                <div style={{ color: 'var(--text-secondary)' }}>{t('resume.loading')}</div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  <div><strong>{stats?.radarr_monitored ?? '-'}</strong> {t('resume.monitored')}</div>
                  <div><strong>{stats?.radarr_matched ?? '-'}</strong> {t('resume.matched')}</div>
                  <div><strong>{stats && stats.radarr_monitored !== null && stats.radarr_matched !== null ? stats.radarr_monitored - stats.radarr_matched : '-'}</strong> {t('resume.unmatched')}</div>
                </div>
              )}
            </div>

            <div className="table-flush" style={{ padding: '1.5rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 800, color: 'var(--primary-slate)', marginBottom: '1rem' }}>{t('resume.sonarrHeading')}</h3>
              {statsError ? (
                <div style={{ textAlign: 'center', color: 'var(--status-failed-text)' }}>
                  <div>{tCommon(`errors.${statsError}`, { defaultValue: tCommon('errors.generic') })}</div>
                  <button onClick={fetchStats} className="btn-secondary" style={{ marginTop: '0.75rem' }}>{t('retry')}</button>
                </div>
              ) : stats?.sonarr_error ? (
                <div style={{ color: 'var(--status-failed-text)' }}>
                  {tCommon(`errors.${stats.sonarr_error}`, { defaultValue: tCommon('errors.generic') })}
                </div>
              ) : statsLoading && stats === null ? (
                <div style={{ color: 'var(--text-secondary)' }}>{t('resume.loading')}</div>
              ) : (
                <div><strong>{stats?.sonarr_monitored ?? '-'}</strong> {t('resume.monitored')}</div>
              )}
            </div>
          </div>
        </Tabs.Content>

        {/* Sous-onglet Radarr (Films) */}
        <Tabs.Content value="radarr" className="tab-panel">
          <section>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
              <div>
                <h2 style={{ fontSize: '1.15rem', fontWeight: 800, color: 'var(--primary-slate)' }}>{t('films.heading')}</h2>
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('films.subtitle')}</p>
              </div>
              <button onClick={fetchFilms} className="btn-secondary">{t('films.refresh')}</button>
            </div>

            <div style={{ display: 'flex', gap: '0.75rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
              <input
                type="text"
                placeholder={t('films.searchPlaceholder')}
                value={filmsSearchInput}
                onChange={e => setFilmsSearchInput(e.target.value)}
                className="custom-input"
                style={{ flex: 1, minWidth: '200px' }}
              />
              <select
                value={filmsFilter}
                onChange={e => setFilmsFilter(e.target.value as MatchStatusFilter)}
                style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: '#fff', fontWeight: 600, color: 'var(--text-secondary)' }}
              >
                <option value="">{t('filterStatus.all')}</option>
                <option value="matched">{t('filterStatus.matched')}</option>
                <option value="no_match">{t('filterStatus.noMatch')}</option>
              </select>
            </div>

            {filmsError ? (
              <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--status-failed-text)' }}>
                <div>{tCommon(`errors.${filmsError}`, { defaultValue: tCommon('errors.generic') })}</div>
                <button onClick={fetchFilms} className="btn-secondary" style={{ marginTop: '0.75rem' }}>{t('retry')}</button>
              </div>
            ) : isMobile ? (
              <div>
                {filmsLoading && filmsItems.length === 0 ? (
                  <div className="mobile-list-empty">{t('films.loading')}</div>
                ) : filmsItems.length === 0 ? (
                  <div className="mobile-list-empty">{filmsEmptyMessage}</div>
                ) : (
                  filmsItems.map(item => (
                    <div key={item.radarr_id} className="mobile-list-card" onClick={() => openMovie(item)}>
                      <div className="mobile-list-card-main">
                        <span className="mobile-list-card-title">{item.title}</span>
                        <span className="mobile-list-card-subtitle">{item.year} · {t('films.table.occurrences')}: {item.occurrence_count}</span>
                      </div>
                      <span className={`badge ${item.matched ? 'badge-success' : 'badge-pending'}`}>
                        {item.matched ? t('films.badge.matched') : t('films.badge.unmatched')}
                      </span>
                    </div>
                  ))
                )}
              </div>
            ) : (
              <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
                <table className="custom-table">
                  <thead>
                    <tr>
                      <th>{t('films.table.title')}</th>
                      <th>{t('films.table.year')}</th>
                      <th>{t('films.table.status')}</th>
                      <th>{t('films.table.occurrences')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filmsLoading && filmsItems.length === 0 ? (
                      <tr><td colSpan={4} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('films.loading')}</td></tr>
                    ) : filmsItems.length === 0 ? (
                      <tr><td colSpan={4} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{filmsEmptyMessage}</td></tr>
                    ) : (
                      filmsItems.map(item => (
                        <tr key={item.radarr_id} className="clickable-row" onClick={() => openMovie(item)}>
                          <td style={{ fontWeight: 700, color: 'var(--primary-slate)' }}>{item.title}</td>
                          <td>{item.year}</td>
                          <td>
                            <span className={`badge ${item.matched ? 'badge-success' : 'badge-pending'}`}>
                              {item.matched ? t('films.badge.matched') : t('films.badge.unmatched')}
                            </span>
                          </td>
                          <td>{item.occurrence_count}</td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            )}

            {!filmsError && (
              <Pagination total={filmsTotal} page={filmsPage} setPage={setFilmsPage} limit={filmsLimit} />
            )}
          </section>
        </Tabs.Content>

        {/* Sous-onglet Sonarr (Séries) */}
        <Tabs.Content value="sonarr" className="tab-panel">
          <section>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
              <div>
                <h2 style={{ fontSize: '1.15rem', fontWeight: 800, color: 'var(--primary-slate)' }}>{t('series.heading')}</h2>
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('series.subtitle')}</p>
              </div>
              <button onClick={refreshSeries} className="btn-secondary">{t('series.refresh')}</button>
            </div>

            <div style={{ display: 'flex', gap: '0.75rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
              <input
                type="text"
                placeholder={t('series.searchPlaceholder')}
                value={seriesSearchInput}
                onChange={e => setSeriesSearchInput(e.target.value)}
                className="custom-input"
                style={{ flex: 1, minWidth: '200px' }}
              />
              <select
                value={seriesFilter}
                onChange={e => setSeriesFilter(e.target.value as MatchStatusFilter)}
                style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: '#fff', fontWeight: 600, color: 'var(--text-secondary)' }}
              >
                <option value="">{t('filterStatus.all')}</option>
                <option value="matched">{t('filterStatus.matched')}</option>
                <option value="no_match">{t('filterStatus.noMatch')}</option>
              </select>
            </div>

            {seriesError ? (
              <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--status-failed-text)' }}>
                <div>{tCommon(`errors.${seriesError}`, { defaultValue: tCommon('errors.generic') })}</div>
                <button onClick={fetchSeries} className="btn-secondary" style={{ marginTop: '0.75rem' }}>{t('retry')}</button>
              </div>
            ) : isMobile ? (
              <div>
                {seriesLoading && seriesItems.length === 0 ? (
                  <div className="mobile-list-empty">{t('series.loading')}</div>
                ) : seriesItems.length === 0 ? (
                  <div className="mobile-list-empty">{seriesEmptyMessage}</div>
                ) : (
                  seriesItems.map(item => (
                    <div key={item.sonarr_id} className="mobile-list-card" onClick={() => openSeries(item)}>
                      <div className="mobile-list-card-main">
                        <span className="mobile-list-card-title">{item.title}</span>
                        <span className="mobile-list-card-subtitle">{item.year} · {t('series.table.occurrences')}: {item.occurrence_count}</span>
                      </div>
                      <span className={`badge ${seriesBadgeClass(item.matched_count, item.monitored_count)}`}>
                        {t('series.ratio', { matched: item.matched_count, monitored: item.monitored_count })}
                      </span>
                    </div>
                  ))
                )}
              </div>
            ) : (
              <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
                <table className="custom-table">
                  <thead>
                    <tr>
                      <th>{t('series.table.title')}</th>
                      <th>{t('series.table.year')}</th>
                      <th>{t('series.table.status')}</th>
                      <th>{t('series.table.occurrences')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {seriesLoading && seriesItems.length === 0 ? (
                      <tr><td colSpan={4} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('series.loading')}</td></tr>
                    ) : seriesItems.length === 0 ? (
                      <tr><td colSpan={4} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{seriesEmptyMessage}</td></tr>
                    ) : (
                      seriesItems.map(item => (
                        <tr key={item.sonarr_id} className="clickable-row" onClick={() => openSeries(item)}>
                          <td style={{ fontWeight: 700, color: 'var(--primary-slate)' }}>{item.title}</td>
                          <td>{item.year}</td>
                          <td>
                            <span className={`badge ${seriesBadgeClass(item.matched_count, item.monitored_count)}`}>
                              {t('series.ratio', { matched: item.matched_count, monitored: item.monitored_count })}
                            </span>
                          </td>
                          <td>{item.occurrence_count}</td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            )}

            {!seriesError && (
              <Pagination total={seriesTotal} page={seriesPage} setPage={setSeriesPage} limit={seriesLimit} />
            )}
          </section>
        </Tabs.Content>
      </Tabs.Root>

      {/* Sidepanel de Détails (Film ou Série) */}
      <Dialog.Root open={!!selectedMovie || !!selectedSeries} onOpenChange={(open) => !open && closeDrawer()}>
        <Dialog.Portal>
          <Dialog.Overlay className="drawer-overlay" />
          <Dialog.Content
            className="drawer-content"
            onInteractOutside={(e) => {
              if ((e.target as HTMLElement)?.closest?.('.drawer-content--secondary')) {
                e.preventDefault();
              }
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1rem' }}>
              <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)', margin: 0 }}>
                {isMobile && drawerView === 'detail'
                  ? tPlaylist('drawer.title')
                  : (selectedMovie ? t('drawer.movieTitle') : t('drawer.seriesTitle'))}
              </Dialog.Title>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                {isMobile && drawerView === 'detail' && (
                  <button
                    type="button"
                    className="btn-secondary"
                    style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem', borderRadius: 'var(--radius-sm)' }}
                    onClick={handleBackToOccurrences}
                  >
                    {t('drawer.backToOccurrences')}
                  </button>
                )}
                <Dialog.Close className="btn-secondary" style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem', borderRadius: 'var(--radius-sm)' }}>
                  {t('drawer.close')}
                </Dialog.Close>
              </div>
            </div>
            <Dialog.Description style={{ display: 'none' }}>{t('drawer.description')}</Dialog.Description>

            {isMobile && drawerView === 'detail' && detailItem ? (
              <MediaOccurrenceDrawerBody item={detailItem} />
            ) : (
              <>
                {detailLoading && (
                  <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('drawer.loading')}</div>
                )}

                {detailError && (
                  <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--status-failed-text)' }}>{t('drawer.loadFailed')}</div>
                )}

                {!detailLoading && !detailError && selectedMovie && movieDetail && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: 'var(--primary-accent)' }}>
                      {selectedMovie.title} <span style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', fontWeight: 600 }}>({selectedMovie.year})</span>
                    </div>

                    {!movieDetail.matched ? (
                      <div style={{ backgroundColor: 'var(--bg-app)', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)', padding: '1rem', textAlign: 'center', color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem' }}>
                        {t('drawer.noMatch')}
                      </div>
                    ) : (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                        <h3 className="drawer-section-title">{t('drawer.occurrencesSection')}</h3>
                        {movieDetail.occurrences.length === 0 ? (
                          <div style={{ color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem' }}>{t('drawer.noOccurrences')}</div>
                        ) : renderOccurrences(movieDetail.occurrences)}
                      </div>
                    )}
                  </div>
                )}

                {!detailLoading && !detailError && selectedSeries && seriesDetail && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
                    <div style={{ fontSize: '1.1rem', fontWeight: 800, color: 'var(--primary-accent)' }}>
                      {selectedSeries.title} <span style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', fontWeight: 600 }}>({selectedSeries.year})</span>
                    </div>

                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                      <h3 className="drawer-section-title">{t('drawer.episodesSection')}</h3>
                      {seasonGroups.length === 0 ? (
                        <div style={{ color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem' }}>{t('series.empty')}</div>
                      ) : (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                          {seasonGroups.map(group => {
                            const isSeasonExpanded = expandedSeason === group.season;
                            return (
                              <div key={group.season} style={{ border: '1px solid var(--border-color)', borderRadius: 'var(--radius-md)', overflow: 'hidden' }}>
                                <div
                                  className={`clickable-row${isSeasonExpanded ? ' clickable-row--active' : ''}`}
                                  style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.85rem 1.25rem' }}
                                  onClick={() => {
                                    setExpandedSeason(isSeasonExpanded ? null : group.season);
                                    setExpandedEpisode(null);
                                  }}
                                >
                                  <span style={{ fontWeight: 700, color: 'var(--primary-slate)' }}>
                                    <span style={{ marginRight: '0.5rem', display: 'inline-block', transform: isSeasonExpanded ? 'rotate(90deg)' : 'none' }}>▸</span>
                                    <span>{t('drawer.season', { season: group.season })}</span>
                                  </span>
                                  <span className={`badge ${seriesBadgeClass(group.matchedCount, group.totalCount)}`}>
                                    {t('series.ratio', { matched: group.matchedCount, monitored: group.totalCount })}
                                  </span>
                                </div>
                                {isSeasonExpanded && (
                                  isMobile ? (
                                    <div style={{ padding: '0.75rem', backgroundColor: 'var(--bg-app)', borderTop: '1px solid var(--border-color)' }}>
                                      {group.episodes.map(ep => {
                                        const isEpExpanded = expandedEpisode === ep.episode;
                                        return (
                                          <React.Fragment key={ep.episode}>
                                            <div
                                              className={`mobile-list-card${isEpExpanded ? ' clickable-row--active' : ''}`}
                                              onClick={() => setExpandedEpisode(isEpExpanded ? null : ep.episode)}
                                            >
                                              <div className="mobile-list-card-main">
                                                <span className="mobile-list-card-title">
                                                  <span style={{ marginRight: '0.4rem', display: 'inline-block', transform: isEpExpanded ? 'rotate(90deg)' : 'none' }}>▸</span>
                                                  <span>{t('drawer.episode', { season: padNumber(ep.season), episode: padNumber(ep.episode) })}</span>
                                                </span>
                                              </div>
                                              <span className={`badge ${ep.matched ? 'badge-success' : 'badge-pending'}`}>
                                                {ep.matched ? t('drawer.episodeMatched') : t('drawer.episodeUnmatched')}
                                              </span>
                                            </div>
                                            {isEpExpanded && (
                                              <div style={{ marginBottom: '0.75rem', padding: '0.5rem', backgroundColor: 'var(--bg-card)', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-md)' }}>
                                                {ep.occurrences.length === 0 ? (
                                                  <div style={{ padding: '0.5rem', color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem' }}>
                                                    {t('drawer.episodeNoOccurrences')}
                                                  </div>
                                                ) : renderOccurrences(ep.occurrences)}
                                              </div>
                                            )}
                                          </React.Fragment>
                                        );
                                      })}
                                    </div>
                                  ) : (
                                    <table className="custom-table" style={{ borderTop: '1px solid var(--border-color)' }}>
                                      <thead>
                                        <tr>
                                          <th></th>
                                          <th>{t('drawer.occurrenceState')}</th>
                                        </tr>
                                      </thead>
                                      <tbody>
                                        {group.episodes.map(ep => {
                                          const isEpExpanded = expandedEpisode === ep.episode;
                                          return (
                                            <React.Fragment key={ep.episode}>
                                              <tr
                                                className={`clickable-row${isEpExpanded ? ' clickable-row--active' : ''}`}
                                                onClick={() => setExpandedEpisode(isEpExpanded ? null : ep.episode)}
                                              >
                                                <td style={{ fontWeight: 600 }}>
                                                  <span style={{ marginRight: '0.5rem', display: 'inline-block', transform: isEpExpanded ? 'rotate(90deg)' : 'none' }}>▸</span>
                                                  <span>{t('drawer.episode', { season: padNumber(ep.season), episode: padNumber(ep.episode) })}</span>
                                                </td>
                                                <td>
                                                  <span className={`badge ${ep.matched ? 'badge-success' : 'badge-pending'}`}>
                                                    {ep.matched ? t('drawer.episodeMatched') : t('drawer.episodeUnmatched')}
                                                  </span>
                                                </td>
                                              </tr>
                                              {isEpExpanded && (
                                                <tr>
                                                  <td colSpan={2} style={{ padding: 0 }}>
                                                    {ep.occurrences.length === 0 ? (
                                                      <div style={{ padding: '0.75rem 1.25rem', color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem', backgroundColor: 'var(--bg-app)' }}>
                                                        {t('drawer.episodeNoOccurrences')}
                                                      </div>
                                                    ) : (
                                                      <div style={{ padding: '1rem 1.25rem', backgroundColor: 'var(--bg-app)', borderTop: '1px solid var(--border-color)', borderBottom: '1px solid var(--border-color)' }}>
                                                        {renderOccurrences(ep.occurrences)}
                                                      </div>
                                                    )}
                                                  </td>
                                                </tr>
                                              )}
                                            </React.Fragment>
                                          );
                                        })}
                                      </tbody>
                                    </table>
                                  )
                                )}
                              </div>
                            );
                          })}
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </>
            )}
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>

      {!isMobile && (
        <MediaOccurrenceDrawer
          item={detailItem}
          onOpenChange={(open) => !open && setDetailItem(null)}
          contentClassName="drawer-content--secondary"
          withOverlay={false}
          modal={false}
        />
      )}
    </Tabs.Content>
  );
}
