import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent, waitFor } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { SettingsGroupCard } from './SettingsGroupCard';
import { SettingsField } from '../types';
import { api } from '../services/api';

vi.mock('../services/api', () => ({
  api: {
    testIntegration: vi.fn(),
  },
}));

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

describe('SettingsGroupCard testConfig', () => {
  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  const testSettings: SettingsField[] = [
    { key: 'radarr.url', value: 'http://old.example.com', sensitive: false, origin: 'config', restart_required: false },
    { key: 'radarr.api_key', value: null, sensitive: true, is_set: true, origin: 'config', restart_required: false },
  ];

  const testFields = [
    { key: 'radarr.url', label: 'URL' },
    { key: 'radarr.api_key', label: 'API Key' },
  ];

  function renderTestCard() {
    const onSetSetting = vi.fn().mockResolvedValue(undefined);
    const onClearSetting = vi.fn().mockResolvedValue(undefined);
    render(
      <I18nextProvider i18n={i18n}>
        <SettingsGroupCard
          title="Radarr"
          fields={testFields}
          settings={testSettings}
          onSetSetting={onSetSetting}
          onClearSetting={onClearSetting}
          testConfig={{ service: 'radarr', urlKey: 'radarr.url', apiKeyKey: 'radarr.api_key' }}
        />
      </I18nextProvider>
    );
    return { onSetSetting, onClearSetting };
  }

  it('does not show the Test action without an unsaved change', () => {
    renderTestCard();
    expect(screen.queryByText('Test connection')).not.toBeInTheDocument();
  });

  it('shows the Test action once a field has an unsaved change, using the resolved in-form values', async () => {
    (api.testIntegration as ReturnType<typeof vi.fn>).mockResolvedValue({ status: 'ok' });
    renderTestCard();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: 'http://new.example.com' } });
    expect(screen.getByText('Test connection')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Test connection'));
    await waitFor(() => expect(api.testIntegration).toHaveBeenCalledWith('radarr', 'http://new.example.com', ''));
  });

  it('disables the Test action when the resolved url is empty', () => {
    renderTestCard();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: '' } });

    expect(screen.getByText('Test connection')).toBeDisabled();
  });

  it('disables the Test action while a test is in flight and shows a testing label', async () => {
    let resolveTest: (value: { status: string }) => void = () => {};
    (api.testIntegration as ReturnType<typeof vi.fn>).mockReturnValue(
      new Promise(resolve => { resolveTest = resolve; })
    );
    renderTestCard();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: 'http://new.example.com' } });
    fireEvent.click(screen.getByText('Test connection'));

    expect(await screen.findByText('Testing…')).toBeDisabled();

    resolveTest({ status: 'ok' });
    await waitFor(() => expect(screen.getByText('Test connection')).toBeInTheDocument());
  });

  it('renders an OK badge on a successful test', async () => {
    (api.testIntegration as ReturnType<typeof vi.fn>).mockResolvedValue({ status: 'ok' });
    renderTestCard();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: 'http://new.example.com' } });
    fireEvent.click(screen.getByText('Test connection'));

    await waitFor(() => expect(screen.getByText('OK')).toBeInTheDocument());
  });

  it('renders a KO badge with its reason on a failed test', async () => {
    (api.testIntegration as ReturnType<typeof vi.fn>).mockResolvedValue({ status: 'ko', reason: 'unauthorized' });
    renderTestCard();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: 'http://new.example.com' } });
    fireEvent.click(screen.getByText('Test connection'));

    await waitFor(() => expect(screen.getByText(/KO/)).toBeInTheDocument());
    expect(screen.getByText(/Identifiants invalides|Invalid credentials/)).toBeInTheDocument();
  });

  it('clears the result badge on the next field edit in the card', async () => {
    (api.testIntegration as ReturnType<typeof vi.fn>).mockResolvedValue({ status: 'ok' });
    renderTestCard();

    fireEvent.change(screen.getByDisplayValue('http://old.example.com'), { target: { value: 'http://new.example.com' } });
    fireEvent.click(screen.getByText('Test connection'));
    await waitFor(() => expect(screen.getByText('OK')).toBeInTheDocument());

    fireEvent.change(screen.getByDisplayValue('http://new.example.com'), { target: { value: 'http://another.example.com' } });

    expect(screen.queryByText('OK')).not.toBeInTheDocument();
  });
});
