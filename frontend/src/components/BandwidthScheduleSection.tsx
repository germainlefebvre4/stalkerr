import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { EffectivePolicy, ScheduleWindow, ScheduleWindowInput, SettingsField } from '../types';
import { SettingsGroupCard } from './SettingsGroupCard';
import { BANDWIDTH_THROTTLE_FIELDS } from '../settingsFieldGroups';
import { ScheduleWindowDialog } from './ScheduleWindowDialog';

interface BandwidthScheduleSectionProps {
  isExpanded: boolean;
  settings: SettingsField[];
  onSetSetting: (key: string, value: unknown) => Promise<unknown>;
  onClearSetting: (key: string) => Promise<unknown>;
  windows: ScheduleWindow[];
  windowsLoading: boolean;
  onCreateWindow: (input: ScheduleWindowInput) => Promise<void>;
  onUpdateWindow: (id: number, input: ScheduleWindowInput) => Promise<void>;
  onDeleteWindow: (id: number) => Promise<void>;
  effectivePolicy: EffectivePolicy | null;
  onFetchEffectivePolicy: () => void;
  searchQuery?: string;
}

// Refresh interval for the effective-policy preview: the underlying policy
// can change purely from time passing (a schedule boundary) or from
// Jellyfin's playback state changing, neither of which is a user action on
// this page. See frontend-bandwidth-schedule-management's "Effective Policy
// Preview".
const EFFECTIVE_POLICY_REFRESH_MS = 15000;

const POLICY_BADGE_CLASS: Record<string, string> = {
  none: 'badge-success',
  throttle: 'badge-progress',
  stop: 'badge-failed',
};

// Configuration-page "Avancé" tab section grouping the weekly bandwidth
// schedule editor, the Jellyfin-throttle settings, and the effective-policy
// preview. See frontend-bandwidth-schedule-management.
export function BandwidthScheduleSection({
  isExpanded, settings, onSetSetting, onClearSetting,
  windows, windowsLoading, onCreateWindow, onUpdateWindow, onDeleteWindow,
  effectivePolicy, onFetchEffectivePolicy, searchQuery,
}: BandwidthScheduleSectionProps) {
  const { t } = useTranslation('settings');
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingWindow, setEditingWindow] = useState<ScheduleWindow | null>(null);

  useEffect(() => {
    if (!isExpanded) return;
    onFetchEffectivePolicy();
    const interval = setInterval(onFetchEffectivePolicy, EFFECTIVE_POLICY_REFRESH_MS);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isExpanded]);

  if (!isExpanded) return null;

  const query = (searchQuery ?? '').trim().toLowerCase();
  const matchesQuery = (label: string) => !query || label.toLowerCase().includes(query);

  const openCreate = () => {
    setEditingWindow(null);
    setIsDialogOpen(true);
  };

  const openEdit = (w: ScheduleWindow) => {
    setEditingWindow(w);
    setIsDialogOpen(true);
  };

  const handleDelete = (w: ScheduleWindow) => {
    if (!confirm(t('bandwidthSchedule.confirmDelete'))) return;
    onDeleteWindow(w.id);
  };

  const formatDays = (days: string[]) => days.map(d => t(`bandwidthSchedule.days.${d}`)).join(', ');

  return (
    <div>
      {matchesQuery(t('bandwidthSchedule.throttleSettingsTitle')) && (
        <div className="settings-cards-grid" style={{ marginBottom: '1.5rem' }}>
          <SettingsGroupCard
            title={t('bandwidthSchedule.throttleSettingsTitle')}
            fields={BANDWIDTH_THROTTLE_FIELDS.map(f => ({ ...f, label: t(`fields.bandwidthThrottle.${f.label}`) }))}
            settings={settings}
            onSetSetting={onSetSetting}
            onClearSetting={onClearSetting}
            searchQuery={searchQuery}
          />
        </div>
      )}

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div>
          <h4 style={{ fontSize: '0.95rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{t('bandwidthSchedule.heading')}</h4>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', marginTop: '0.25rem' }}>{t('bandwidthSchedule.subtitle')}</p>
        </div>
        <button onClick={openCreate} className="btn-primary">{t('bandwidthSchedule.createButton')}</button>
      </div>

      {effectivePolicy && (
        <div className="filter-card" style={{ marginBottom: '1rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span style={{ fontWeight: 700 }}>{t('bandwidthSchedule.effectivePolicyLabel')}:</span>
            <span className={`badge ${POLICY_BADGE_CLASS[effectivePolicy.action] ?? 'badge-muted'}`}>
              {t(`bandwidthSchedule.actions.${effectivePolicy.action}`)}
            </span>
          </div>
          {effectivePolicy.action !== 'none' && (
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', marginTop: '0.4rem' }}>
              {[
                effectivePolicy.schedule_action !== 'none' ? t('bandwidthSchedule.contributingSchedule') : null,
                effectivePolicy.jellyfin_action !== 'none' ? t('bandwidthSchedule.contributingJellyfin') : null,
              ].filter(Boolean).join(', ')}
            </p>
          )}
        </div>
      )}

      {windowsLoading && windows.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-secondary)' }}>{t('loading')}</div>
      ) : windows.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '2rem', border: '1px dashed var(--border-color)', borderRadius: 'var(--radius-md)' }}>
          <p style={{ color: 'var(--text-secondary)', fontWeight: 600 }}>{t('bandwidthSchedule.emptyTitle')}</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {windows.map(w => (
            <div key={w.id} className="filter-card">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <h3 style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--primary-slate)' }}>{formatDays(w.days_of_week)}</h3>
                  <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.25rem' }}>
                    {w.start_time} → {w.end_time}
                    {w.end_time <= w.start_time && (
                      <span className="badge badge-progress" style={{ marginLeft: '0.5rem', fontSize: '0.65rem' }}>
                        {t('bandwidthSchedule.overnightBadge')}
                      </span>
                    )}
                  </p>
                  <span className={`badge ${POLICY_BADGE_CLASS[w.action] ?? 'badge-muted'}`} style={{ marginTop: '0.4rem', fontSize: '0.65rem' }}>
                    {t(`bandwidthSchedule.actions.${w.action}`)}
                  </span>
                </div>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button onClick={() => openEdit(w)} className="btn-secondary" style={{ padding: '0.3rem 0.6rem' }}>
                    {t('bandwidthSchedule.edit')}
                  </button>
                  <button onClick={() => handleDelete(w)} className="btn-danger" style={{ padding: '0.3rem 0.6rem' }}>
                    {t('bandwidthSchedule.delete')}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <ScheduleWindowDialog
        isOpen={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        window={editingWindow}
        onSubmit={async input => {
          if (editingWindow) {
            await onUpdateWindow(editingWindow.id, input);
          } else {
            await onCreateWindow(input);
          }
        }}
      />
    </div>
  );
}
