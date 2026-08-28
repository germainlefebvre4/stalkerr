export function getLogStatusBadgeClass(status: string): string {
  return status === 'success' ? 'badge-success' : status === 'failed' ? 'badge-failed' : 'badge-progress';
}
