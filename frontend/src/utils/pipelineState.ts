export function getPipelineStateBadgeClass(state: string): string {
  if (state === 'processed' || state === 'downloaded') return 'badge-success';
  if (state === 'downloading' || state === 'organizing') return 'badge-progress';
  if (state === 'failed') return 'badge-failed';
  return 'badge-pending';
}
