import React from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { RadarrMovieListItem, SonarrSeriesListItem, RadarrMovieMatchesResponse, SonarrSeriesEpisodesResponse } from '../types';
import { useIsMobile } from '../hooks/useMediaQuery';
import { getPipelineStateBadgeClass } from '../utils/pipelineState';
import { Pagination } from './Pagination';
import { api, ApiError } from '../services/api';

interface RadarrSonarrTabProps {
  filmsItems: RadarrMovieListItem[];
  filmsLoading: boolean;
  filmsError: string | null;
  filmsTotal: number;
  filmsPage: number;
  setFilmsPage: React.Dispatch<React.SetStateAction<number>>;
  filmsLimit: number;
  fetchFilms: () => void;

  seriesItems: SonarrSeriesListItem[];
  seriesLoading: boolean;
  seriesError: string | null;
  seriesTotal: number;
  seriesPage: number;
  setSeriesPage: React.Dispatch<React.SetStateAction<number>>;
  seriesLimit: number;
  fetchSeries: () => void;
}

function seriesBadgeClass(matched: number, monitored: number): string {
  if (monitored === 0 || matched === 0) return 'badge-pending';
  if (matched === monitored) return 'badge-success';
  return 'badge-progress';
}

export function RadarrSonarrTab({
  filmsItems, filmsLoading, filmsError, filmsTotal, filmsPage, setFilmsPage, filmsLimit, fetchFilms,
  seriesItems, seriesLoading, seriesError, seriesTotal, seriesPage, setSeriesPage, seriesLimit, fetchSeries,
}: RadarrSonarrTabProps) {
  const { t } = useTranslation('radarrSonarr');
  const { t: tCommon } = useTranslation('common');
  const isMobile = useIsMobile();

  const [selectedMovie, setSelectedMovie] = React.useState<RadarrMovieListItem | null>(null);
  const [selectedSeries, setSelectedSeries] = React.useState<SonarrSeriesListItem | null>(null);
  const [movieDetail, setMovieDetail] = React.useState<RadarrMovieMatchesResponse | null>(null);
  const [seriesDetail, setSeriesDetail] = React.useState<SonarrSeriesEpisodesResponse | null>(null);
  const [detailLoading, setDetailLoading] = React.useState(false);
  const [detailError, setDetailError] = React.useState<string | null>(null);

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
  };
  const openSeries = (item: SonarrSeriesListItem) => {
    setSelectedMovie(null);
    setSelectedSeries(item);
  };
  const closeDrawer = () => {
    setSelectedMovie(null);
    setSelectedSeries(null);
    setMovieDetail(null);
    setSeriesDetail(null);
  };

  return (
    <Tabs.Content value="radarr-sonarr" className="card tab-panel">
      <div style={{ display: 'flex', flexDirection: 'column', gap: '2.5rem' }}>
        {/* Section Films (Radarr) */}
        <section>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
            <div>
              <h2 style={{ fontSize: '1.15rem', fontWeight: 800, color: 'var(--primary-slate)' }}>{t('films.heading')}</h2>
              <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('films.subtitle')}</p>
            </div>
            <button onClick={fetchFilms} className="btn-secondary">{t('films.refresh')}</button>
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
                <div className="mobile-list-empty">{t('films.empty')}</div>
              ) : (
                filmsItems.map(item => (
                  <div key={item.radarr_id} className="mobile-list-card" onClick={() => openMovie(item)}>
                    <div className="mobile-list-card-main">
                      <span className="mobile-list-card-title">{item.title}</span>
                      <span className="mobile-list-card-subtitle">{item.year}</span>
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
                  </tr>
                </thead>
                <tbody>
                  {filmsLoading && filmsItems.length === 0 ? (
                    <tr><td colSpan={3} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('films.loading')}</td></tr>
                  ) : filmsItems.length === 0 ? (
                    <tr><td colSpan={3} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('films.empty')}</td></tr>
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

        {/* Section Séries (Sonarr) */}
        <section>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
            <div>
              <h2 style={{ fontSize: '1.15rem', fontWeight: 800, color: 'var(--primary-slate)' }}>{t('series.heading')}</h2>
              <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('series.subtitle')}</p>
            </div>
            <button onClick={fetchSeries} className="btn-secondary">{t('series.refresh')}</button>
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
                <div className="mobile-list-empty">{t('series.empty')}</div>
              ) : (
                seriesItems.map(item => (
                  <div key={item.sonarr_id} className="mobile-list-card" onClick={() => openSeries(item)}>
                    <div className="mobile-list-card-main">
                      <span className="mobile-list-card-title">{item.title}</span>
                      <span className="mobile-list-card-subtitle">{item.year}</span>
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
                  </tr>
                </thead>
                <tbody>
                  {seriesLoading && seriesItems.length === 0 ? (
                    <tr><td colSpan={3} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('series.loading')}</td></tr>
                  ) : seriesItems.length === 0 ? (
                    <tr><td colSpan={3} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('series.empty')}</td></tr>
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
      </div>

      {/* Sidepanel de Détails (Film ou Série) */}
      <Dialog.Root open={!!selectedMovie || !!selectedSeries} onOpenChange={(open) => !open && closeDrawer()}>
        <Dialog.Portal>
          <Dialog.Overlay className="drawer-overlay" />
          <Dialog.Content className="drawer-content">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1rem' }}>
              <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)', margin: 0 }}>
                {selectedMovie ? t('drawer.movieTitle') : t('drawer.seriesTitle')}
              </Dialog.Title>
              <Dialog.Close className="btn-secondary" style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem', borderRadius: 'var(--radius-sm)' }}>
                {t('drawer.close')}
              </Dialog.Close>
            </div>
            <Dialog.Description style={{ display: 'none' }}>{t('drawer.description')}</Dialog.Description>

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
                    ) : (
                      <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
                        <table className="custom-table">
                          <thead>
                            <tr>
                              <th>{t('drawer.occurrenceResolution')}</th>
                              <th>{t('drawer.occurrenceState')}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {movieDetail.occurrences.map(occ => (
                              <tr key={occ.id}>
                                <td>{occ.resolution || t('drawer.unknownResolution')}</td>
                                <td><span className={`badge ${getPipelineStateBadgeClass(occ.state)}`}>{occ.state}</span></td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
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
                  {seriesDetail.episodes.length === 0 ? (
                    <div style={{ color: 'var(--text-muted)', fontStyle: 'italic', fontSize: '0.85rem' }}>{t('series.empty')}</div>
                  ) : (
                    <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
                      <table className="custom-table">
                        <thead>
                          <tr>
                            <th></th>
                            <th>{t('drawer.occurrenceState')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {seriesDetail.episodes.map(ep => (
                            <tr key={`${ep.season}-${ep.episode}`}>
                              <td style={{ fontWeight: 600 }}>{t('drawer.episode', { season: ep.season, episode: ep.episode })}</td>
                              <td>
                                <span className={`badge ${ep.matched ? 'badge-success' : 'badge-pending'}`}>
                                  {ep.matched ? t('drawer.episodeMatched') : t('drawer.episodeUnmatched')}
                                </span>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              </div>
            )}
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </Tabs.Content>
  );
}
