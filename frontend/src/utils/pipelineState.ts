export function getPipelineStateBadgeClass(state: string): string {
  if (state === 'processed' || state === 'downloaded') return 'badge-success';
  if (state === 'downloading' || state === 'organizing') return 'badge-progress';
  if (state === 'failed') return 'badge-failed';
  return 'badge-pending';
}

export type ProcessingStatus = 'pending' | 'processed';
export type DownloadStatus = 'not_downloaded' | 'downloading' | 'organizing' | 'downloaded' | 'failed';

export function getProcessingStatus(state: string): ProcessingStatus {
  return state === 'pending' ? 'pending' : 'processed';
}

export function getDownloadStatus(state: string): DownloadStatus {
  if (state === 'pending' || state === 'processed') return 'not_downloaded';
  return state as DownloadStatus;
}

export function getProcessingStatusBadgeClass(status: ProcessingStatus): string {
  return status === 'processed' ? 'badge-success' : 'badge-pending';
}

export function getDownloadStatusBadgeClass(status: DownloadStatus): string {
  if (status === 'downloaded') return 'badge-success';
  if (status === 'downloading' || status === 'organizing') return 'badge-progress';
  if (status === 'failed') return 'badge-failed';
  return 'badge-neutral';
}
