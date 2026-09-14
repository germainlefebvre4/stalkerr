import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { SettingsGroupCard } from './SettingsGroupCard';
import { SettingsField } from '../types';

const settings: SettingsField[] = [
  { key: 'radarr.url', value: 'http://old.example.com', sensitive: false, origin: 'config', restart_required: false },
  { key: 'radarr.enabled', value: false, sensitive: false, origin: 'config', restart_required: false },
];

const fields = [
  { key: 'radarr.url', label: 'URL' },
  { key: 'radarr.enabled', label: 'Enabled', type: 'boolean' as const },
];

function renderCard(onSetSetting = vi.fn().mockResolvedValue(undefined), onClearSetting = vi.fn().mockResolvedValue(undefined)) {
  render(
    <I18nextProvider i18n={i18n}>
      <SettingsGroupCard
        title="Radarr"
        fields={fields}
        settings={settings}
        onSetSetting={onSetSetting}
        onClearSetting={onClearSetting}
      />
    </I18nextProvider>
  );
  return { onSetSetting, onClearSetting };
}

describe('SettingsGroupCard', () => {
  afterEach(() => cleanup());

  it('stages a pending change without calling the API, until Save is clicked', async () => {
    const { onSetSetting } = renderCard();

    expect(screen.queryByText('Save')).not.toBeInTheDocument();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: 'http://new.example.com' } });
    expect(onSetSetting).not.toHaveBeenCalled();
    expect(screen.getByText('Save')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Save'));
    await waitFor(() => expect(onSetSetting).toHaveBeenCalledWith('radarr.url', 'http://new.example.com'));
  });

  it('submits every pending field in the group on confirm', async () => {
    const { onSetSetting } = renderCard();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: 'http://new.example.com' } });
    fireEvent.click(screen.getByRole('checkbox'));
    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => expect(onSetSetting).toHaveBeenCalledTimes(2));
    expect(onSetSetting).toHaveBeenCalledWith('radarr.url', 'http://new.example.com');
    expect(onSetSetting).toHaveBeenCalledWith('radarr.enabled', true);
  });

  it('discards every pending change in the group on cancel', () => {
    const { onSetSetting } = renderCard();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: 'http://new.example.com' } });
    fireEvent.click(screen.getByRole('checkbox'));
    expect(screen.getByText('Cancel')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Cancel'));

    expect(onSetSetting).not.toHaveBeenCalled();
    expect(screen.queryByText('Cancel')).not.toBeInTheDocument();
    expect(screen.getByDisplayValue('http://old.example.com')).toBeInTheDocument();
  });

  it('resets a field to config immediately and independently of a pending edit on another field', async () => {
    const overriddenSettings: SettingsField[] = [
      { key: 'radarr.url', value: 'http://overridden.example.com', sensitive: false, origin: 'interface', restart_required: false },
      { key: 'radarr.enabled', value: false, sensitive: false, origin: 'config', restart_required: false },
    ];
    const onSetSetting = vi.fn().mockResolvedValue(undefined);
    const onClearSetting = vi.fn().mockResolvedValue(undefined);

    render(
      <I18nextProvider i18n={i18n}>
        <SettingsGroupCard
          title="Radarr"
          fields={fields}
          settings={overriddenSettings}
          onSetSetting={onSetSetting}
          onClearSetting={onClearSetting}
        />
      </I18nextProvider>
    );

    // Stage a pending edit on the boolean field.
    fireEvent.click(screen.getByRole('checkbox'));
    expect(screen.getByText('Save')).toBeInTheDocument();

    // Reset-to-config on the URL field fires immediately, not staged.
    fireEvent.click(screen.getByLabelText('Reset to config'));
    await waitFor(() => expect(onClearSetting).toHaveBeenCalledWith('radarr.url'));

    // The unrelated pending boolean change is untouched.
    expect(onSetSetting).not.toHaveBeenCalled();
    expect(screen.getByText('Save')).toBeInTheDocument();
  });
});
