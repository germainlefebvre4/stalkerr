import { useEffect, useState, useCallback } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { api } from '../services/api';
import { ProcessingLog, PlaylistItem } from '../types';
import { formatDate } from '../utils/date';
import { PlaylistItemsTable } from './PlaylistItemsTable';
import { Pagination } from './Pagination';
import { getLogStatusBadgeClass } from '../utils/logState';

const RUN_ITEMS_LIMIT = 10;

interface RunItemsDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  log: ProcessingLog | null;
  onOpenOverride: (item: PlaylistItem, onSuccess?: () => void) => void;
  onResetPipeline: (id: number, contentType: string, onDone?: () => void) => void;
}

export function RunItemsDialog({
  isOpen,
  onOpenChange,
  log,
  onOpenOverride,
  onResetPipeline,
}: RunItemsDialogProps) {
  const { t, i18n } = useTranslation('logs');
  const [items, setItems] = useState<PlaylistItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);

  useEffect(() => {
    if (!isOpen) return;
    void Promise.resolve().then(() => setPage(1));
  }, [isOpen, log?.id]);

  const refetch = useCallback(() => {
    if (!log) return;
    setLoading(true);
    api.getRunItems(log.id, page, RUN_ITEMS_LIMIT)
      .then(data => {
        setItems(data.data || []);
        setTotal(data.total || 0);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [log, page]);

  useEffect(() => {
    if (!isOpen || !log) return;
    void Promise.resolve().then(() => refetch());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, log?.id, page]);

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content className="dialog-content" style={{ maxWidth: '90%', display: 'flex', flexDirection: 'column', maxHeight: '85vh' }}>
          <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--primary-slate)', marginBottom: '0.5rem' }}>
            {t('runItemsDialog.title')}
          </Dialog.Title>
          <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', lineHeight: 1.4, marginBottom: '1rem', fontWeight: 500 }}>
            {t('runItemsDialog.description')}
          </Dialog.Description>

          {log && (
            <div style={{ backgroundColor: 'var(--bg-app)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-color)', marginBottom: '1rem', display: 'flex', flexWrap: 'wrap', gap: '1rem', fontSize: '0.8rem', alignItems: 'center' }}>
              <div><strong>{t('table.action')}:</strong> <span style={{ color: 'var(--text-secondary)' }}>{log.action}</span></div>
              <span className={`badge ${getLogStatusBadgeClass(log.status)}`}>
                {log.status === 'in_progress' ? t('status.inProgress') : log.status === 'success' ? t('status.success') : t('status.failed')}
              </span>
              <div><strong>{t('table.startedAt')}:</strong> <span style={{ color: 'var(--text-secondary)' }}>{formatDate(log.started_at, i18n.language)}</span></div>
              <div><strong>{t('table.itemCount')}:</strong> <span style={{ color: 'var(--text-secondary)' }}>{log.item_count}</span></div>
            </div>
          )}

          <div style={{ flex: 1, overflowY: 'auto' }}>
            <PlaylistItemsTable
              items={items}
              loading={loading}
              showDateGroups={false}
              onOpenOverride={(item) => onOpenOverride(item, refetch)}
              onResetPipeline={(id, contentType) => onResetPipeline(id, contentType, refetch)}
            />
          </div>
          <Pagination total={total} page={page} setPage={setPage} limit={RUN_ITEMS_LIMIT} />

          <div style={{ display: 'flex', justifyContent: 'flex-end', paddingTop: '1rem' }}>
            <Dialog.Close className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
              {t('playlist:drawer.close')}
            </Dialog.Close>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
