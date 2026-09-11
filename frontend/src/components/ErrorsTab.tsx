import React, { useEffect, useState } from 'react';
import * as Tabs from '@radix-ui/react-tabs';
import { useTranslation } from 'react-i18next';
import { DownloadEnriched } from '../types';
import { ErrorsTable } from './ErrorsTable';
import { ErrorsSidepanel } from './ErrorsSidepanel';
import { Pagination } from './Pagination';

interface ErrorsTabProps {
  errors: DownloadEnriched[];
  errorsLoading: boolean;
  reasonFilter: string;
  setReasonFilter: (reason: string) => void;
  errorsTotal: number;
  errorsPage: number;
  setErrorsPage: React.Dispatch<React.SetStateAction<number>>;
  errorsLimit: number;
  setErrorsLimit: (limit: number) => void;
  onFetchErrors: () => void;
}

export function ErrorsTab({
  errors,
  errorsLoading,
  reasonFilter,
  setReasonFilter,
  errorsTotal,
  errorsPage,
  setErrorsPage,
  errorsLimit,
  setErrorsLimit,
  onFetchErrors,
}: ErrorsTabProps) {
  const { t } = useTranslation('errors');
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const selectedItem = errors.find(d => d.id === selectedId) ?? null;

  useEffect(() => {
    if (selectedId !== null && !selectedItem) {
      setSelectedId(null);
    }
  }, [selectedId, selectedItem]);

  return (
    <Tabs.Content value="errors" className="card tab-panel">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)' }}>{t('heading')}</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem', fontWeight: 500 }}>{t('subtitle')}</p>
        </div>
        <button onClick={onFetchErrors} className="btn-secondary" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>{t('refresh')}</button>
      </div>

      <div className="filters" style={{ display: 'flex', gap: '0.75rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
        <select
          value={reasonFilter}
          onChange={e => setReasonFilter(e.target.value)}
          style={{ padding: '0.5rem 1rem', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', fontSize: '0.875rem', background: 'var(--bg-card)', fontWeight: 600, color: 'var(--text-secondary)' }}
        >
          <option value="">{t('reasonFilter.all')}</option>
          <option value="unknown_format">{t('reasonFilter.unknownFormat')}</option>
          <option value="missing_year">{t('reasonFilter.missingYear')}</option>
          <option value="year_mismatch">{t('reasonFilter.yearMismatch')}</option>
        </select>
      </div>

      <ErrorsTable
        downloads={errors}
        loading={errorsLoading}
        onRowClick={item => setSelectedId(item.id)}
      />

      <Pagination
        total={errorsTotal}
        page={errorsPage}
        setPage={setErrorsPage}
        limit={errorsLimit}
        setLimit={setErrorsLimit}
        limitOptions={[10, 50, 100]}
      />

      <ErrorsSidepanel item={selectedItem} onOpenChange={(open) => !open && setSelectedId(null)} />
    </Tabs.Content>
  );
}
