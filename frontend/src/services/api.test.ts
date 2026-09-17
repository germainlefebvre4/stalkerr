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

describe('api.listRadarrMovies', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('sends the selected État values as repeated status query parameters', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: [], total: 0, limit: 20, offset: 0, total_pages: 0 }),
    });
    vi.stubGlobal('fetch', fetchMock);

    await api.listRadarrMovies(1, 20, undefined, undefined, new Set(['monitored', 'missing']));

    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).toContain('&status=monitored');
    expect(url).toContain('&status=missing');
  });

  it('omits the status parameter entirely when none is supplied', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: [], total: 0, limit: 20, offset: 0, total_pages: 0 }),
    });
    vi.stubGlobal('fetch', fetchMock);

    await api.listRadarrMovies(1, 20);

    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).not.toContain('status=');
  });
});

describe('api.listSonarrSeries', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('sends the selected État values as repeated status query parameters', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: [], total: 0, limit: 20, offset: 0, total_pages: 0 }),
    });
    vi.stubGlobal('fetch', fetchMock);

    await api.listSonarrSeries(1, 20, undefined, undefined, undefined, new Set(['unmonitored', 'missing']));

    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).toContain('&status=unmonitored');
    expect(url).toContain('&status=missing');
  });
});

describe('api.dryRunFilter', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('posts the source/attribute/patterns to the dry-run endpoint and returns the summary', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ no_archive: false, total_lines: 3, matched_count: 2, excluded_count: 1, top_matched: [], top_excluded: [] }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const result = await api.dryRunFilter({
      source_name: 'main',
      attributes: ['group_title'],
      group_title_include_patterns: 'FRENCH',
      group_title_exclude_patterns: 'XXX',
    });

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/filters/dryrun', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        source_name: 'main',
        attributes: ['group_title'],
        group_title_include_patterns: 'FRENCH',
        group_title_exclude_patterns: 'XXX',
      }),
    });
    expect(result).toEqual({ no_archive: false, total_lines: 3, matched_count: 2, excluded_count: 1, top_matched: [], top_excluded: [] });
  });

  it('returns the search-results shape when search is set', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ no_archive: false, results: [{ group_title: 'Movies', tvg_name: 'A', matched: true }], truncated: false }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const result = await api.dryRunFilter({ source_name: 'main', attributes: ['group_title'], search: 'Movies' });

    expect(result).toEqual({ no_archive: false, results: [{ group_title: 'Movies', tvg_name: 'A', matched: true }], truncated: false });
  });
});
