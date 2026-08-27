import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { MediaGroupItem, PlaylistItem } from '../types';
import { api } from '../services/api';
import { getDateGroupLabel, getDateGroupStarts } from '../utils/date';
import { useIsMobile } from '../hooks/useMediaQuery';
import { Pagination } from './Pagination';
import { PlaylistItemsTable } from './PlaylistItemsTable';

interface PlaylistGroupedViewProps {
  groups: MediaGroupItem[];
  groupsLoading: boolean;
  groupsTotal: number;
  groupsPage: number;
  setGroupsPage: React.Dispatch<React.SetStateAction<number>>;
  groupsLimit: number;
  onOpenOverride: (item: PlaylistItem) => void;
  onResetPipeline: (id: number, contentType: string) => void;
}

const EXPANDED_LIMIT = 10;

function groupKey(group: MediaGroupItem): string {
  if (group.type === 'movie') return `movie:${group.movie_id}`;
  if (group.type === 'tvshow') return `tvshow:${group.tmdb_id}`;
  return group.type;
}

function seasonLabel(group: MediaGroupItem): string | null {
  if (group.season_start == null || group.season_end == null) return null;
  const start = `S${String(group.season_start).padStart(2, '0')}`;
  if (group.season_start === group.season_end) return start;
  return `${start}-S${String(group.season_end).padStart(2, '0')}`;
}

export function PlaylistGroupedView({
  groups,
  groupsLoading,
  groupsTotal,
  groupsPage,
  setGroupsPage,
  groupsLimit,
  onOpenOverride,
  onResetPipeline,
}: PlaylistGroupedViewProps) {
  const { t, i18n } = useTranslation('playlist');
  const isMobile = useIsMobile();

  const [expandedKey, setExpandedKey] = useState<string | null>(null);
  const [expandedItems, setExpandedItems] = useState<PlaylistItem[]>([]);
  const [expandedTotal, setExpandedTotal] = useState(0);
  const [expandedLoading, setExpandedLoading] = useState(false);
  const [expandedPage, setExpandedPage] = useState(1);

  const expandedGroup = groups.find(g => groupKey(g) === expandedKey) || null;

  useEffect(() => {
    if (!expandedGroup) return;
    // Re-fetches on page change; re-fetching on group identity change (not the
    // group object reference, which can churn on every parent refetch) is
    // handled by keying off expandedKey. Deferred a tick (rather than calling
    // setExpandedLoading directly in the effect body) to avoid synchronous
    // cascading renders from an effect.
    void Promise.resolve().then(() => {
      setExpandedLoading(true);
      return api.getGroupItems(expandedGroup, expandedPage, EXPANDED_LIMIT)
        .then(data => {
          setExpandedItems(data.data || []);
          setExpandedTotal(data.total || 0);
        })
        .catch(() => {})
        .finally(() => setExpandedLoading(false));
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [expandedKey, expandedPage]);

  const toggleExpand = (group: MediaGroupItem) => {
    const key = groupKey(group);
    if (expandedKey === key) {
      setExpandedKey(null);
      return;
    }
    setExpandedKey(key);
    setExpandedPage(1);
  };

  const titleFor = (group: MediaGroupItem): string => {
    if (group.type === 'unmatched_movies') return t('grouped.unmatchedMovies');
    if (group.type === 'unmatched_tvshows') return t('grouped.unmatchedTVShows');
    return group.title || '';
  };

  const dateGroupStarts = getDateGroupStarts(groups, g => g.latest_activity);

  const renderExpandedSection = () => (
    <div style={{ padding: '1rem', backgroundColor: 'var(--bg-app)', borderTop: '1px solid var(--border-color)', borderBottom: '1px solid var(--border-color)' }}>
      <PlaylistItemsTable
        items={expandedItems}
        loading={expandedLoading}
        showDateGroups={false}
        onOpenOverride={onOpenOverride}
        onResetPipeline={onResetPipeline}
      />
      <Pagination
        total={expandedTotal}
        page={expandedPage}
        setPage={setExpandedPage}
        limit={EXPANDED_LIMIT}
      />
    </div>
  );

  if (isMobile) {
    return (
      <div>
        {groupsLoading ? (
          <div className="mobile-list-empty">{t('grouped.table.loading')}</div>
        ) : groups.length === 0 ? (
          <div className="mobile-list-empty">{t('grouped.table.empty')}</div>
        ) : (
          groups.map((group, index) => {
            const key = groupKey(group);
            const isExpanded = expandedKey === key;
            const season = seasonLabel(group);
            return (
              <React.Fragment key={key}>
                {dateGroupStarts[index] && (
                  <div className="date-group-header">
                    {getDateGroupLabel(new Date(group.latest_activity), t, i18n.language)}
                  </div>
                )}
                <div className="mobile-list-card" onClick={() => toggleExpand(group)}>
                  <div className="mobile-list-card-main">
                    <span className="mobile-list-card-title">{titleFor(group)}</span>
                    <span className="mobile-list-card-subtitle">
                      {group.year ?? ''}{season ? ` · ${season}` : ''}
                    </span>
                  </div>
                  <span style={{ transform: isExpanded ? 'rotate(90deg)' : 'none', display: 'inline-block' }}>▸</span>
                </div>
                {isExpanded && renderExpandedSection()}
              </React.Fragment>
            );
          })
        )}
        <Pagination total={groupsTotal} page={groupsPage} setPage={setGroupsPage} limit={groupsLimit} />
      </div>
    );
  }

  return (
    <div>
      <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
        <table className="custom-table">
          <thead>
            <tr>
              <th>{t('grouped.table.headers.title')}</th>
              <th>{t('grouped.table.headers.year')}</th>
              <th>{t('grouped.table.headers.season')}</th>
            </tr>
          </thead>
          <tbody>
            {groupsLoading ? (
              <tr>
                <td colSpan={3} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
                  <span style={{ fontWeight: 600 }}>{t('grouped.table.loading')}</span>
                </td>
              </tr>
            ) : groups.length === 0 ? (
              <tr>
                <td colSpan={3} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('grouped.table.empty')}</td>
              </tr>
            ) : (
              groups.map((group, index) => {
                const key = groupKey(group);
                const isExpanded = expandedKey === key;
                const season = seasonLabel(group);
                return (
                  <React.Fragment key={key}>
                    {dateGroupStarts[index] && (
                      <tr>
                        <td colSpan={3} className="date-group-header">
                          {getDateGroupLabel(new Date(group.latest_activity), t, i18n.language)}
                        </td>
                      </tr>
                    )}
                    <tr className="clickable-row" onClick={() => toggleExpand(group)}>
                      <td style={{ fontWeight: 600, color: 'var(--primary-slate)' }}>
                        <span style={{ marginRight: '0.5rem', display: 'inline-block', transform: isExpanded ? 'rotate(90deg)' : 'none' }}>▸</span>
                        {titleFor(group)}
                      </td>
                      <td style={{ color: 'var(--text-secondary)' }}>{group.year ?? '—'}</td>
                      <td style={{ color: 'var(--text-secondary)' }}>{season ?? '—'}</td>
                    </tr>
                    {isExpanded && (
                      <tr>
                        <td colSpan={3} style={{ padding: 0 }}>
                          {renderExpandedSection()}
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })
            )}
          </tbody>
        </table>
      </div>
      <Pagination total={groupsTotal} page={groupsPage} setPage={setGroupsPage} limit={groupsLimit} />
    </div>
  );
}
