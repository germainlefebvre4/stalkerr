import { useState, FormEvent } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { ScheduleWindow, ScheduleWindowAction, ScheduleWindowInput } from '../types';
import { useApiErrorMessage } from '../hooks/useApiErrorMessage';

interface ScheduleWindowDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  window: ScheduleWindow | null; // null = create mode
  onSubmit: (input: ScheduleWindowInput) => Promise<void>;
}

const DAYS_OF_WEEK = [
  'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday',
];

const ACTIONS: ScheduleWindowAction[] = ['throttle', 'stop'];

const emptyInput = (): ScheduleWindowInput => ({
  days_of_week: [],
  start_time: '08:00',
  end_time: '18:00',
  action: 'throttle',
});

// Create/edit dialog for a weekly bandwidth schedule window. An end time
// earlier than (or equal to) the start time is accepted without a
// validation error and denotes a window spanning past midnight. See
// download-bandwidth-schedule's "Overnight Windows" and
// frontend-bandwidth-schedule-management's "Overnight Window Input".
export function ScheduleWindowDialog({ isOpen, onOpenChange, window: editingWindow, onSubmit }: ScheduleWindowDialogProps) {
  const { t } = useTranslation('settings');
  const translateApiError = useApiErrorMessage();
  const isEdit = editingWindow !== null;

  const [input, setInput] = useState<ScheduleWindowInput>(emptyInput());
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Reset the form whenever the dialog opens (or opens for a different
  // window), rather than in an effect - the dialog stays mounted across
  // opens/closes for Radix's animations, so it never remounts on its own.
  const [formResetKey, setFormResetKey] = useState({ isOpen, editingWindow });
  if (isOpen && (formResetKey.isOpen !== isOpen || formResetKey.editingWindow !== editingWindow)) {
    setFormResetKey({ isOpen, editingWindow });
    setError(null);
    if (editingWindow) {
      setInput({
        days_of_week: editingWindow.days_of_week,
        start_time: editingWindow.start_time,
        end_time: editingWindow.end_time,
        action: editingWindow.action === 'none' ? 'throttle' : editingWindow.action,
      });
    } else {
      setInput(emptyInput());
    }
  } else if (formResetKey.isOpen !== isOpen) {
    setFormResetKey({ isOpen, editingWindow });
  }

  const isOvernight = input.end_time <= input.start_time;

  const toggleDay = (day: string) => {
    setInput(prev => ({
      ...prev,
      days_of_week: prev.days_of_week.includes(day)
        ? prev.days_of_week.filter(d => d !== day)
        : [...prev.days_of_week, day],
    }));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (input.days_of_week.length === 0) {
      setError(t('bandwidthSchedule.daysRequired'));
      return;
    }

    setSaving(true);
    setError(null);
    try {
      await onSubmit(input);
      onOpenChange(false);
    } catch (err) {
      setError(translateApiError(err));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="dialog-overlay" />
        <Dialog.Content className="dialog-content">
          <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)', marginBottom: '1rem' }}>
            {isEdit ? t('bandwidthSchedule.editTitle') : t('bandwidthSchedule.createTitle')}
          </Dialog.Title>
          <Dialog.Description style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginBottom: '1.5rem' }}>
            {t('bandwidthSchedule.dialogDescription')}
          </Dialog.Description>

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <span style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>
                {t('bandwidthSchedule.daysOfWeekLabel')}
              </span>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.75rem' }}>
                {DAYS_OF_WEEK.map(day => (
                  <label key={day} style={{ display: 'flex', alignItems: 'center', gap: '0.35rem', fontSize: '0.85rem' }}>
                    <input
                      type="checkbox"
                      checked={input.days_of_week.includes(day)}
                      onChange={() => toggleDay(day)}
                    />
                    {t(`bandwidthSchedule.days.${day}`)}
                  </label>
                ))}
              </div>
              {input.days_of_week.length === 0 && (
                <p className="settings-field-hint" style={{ color: 'var(--status-failed-text)' }}>
                  {t('bandwidthSchedule.daysRequired')}
                </p>
              )}
            </div>

            <div style={{ display: 'flex', gap: '1rem' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', flex: 1 }}>
                <label htmlFor="schedule-window-start" style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>
                  {t('bandwidthSchedule.startTimeLabel')}
                </label>
                <input
                  id="schedule-window-start"
                  type="time"
                  className="custom-input"
                  value={input.start_time}
                  onChange={e => setInput(prev => ({ ...prev, start_time: e.target.value }))}
                  required
                />
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem', flex: 1 }}>
                <label htmlFor="schedule-window-end" style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>
                  {t('bandwidthSchedule.endTimeLabel')}
                </label>
                <input
                  id="schedule-window-end"
                  type="time"
                  className="custom-input"
                  value={input.end_time}
                  onChange={e => setInput(prev => ({ ...prev, end_time: e.target.value }))}
                  required
                />
              </div>
            </div>

            {isOvernight && (
              <p className="settings-field-hint">{t('bandwidthSchedule.overnightHint')}</p>
            )}

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
              <label htmlFor="schedule-window-action" style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)' }}>
                {t('bandwidthSchedule.actionLabel')}
              </label>
              <select
                id="schedule-window-action"
                className="custom-select"
                value={input.action}
                onChange={e => setInput(prev => ({ ...prev, action: e.target.value as ScheduleWindowAction }))}
              >
                {ACTIONS.map(action => (
                  <option key={action} value={action}>{t(`bandwidthSchedule.actions.${action}`)}</option>
                ))}
              </select>
            </div>

            {error && (
              <div style={{ padding: '0.75rem 1rem', backgroundColor: 'var(--status-failed-bg)', color: 'var(--status-failed-text)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', fontWeight: 600 }}>
                ⚠️ {error}
              </div>
            )}

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
              <button type="button" disabled={saving} onClick={() => onOpenChange(false)} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
                {t('cancel')}
              </button>
              <button type="submit" disabled={saving} className="btn-primary" style={{ padding: '0.5rem 1.25rem' }}>
                {saving ? t('saving') : t('save')}
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
