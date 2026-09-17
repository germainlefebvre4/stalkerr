import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { BandwidthScheduleSection } from './BandwidthScheduleSection';
import { EffectivePolicy, ScheduleWindow, SettingsField } from '../types';

const throttleSettings: SettingsField[] = [
  { key: 'downloads.throttle_rate_kbps', value: 0, sensitive: false, origin: 'config', restart_required: false },
  { key: 'jellyfin.playback_check_enabled', value: false, sensitive: false, origin: 'config', restart_required: false },
  { key: 'jellyfin.playback_action', value: 'throttle', sensitive: false, origin: 'config', restart_required: false },
  { key: 'jellyfin.playback_poll_interval_seconds', value: 20, sensitive: false, origin: 'config', restart_required: false },
];

const throttleWindow: ScheduleWindow = {
  id: 1, days_of_week: ['monday', 'tuesday'], start_time: '08:00', end_time: '18:00', action: 'throttle',
};

const overnightWindow: ScheduleWindow = {
  id: 2, days_of_week: ['friday'], start_time: '22:00', end_time: '07:00', action: 'stop',
};

function renderSection(opts: {
  windows?: ScheduleWindow[];
  effectivePolicy?: EffectivePolicy | null;
  settings?: SettingsField[];
} = {}) {
  const onSetSetting = vi.fn().mockResolvedValue(undefined);
  const onClearSetting = vi.fn().mockResolvedValue(undefined);
  const onCreateWindow = vi.fn().mockResolvedValue(undefined);
  const onUpdateWindow = vi.fn().mockResolvedValue(undefined);
  const onDeleteWindow = vi.fn().mockResolvedValue(undefined);
  const onFetchEffectivePolicy = vi.fn();

  const utils = render(
    <I18nextProvider i18n={i18n}>
      <BandwidthScheduleSection
        isExpanded
        settings={opts.settings ?? throttleSettings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
        windows={opts.windows ?? []}
        windowsLoading={false}
        onCreateWindow={onCreateWindow}
        onUpdateWindow={onUpdateWindow}
        onDeleteWindow={onDeleteWindow}
        effectivePolicy={opts.effectivePolicy ?? null}
        onFetchEffectivePolicy={onFetchEffectivePolicy}
      />
    </I18nextProvider>
  );

  return { ...utils, onSetSetting, onClearSetting, onCreateWindow, onUpdateWindow, onDeleteWindow, onFetchEffectivePolicy };
}

describe('BandwidthScheduleSection', () => {
  afterEach(() => cleanup());

  it('creates a new schedule window', async () => {
    const { onCreateWindow } = renderSection();

    fireEvent.click(screen.getByText('+ Add window'));
    fireEvent.click(screen.getByLabelText('Monday'));
    fireEvent.change(screen.getByLabelText('Start time'), { target: { value: '09:00' } });
    fireEvent.change(screen.getByLabelText('End time'), { target: { value: '17:00' } });
    fireEvent.change(screen.getByLabelText('Action'), { target: { value: 'stop' } });
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onCreateWindow).toHaveBeenCalled());
    expect(onCreateWindow.mock.calls[0][0]).toEqual({
      days_of_week: ['monday'], start_time: '09:00', end_time: '17:00', action: 'stop',
    });
  });

  it('accepts an overnight window (end time earlier than start time) without a validation error', async () => {
    const { onCreateWindow } = renderSection();

    fireEvent.click(screen.getByText('+ Add window'));
    fireEvent.click(screen.getByLabelText('Friday'));
    fireEvent.change(screen.getByLabelText('Start time'), { target: { value: '22:00' } });
    fireEvent.change(screen.getByLabelText('End time'), { target: { value: '07:00' } });
    expect(screen.getByText(/spans past midnight/)).toBeInTheDocument();

    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onCreateWindow).toHaveBeenCalled());
    expect(onCreateWindow.mock.calls[0][0]).toEqual({
      days_of_week: ['friday'], start_time: '22:00', end_time: '07:00', action: 'throttle',
    });
  });

  it('displays a badge on an existing window spanning past midnight', () => {
    renderSection({ windows: [overnightWindow] });
    expect(screen.getByText('overnight')).toBeInTheDocument();
  });

  it('editing a window calls onUpdateWindow with its id', async () => {
    const { onUpdateWindow } = renderSection({ windows: [throttleWindow] });

    fireEvent.click(screen.getByText('Edit'));
    fireEvent.change(screen.getByLabelText('Action'), { target: { value: 'stop' } });
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onUpdateWindow).toHaveBeenCalled());
    expect(onUpdateWindow.mock.calls[0][0]).toBe(1);
    expect(onUpdateWindow.mock.calls[0][1].action).toBe('stop');
  });

  it('deleting a window calls onDeleteWindow with its id after confirmation', () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    const { onDeleteWindow } = renderSection({ windows: [throttleWindow] });

    fireEvent.click(screen.getByText('Delete'));

    expect(onDeleteWindow).toHaveBeenCalledWith(1);
    confirmSpy.mockRestore();
  });

  it('leaves other windows unchanged after deleting one', () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    const { onDeleteWindow } = renderSection({ windows: [throttleWindow, overnightWindow] });

    fireEvent.click(screen.getAllByText('Delete')[0]);

    expect(onDeleteWindow).toHaveBeenCalledWith(1);
    expect(onDeleteWindow).not.toHaveBeenCalledWith(2);
    confirmSpy.mockRestore();
  });

  it('renders each Jellyfin throttle control and edits the shared throttle rate field', async () => {
    const { onSetSetting } = renderSection();

    expect(screen.getByText('Shared throttle rate (Kbps)')).toBeInTheDocument();
    expect(screen.getByLabelText('Enable Jellyfin playback detection')).toBeInTheDocument();
    expect(screen.getByLabelText('Action while Jellyfin is playing')).toBeInTheDocument();
    expect(screen.getByText('Jellyfin poll interval (seconds)')).toBeInTheDocument();

    fireEvent.change(screen.getByDisplayValue('0'), { target: { value: '2048' } });
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onSetSetting).toHaveBeenCalledWith('downloads.throttle_rate_kbps', 2048));
  });

  it('toggles the Jellyfin playback-check-enabled checkbox', async () => {
    const { onSetSetting } = renderSection();

    fireEvent.click(screen.getByLabelText('Enable Jellyfin playback detection'));
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onSetSetting).toHaveBeenCalledWith('jellyfin.playback_check_enabled', true));
  });

  it('changes the Jellyfin playback action select between throttle and stop', async () => {
    const { onSetSetting } = renderSection();

    fireEvent.change(screen.getByLabelText('Action while Jellyfin is playing'), { target: { value: 'stop' } });
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onSetSetting).toHaveBeenCalledWith('jellyfin.playback_action', 'stop'));
  });

  it('edits the Jellyfin poll interval field', async () => {
    const { onSetSetting } = renderSection();

    fireEvent.change(screen.getByDisplayValue('20'), { target: { value: '45' } });
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onSetSetting).toHaveBeenCalledWith('jellyfin.playback_poll_interval_seconds', 45));
  });

  it('shows the effective policy preview and which signal contributes', () => {
    const { container } = renderSection({
      effectivePolicy: { action: 'throttle', schedule_action: 'throttle', jellyfin_action: 'none' },
    });

    expect(screen.getByText('Effective policy:')).toBeInTheDocument();
    expect(container.querySelector('.badge.badge-progress')).toHaveTextContent('Throttle');
    expect(screen.getByText('schedule')).toBeInTheDocument();
  });

  it('fetches the effective policy on mount and refreshes it periodically without a page reload', () => {
    vi.useFakeTimers();
    try {
      const { onFetchEffectivePolicy } = renderSection();
      expect(onFetchEffectivePolicy).toHaveBeenCalledTimes(1);

      vi.advanceTimersByTime(15000);
      expect(onFetchEffectivePolicy).toHaveBeenCalledTimes(2);

      vi.advanceTimersByTime(15000);
      expect(onFetchEffectivePolicy).toHaveBeenCalledTimes(3);
    } finally {
      vi.useRealTimers();
    }
  });
});
