import { describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { getErrorReasons } from './errorReasons';
import { DownloadEnriched } from '../types';

const t = i18n.getFixedT('en', 'errors');

const baseDownload: DownloadEnriched = {
  id: 1,
  url: 'http://example.com/movie.mkv',
  status: 'completed',
  retry_count: 0,
  updated_at: '2026-08-27T10:00:00Z',
};

describe('getErrorReasons', () => {
  it('produces the duplicate download badge when duplicate_siblings is present', () => {
    const download: DownloadEnriched = {
      ...baseDownload,
      duplicate_siblings: [{ download_info_id: 2, download_path: '/movies/dup.mkv' }],
    };

    const reasons = getErrorReasons(download, t);

    expect(reasons).toContainEqual({ key: 'duplicate_download', label: t('badges.duplicateDownload') });
  });

  it('does not produce the duplicate download badge when duplicate_siblings is absent', () => {
    const reasons = getErrorReasons(baseDownload, t);

    expect(reasons.some(r => r.key === 'duplicate_download')).toBe(false);
  });

  it('does not produce the duplicate download badge when duplicate_siblings is empty', () => {
    const download: DownloadEnriched = { ...baseDownload, duplicate_siblings: [] };

    const reasons = getErrorReasons(download, t);

    expect(reasons.some(r => r.key === 'duplicate_download')).toBe(false);
  });

  it('combines the duplicate badge with other applicable reasons', () => {
    const download: DownloadEnriched = {
      ...baseDownload,
      duplicate_siblings: [{ download_info_id: 2, download_path: '/movies/dup.mkv' }],
      file_info: {
        extension: '.mkv',
        folder_name: 'Movie',
        file_name: 'movie.mkv',
        has_year_in_path: false,
        year_mismatch: false,
        is_valid_format: true,
      },
    };

    const reasons = getErrorReasons(download, t);

    expect(reasons.map(r => r.key)).toEqual(['duplicate_download', 'missing_year']);
  });
});
