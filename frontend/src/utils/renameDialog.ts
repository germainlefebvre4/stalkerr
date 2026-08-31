import { DownloadEnriched } from '../types';

export function resolveRenameFolderName(item: DownloadEnriched, filename: string): string {
  return item.rename_folder_name || item.file_info?.folder_name || filename;
}
