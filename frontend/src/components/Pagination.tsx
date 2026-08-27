import React from 'react';
import { useTranslation } from 'react-i18next';
import { useIsMobile } from '../hooks/useMediaQuery';

interface PaginationProps {
  total: number;
  page: number;
  setPage: React.Dispatch<React.SetStateAction<number>>;
  limit: number;
  setLimit?: (limit: number) => void;
  limitOptions?: number[];
}

export function Pagination({ total, page, setPage, limit, setLimit, limitOptions }: PaginationProps) {
  const { t } = useTranslation('playlist');
  const isMobile = useIsMobile();
  const [gotoPageInput, setGotoPageInput] = React.useState('');

  const totalPages = Math.ceil(total / limit);

  const handleGotoPage = () => {
    const parsed = parseInt(gotoPageInput, 10);
    if (!isNaN(parsed)) {
      setPage(Math.min(Math.max(1, parsed), totalPages));
    }
    setGotoPageInput('');
  };

  if (total === 0) return null;

  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
        <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
          {t('pagination.showing', {
            from: total === 0 ? 0 : (page - 1) * limit + 1,
            to: Math.min(page * limit, total),
            total,
          })}
        </span>

        {limitOptions && setLimit && (
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>{t('pagination.show')}</span>
            <select
              value={limit}
              onChange={e => setLimit(parseInt(e.target.value, 10))}
              className="custom-select"
              style={{ padding: '0.2rem 1.5rem 0.2rem 0.5rem', fontSize: '0.8rem', width: 'auto', height: 'auto' }}
            >
              {limitOptions.map(option => (
                <option key={option} value={option}>{option}</option>
              ))}
            </select>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>{t('pagination.perPage')}</span>
          </div>
        )}
      </div>

      {totalPages > 1 && (
        <div style={{ display: 'flex', gap: '0.35rem', alignItems: 'center' }}>
          <button
            disabled={page === 1}
            onClick={() => setPage(1)}
            className="btn-secondary"
            style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: page === 1 ? 0.5 : 1 }}
            title={t('pagination.firstPage')}
          >
            &lt;&lt;
          </button>
          <button
            disabled={page === 1}
            onClick={() => setPage(p => Math.max(1, p - 1))}
            className="btn-secondary"
            style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: page === 1 ? 0.5 : 1 }}
            title={t('pagination.prevPage')}
          >
            &lt;
          </button>

          {isMobile ? (
            <span style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
              {t('pagination.page', { current: page, total: totalPages })}
            </span>
          ) : (
            getPaginationRange(page, totalPages).map((p, index) => {
              if (p === '...') {
                return (
                  <span key={`ellipsis-${index}`} style={{ padding: '0.4rem 0.6rem', color: 'var(--text-muted)', fontWeight: 'bold' }}>
                    ...
                  </span>
                );
              }
              return (
                <button
                  key={`page-${p}`}
                  onClick={() => setPage(p as number)}
                  className={page === p ? 'btn-primary' : 'btn-secondary'}
                  style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem' }}
                >
                  {p}
                </button>
              );
            })
          )}

          <button
            disabled={page === totalPages}
            onClick={() => setPage(p => Math.min(totalPages, p + 1))}
            className="btn-secondary"
            style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: page === totalPages ? 0.5 : 1 }}
            title={t('pagination.nextPage')}
          >
            &gt;
          </button>
          <button
            disabled={page === totalPages}
            onClick={() => setPage(totalPages)}
            className="btn-secondary"
            style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem', opacity: page === totalPages ? 0.5 : 1 }}
            title={t('pagination.lastPage')}
          >
            &gt;&gt;
          </button>

          {!isMobile && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.3rem', marginLeft: '0.4rem' }}>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>{t('pagination.goTo')}</span>
              <input
                type="number"
                min={1}
                max={totalPages}
                value={gotoPageInput}
                onChange={e => setGotoPageInput(e.target.value)}
                onKeyDown={e => { if (e.key === 'Enter') handleGotoPage(); }}
                placeholder={String(page)}
                className="custom-input"
                style={{ width: '4rem', padding: '0.35rem 0.5rem', fontSize: '0.8rem' }}
              />
              <button
                onClick={handleGotoPage}
                className="btn-secondary"
                style={{ padding: '0.4rem 0.6rem', fontSize: '0.8rem' }}
              >
                {t('pagination.go')}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function getPaginationRange(current: number, total: number): (number | string)[] {
  const range: number[] = [];
  const delta = 1;

  for (let i = 1; i <= total; i++) {
    if (i === 1 || i === total || (i >= current - delta && i <= current + delta)) {
      range.push(i);
    }
  }

  const result: (number | string)[] = [];
  let prev: number | null = null;
  for (const i of range) {
    if (prev !== null) {
      if (i - prev === 2) {
        result.push(prev + 1);
      } else if (i - prev > 2) {
        result.push('...');
      }
    }
    result.push(i);
    prev = i;
  }
  return result;
}
