import { afterEach, describe, expect, it, vi } from 'vitest';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '../i18n';
import { M3uSourcesSection } from './M3uSourcesSection';
import { M3uSource } from '../types';

const originSource: M3uSource = {
  name: 'a', file_path: '/tmp/a.m3u', enabled: true, url: '', archive_dir: '',
  retention_count: 5, max_file_size_mb: 500, timeout_seconds: 300, retry_attempts: 3,
  auth_username: '', has_auth_password: false, is_runtime: false,
};

const overriddenSource: M3uSource = { ...originSource, name: 'b', is_runtime: true };
const runtimeOnlySource: M3uSource = { ...originSource, name: 'c', is_runtime: true };

function renderSection(sources: M3uSource[], originNames: Set<string>, searchQuery?: string) {
  const onCreate = vi.fn().mockResolvedValue(undefined);
  const onUpdate = vi.fn().mockResolvedValue(undefined);
  const onDelete = vi.fn().mockResolvedValue(undefined);

  render(
    <I18nextProvider i18n={i18n}>
      <M3uSourcesSection
        isExpanded
        sources={sources}
        originNames={originNames}
        loading={false}
        onCreate={onCreate}
        onUpdate={onUpdate}
        onDelete={onDelete}
        searchQuery={searchQuery}
      />
    </I18nextProvider>
  );

  return { onCreate, onUpdate, onDelete };
}

describe('M3uSourcesSection', () => {
  afterEach(() => cleanup());

  it('creates a new runtime-only source', async () => {
    const { onCreate } = renderSection([originSource], new Set(['a']));

    fireEvent.click(screen.getByText('+ Add source'));
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'new-source' } });
    fireEvent.change(screen.getByLabelText('File path'), { target: { value: '/tmp/new.m3u' } });
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onCreate).toHaveBeenCalled());
    expect(onCreate.mock.calls[0][0]).toBe('new-source');
    expect(onCreate.mock.calls[0][1].file_path).toBe('/tmp/new.m3u');
  });

  it('editing an origin-defined source calls onUpdate, creating a runtime override for it', async () => {
    const { onUpdate } = renderSection([originSource], new Set(['a']));

    fireEvent.click(screen.getByText('Edit'));
    fireEvent.change(screen.getByLabelText('File path'), { target: { value: '/tmp/a-override.m3u' } });
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onUpdate).toHaveBeenCalled());
    expect(onUpdate.mock.calls[0][0]).toBe('a');
    expect(onUpdate.mock.calls[0][1].file_path).toBe('/tmp/a-override.m3u');
  });

  it('shows an inline warning banner (not a native confirm) before replacing an existing runtime-only source', async () => {
    const { onCreate } = renderSection([runtimeOnlySource], new Set());

    fireEvent.click(screen.getByText('+ Add source'));
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'c' } });
    expect(screen.getByRole('alert')).toHaveTextContent('already exists as a source created here');

    fireEvent.change(screen.getByLabelText('File path'), { target: { value: '/tmp/c-new.m3u' } });
    fireEvent.click(screen.getByText('Save'));

    await vi.waitFor(() => expect(onCreate).toHaveBeenCalled());
  });

  it('deleting a runtime-overridden source calls onDelete with its name', () => {
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    const { onDelete } = renderSection([overriddenSource], new Set(['b']));

    fireEvent.click(screen.getByText('Delete'));

    expect(onDelete).toHaveBeenCalledWith('b');
    confirmSpy.mockRestore();
  });

  it('does not show a delete button for an origin-only source', () => {
    renderSection([originSource], new Set(['a']));
    expect(screen.queryByText('Delete')).not.toBeInTheDocument();
  });

  it('filters the visible source list by the search query', () => {
    renderSection([originSource, overriddenSource], new Set(['a']), 'b');
    expect(screen.getByText('b')).toBeInTheDocument();
    expect(screen.queryByText('a')).not.toBeInTheDocument();
  });

  it('shows no auth-password text for a source with no password set', () => {
    renderSection([originSource], new Set(['a']));
    expect(screen.queryByText(/Auth password/)).not.toBeInTheDocument();
  });

  it('shows the "set" indicator for a source with a password set', () => {
    renderSection([{ ...originSource, has_auth_password: true }], new Set(['a']));
    expect(screen.getByText('Auth password: Set')).toBeInTheDocument();
  });
});
