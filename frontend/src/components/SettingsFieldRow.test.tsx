import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { SettingsFieldRow } from './SettingsFieldRow';
import { SettingsField } from '../types';

function renderRow(field: SettingsField, onSave = vi.fn(), onClear = vi.fn(), type: 'text' | 'number' | 'boolean' = 'text') {
  render(
    <I18nextProvider i18n={i18n}>
      <SettingsFieldRow label="API key" field={field} type={type} onSave={onSave} onClear={onClear} />
    </I18nextProvider>
  );
  return { onSave, onClear };
}

const sensitiveUnset: SettingsField = {
  key: 'radarr.api_key', is_set: false, sensitive: true, origin: 'config', restart_required: false,
};
const sensitiveSet: SettingsField = {
  key: 'radarr.api_key', is_set: true, sensitive: true, origin: 'config', restart_required: false,
};
const plainField: SettingsField = {
  key: 'radarr.url', value: 'http://example.com', sensitive: false, origin: 'config', restart_required: false,
};
const restartRequiredField: SettingsField = {
  key: 'tmdb.api_key', is_set: false, sensitive: true, origin: 'config', restart_required: true,
};

describe('SettingsFieldRow', () => {
  afterEach(() => cleanup());

  it('never displays the raw value for a sensitive field, only whether one is set', () => {
    renderRow(sensitiveSet);
    expect(screen.getByPlaceholderText('Set')).toBeInTheDocument();
    renderRow(sensitiveUnset);
    expect(screen.getByPlaceholderText('Not set')).toBeInTheDocument();
  });

  it('does not submit any change when a sensitive field is left untouched', () => {
    const { onSave } = renderRow(sensitiveSet);
    fireEvent.click(screen.getByText('Save'));
    expect(onSave).not.toHaveBeenCalled();
  });

  it('submits the typed value when a sensitive field is edited', () => {
    const { onSave } = renderRow(sensitiveUnset);
    const input = screen.getByPlaceholderText('Not set');
    fireEvent.change(input, { target: { value: 'new-secret-key' } });
    fireEvent.click(screen.getByText('Save'));
    expect(onSave).toHaveBeenCalledWith('new-secret-key');
  });

  it('shows the "restart required" indicator only for flagged fields', () => {
    renderRow(restartRequiredField);
    expect(screen.getByText('Restart required')).toBeInTheDocument();
    cleanup();
    renderRow(plainField);
    expect(screen.queryByText('Restart required')).not.toBeInTheDocument();
  });

  it('shows the Interface badge when overridden and Config otherwise', () => {
    renderRow(plainField);
    expect(screen.getByText('Config')).toBeInTheDocument();
    expect(screen.queryByText('Reset to config')).not.toBeInTheDocument();

    cleanup();
    renderRow({ ...plainField, origin: 'interface' });
    expect(screen.getByText('Interface')).toBeInTheDocument();
    expect(screen.getByText('Reset to config')).toBeInTheDocument();
  });

  it('calls onClear when "Reset to config" is clicked', () => {
    const { onClear } = renderRow({ ...plainField, origin: 'interface' });
    fireEvent.click(screen.getByText('Reset to config'));
    expect(onClear).toHaveBeenCalled();
  });
});
