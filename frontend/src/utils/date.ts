export function formatDate(value: string | Date, locale: string = 'en'): string {
  const date = typeof value === 'string' ? new Date(value) : value;
  return new Intl.DateTimeFormat(locale, { year: 'numeric', month: '2-digit', day: '2-digit' }).format(date);
}

function isSameLocalDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

/**
 * Given a sequence of items sorted by `created_at`, returns for each index whether
 * that item starts a new local-calendar-day group (the first item always does).
 */
export function getDateGroupStarts<T extends { created_at: string }>(items: T[]): boolean[] {
  return items.map((item, index) => {
    if (index === 0) return true;
    const previous = new Date(items[index - 1].created_at);
    const current = new Date(item.created_at);
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
