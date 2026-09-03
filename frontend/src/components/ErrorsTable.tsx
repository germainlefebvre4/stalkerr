import { useTranslation } from 'react-i18next';
import { TFunction } from 'i18next';
import { DownloadEnriched } from '../types';
import { formatDate } from '../utils/date';

interface ErrorsTableProps {
  downloads: DownloadEnriched[];
  loading: boolean;
  onRowClick: (item: DownloadEnriched) => void;
}

export interface ErrorReasonBadge {
  key: 'missing_year' | 'year_mismatch' | 'unknown_format';
  label: string;
}

export function getErrorReasons(item: DownloadEnriched, t: TFunction): ErrorReasonBadge[] {
  const badges: ErrorReasonBadge[] = [];
  if (!item.file_info) return badges;
  if (!item.file_info.has_year_in_path) {
    badges.push({ key: 'missing_year', label: t('badges.missingYear') });
  }
  if (item.file_info.year_mismatch) {
    badges.push({ key: 'year_mismatch', label: t('badges.yearMismatch') });
  }
  if (!item.file_info.is_valid_format) {
    badges.push({ key: 'unknown_format', label: t('badges.unknownFormat') });
  }
  return badges;
}

function buildRow(item: DownloadEnriched, t: TFunction) {
  const title = item.content?.title || item.file_info?.file_name || item.file_info?.folder_name || item.url;
  const year = item.content?.year ? `(${item.content.year})` : '';
  const typeIcon = item.content?.type === 'movies' ? '🎬' : item.content?.type === 'tvshows' ? '📺' : '🔗';
  const reasons = getErrorReasons(item, t);
  const hasYearReason = reasons.some(r => r.key === 'missing_year' || r.key === 'year_mismatch');

  return { item, title, year, typeIcon, reasons, hasYearReason };
}

export function ErrorsTable({ downloads, loading, onRowClick }: ErrorsTableProps) {
  const { t, i18n } = useTranslation('errors');

  const rows = downloads.map(item => buildRow(item, t));

  return (
    <div className="table-flush" style={{ overflowX: 'auto', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)' }}>
      <table className="custom-table">
        <thead>
          <tr>
            <th>{t('table.headers.title')}</th>
            <th>{t('table.headers.reasons')}</th>
            <th>{t('table.headers.year')}</th>
            <th>{t('table.headers.extension')}</th>
            <th>{t('table.headers.location')}</th>
            <th>{t('table.headers.completedAt')}</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={6} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
                <span style={{ fontWeight: 600 }}>{t('loading')}</span>
              </td>
            </tr>
          ) : downloads.length === 0 ? (
            <tr>
              <td colSpan={6} style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>{t('empty')}</td>
            </tr>
          ) : (
            rows.map(({ item, title, year, typeIcon, reasons, hasYearReason }) => (
              <tr key={item.id} className="clickable-row" onClick={() => onRowClick(item)}>
                <td style={{ fontWeight: 600, color: 'var(--primary-slate)' }}>
                  <span style={{ marginRight: '0.5rem' }}>{typeIcon}</span>{title} {year}
                </td>
                <td>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.35rem' }}>
                    {reasons.map(reason => (
                      <span key={reason.key} className="badge badge-failed" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', fontWeight: 600 }}>
                        {reason.label}
                      </span>
                    ))}
                  </div>
                </td>
                <td>
                  {hasYearReason ? (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.15rem', fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                      <span>{t('year.detected', { year: item.file_info?.detected_year ?? t('year.unknown') })}</span>
                      <span>{t('year.expected', { year: item.content?.year ?? t('year.unknown') })}</span>
                    </div>
                  ) : null}
                </td>
                <td style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                  {item.file_info?.extension || t('unknownExtension')}
                </td>
                <td style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                  {item.file_info?.folder_name}
                </td>
                <td style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                  {item.completed_at ? formatDate(item.completed_at, i18n.language) : ''}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
