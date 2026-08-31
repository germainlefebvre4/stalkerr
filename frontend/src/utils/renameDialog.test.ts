import { describe, expect, it } from 'vitest';
import { resolveRenameFolderName } from './renameDialog';
import { DownloadEnriched } from '../types';

const baseItem: DownloadEnriched = {
  id: 1,
  url: 'http://example.com/download',
  status: 'completed',
  retry_count: 0,
  updated_at: '2026-08-27T10:00:00Z',
};

describe('resolveRenameFolderName', () => {
  it('pre-fills the series folder name for a TV episode, not the season folder', () => {
    const item: DownloadEnriched = {
      ...baseItem,
      file_info: {
        extension: '.mkv',
        folder_name: 'Season 01',
        file_name: 'Breaking Bad (2008) - S01E01.mkv',
        has_year_in_path: true,
        year_mismatch: false,
        is_valid_format: true,
      },
      rename_folder_name: 'Breaking Bad (2008)',
    };

    expect(resolveRenameFolderName(item, 'S01E01.mkv')).toBe('Breaking Bad (2008)');
  });

  it('pre-fills the folder name for a movie, matching file_info.folder_name', () => {
    const item: DownloadEnriched = {
      ...baseItem,
      file_info: {
        extension: '.mkv',
        folder_name: 'Interstellar (2014)',
        file_name: 'Interstellar (2014).mkv',
        has_year_in_path: true,
        year_mismatch: false,
        is_valid_format: true,
      },
      rename_folder_name: 'Interstellar (2014)',
    };

    expect(resolveRenameFolderName(item, 'Interstellar (2014).mkv')).toBe('Interstellar (2014)');
  });

  it('falls back to file_info.folder_name when rename_folder_name is absent', () => {
    const item: DownloadEnriched = {
      ...baseItem,
      file_info: {
        extension: '.mkv',
        folder_name: 'Some Folder',
        file_name: 'file.mkv',
        has_year_in_path: true,
        year_mismatch: false,
        is_valid_format: true,
      },
    };

    expect(resolveRenameFolderName(item, 'file.mkv')).toBe('Some Folder');
  });

  it('falls back to the filename when neither rename_folder_name nor file_info are present', () => {
    expect(resolveRenameFolderName(baseItem, 'file.mkv')).toBe('file.mkv');
  });
});
