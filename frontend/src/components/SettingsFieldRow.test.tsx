import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { SettingsFieldRow } from './SettingsFieldRow';
import { SettingsField } from '../types';

function renderRow(props: Partial<Parameters<typeof SettingsFieldRow>[0]> & { field: SettingsField }) {
  const onChange = vi.fn();
  const onClear = vi.fn().mockResolvedValue(undefined);
  render(
    <I18nextProvider i18n={i18n}>
      <SettingsFieldRow label="API key" onChange={onChange} onClear={onClear} {...props} />
    </I18nextProvider>
  );
  return { onChange, onClear };
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
const booleanField: SettingsField = {
  key: 'radarr.enabled', value: true, sensitive: false, origin: 'config', restart_required: false,
};

describe('SettingsFieldRow', () => {
  afterEach(() => cleanup());

  it('never displays the raw value for a sensitive field, only whether one is set', () => {
    renderRow({ field: sensitiveSet });
    expect(screen.getByPlaceholderText('Set')).toBeInTheDocument();
    cleanup();
    renderRow({ field: sensitiveUnset });
    expect(screen.getByPlaceholderText('Not set')).toBeInTheDocument();
  });

  it('reports typed changes upward without any save button', () => {
    const { onChange } = renderRow({ field: sensitiveUnset });
    const input = screen.getByPlaceholderText('Not set');
    fireEvent.change(input, { target: { value: 'new-secret-key' } });
    expect(onChange).toHaveBeenCalledWith('new-secret-key');
    expect(screen.queryByText('Save')).not.toBeInTheDocument();
  });

  it('renders a toggle switch for boolean fields', () => {
    const { onChange } = renderRow({ field: booleanField, type: 'boolean', value: 'true' });
    const checkbox = screen.getByRole('checkbox');
    expect(checkbox).toBeChecked();
    fireEvent.click(checkbox);
    expect(onChange).toHaveBeenCalledWith('false');
    expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
  });

  it('shows the "restart required" indicator only for flagged fields', () => {
    renderRow({ field: restartRequiredField });
    expect(screen.getByText('Restart required')).toBeInTheDocument();
    cleanup();
    renderRow({ field: plainField });
    expect(screen.queryByText('Restart required')).not.toBeInTheDocument();
  });

  it('hides the origin label until the compact indicator is hovered or focused', () => {
    renderRow({ field: { ...plainField, origin: 'interface' } });
    expect(screen.queryByText('Interface')).not.toBeInTheDocument();

    const indicator = screen.getByRole('button', { name: 'Interface' });
    fireEvent.focus(indicator);
    expect(screen.getByText('Interface')).toBeInTheDocument();
  });

  it('shows the Config indicator when not overridden and no reset icon', () => {
    renderRow({ field: plainField });
    expect(screen.getByRole('button', { name: 'Config' })).toBeInTheDocument();
    expect(screen.queryByLabelText('Reset to config')).not.toBeInTheDocument();
  });

  it('shows an icon-only reset control (not a text button) when overridden, and calls onClear', () => {
    const { onClear } = renderRow({ field: { ...plainField, origin: 'interface' } });
    expect(screen.queryByText('Reset to config')).not.toBeInTheDocument();
    const resetBtn = screen.getByLabelText('Reset to config');
    fireEvent.click(resetBtn);
    expect(onClear).toHaveBeenCalled();
  });

  it('renders a select with the given options for a select field, staging the chosen value via onChange', () => {
    const options = [
      { value: 'debug', labelKey: 'logLevels.debug' },
      { value: 'info', labelKey: 'logLevels.info' },
      { value: 'warn', labelKey: 'logLevels.warn' },
      { value: 'error', labelKey: 'logLevels.error' },
    ];
    const { onChange } = renderRow({
      field: { key: 'logging.app.level', value: 'info', sensitive: false, origin: 'config', restart_required: false },
      type: 'select',
      options,
      value: 'info',
    });

    const select = screen.getByRole('combobox');
    expect(screen.getByRole('option', { name: 'Debug' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Info' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Warning' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Error' })).toBeInTheDocument();

    fireEvent.change(select, { target: { value: 'error' } });
    expect(onChange).toHaveBeenCalledWith('error');
  });
});
