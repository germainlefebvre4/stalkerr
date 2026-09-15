import { afterEach, describe, expect, it, vi } from 'vitest';
import { api } from './api';

describe('api.getGroupItems', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('appends processing_log_id when the group reports a latest_processing_log_id', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: [], total: 0, limit: 10, offset: 0, total_pages: 0 }),
    });
    vi.stubGlobal('fetch', fetchMock);

    await api.getGroupItems({ type: 'movie', movie_id: 5, latest_processing_log_id: 42 }, 1);

    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).toContain('&movie_id=5');
    expect(url).toContain('&processing_log_id=42');
  });

  it('omits processing_log_id when the group has no latest_processing_log_id', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: [], total: 0, limit: 10, offset: 0, total_pages: 0 }),
    });
    vi.stubGlobal('fetch', fetchMock);

    await api.getGroupItems({ type: 'movie', movie_id: 5 }, 1);

    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).toContain('&movie_id=5');
    expect(url).not.toContain('processing_log_id');
  });
});

describe('api.testIntegration', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('posts the service/url/api_key to the connectivity-test endpoint and returns the result', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ status: 'ok' }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const result = await api.testIntegration('radarr', 'http://radarr.local', 'my-key');

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/settings/integrations/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ service: 'radarr', url: 'http://radarr.local', api_key: 'my-key' }),
    });
    expect(result).toEqual({ status: 'ok' });
  });

  it('returns a KO result with reason', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ status: 'ko', reason: 'unauthorized' }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const result = await api.testIntegration('jellyfin', 'http://jellyfin.local', '');

    expect(result).toEqual({ status: 'ko', reason: 'unauthorized' });
  });
});
