import { TFunction } from 'i18next';
import { DownloadEnriched } from '../types';

export interface ErrorReasonBadge {
  key: 'missing_year' | 'year_mismatch' | 'unknown_format';
  label: string;
}

export function getErrorReasons(item: DownloadEnriched, t: TFunction): ErrorReasonBadge[] {
  const badges: ErrorReasonBadge[] = [];
  if (!item.file_info) return badges;
  if (!item.file_info.has_year_in_path) {
    badges.push({ key: 'missing_year', label: t('badges.missingYear') });
  }
  if (item.file_info.year_mismatch) {
    badges.push({ key: 'year_mismatch', label: t('badges.yearMismatch') });
  }
  if (!item.file_info.is_valid_format) {
    badges.push({ key: 'unknown_format', label: t('badges.unknownFormat') });
  }
  return badges;
}
