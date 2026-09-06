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
