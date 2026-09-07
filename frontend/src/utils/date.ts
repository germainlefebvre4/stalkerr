export function formatDate(value: string | Date, locale: string = 'en'): string {
  const date = typeof value === 'string' ? new Date(value) : value;
  return new Intl.DateTimeFormat(locale, { year: 'numeric', month: '2-digit', day: '2-digit' }).format(date);
}

/**
 * Returns the elapsed duration between `startedAt` and `completedAt` as a
 * fixed-format string (independent of locale, mirroring formatDate's
 * deterministic DD/MM/YYYY convention), or `null` when the run is still in
 * progress (no `completedAt` yet) - callers render an "in progress" state instead.
 */
export function formatRunDuration(startedAt: string, completedAt?: string | null): string | null {
  if (!completedAt) return null;

  const totalSeconds = Math.max(0, Math.round((new Date(completedAt).getTime() - new Date(startedAt).getTime()) / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) return `${hours}h ${String(minutes).padStart(2, '0')}m`;
  if (minutes > 0) return `${minutes}m ${String(seconds).padStart(2, '0')}s`;
  return `${seconds}s`;
}

function isSameLocalDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

/**
 * Given a sequence of items sorted by timestamp (by default `created_at`, or
 * another field via `getTimestamp`, e.g. a grouped row's `latest_activity`),
 * returns for each index whether that item starts a new local-calendar-day
 * group (the first item always does).
 */
export function getDateGroupStarts<T extends { created_at: string }>(items: T[]): boolean[];
export function getDateGroupStarts<T>(items: T[], getTimestamp: (item: T) => string): boolean[];
export function getDateGroupStarts<T>(items: T[], getTimestamp?: (item: T) => string): boolean[] {
  const extract = getTimestamp ?? ((item: T) => (item as unknown as { created_at: string }).created_at);
  return items.map((item, index) => {
    if (index === 0) return true;
    const previous = new Date(extract(items[index - 1]));
    const current = new Date(extract(item));
    return !isSameLocalDay(current, previous);
  });
}

/**
 * Returns the translated "Today"/"Yesterday" label for a date-group header, or the
 * full formatted date (same format as the date cells) for any earlier day.
 */
export function getDateGroupLabel(date: Date, t: (key: string) => string, locale: string = 'en'): string {
  const now = new Date();
  if (isSameLocalDay(date, now)) return t('dateGroups.today');

  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  if (isSameLocalDay(date, yesterday)) return t('dateGroups.yesterday');

  return formatDate(date, locale);
}
