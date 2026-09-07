import { describe, expect, it } from 'vitest';
import { formatRunDuration, getDateGroupLabel, getDateGroupStarts } from './date';

describe('formatRunDuration', () => {
  it('returns null for an in-progress run (no completedAt)', () => {
    expect(formatRunDuration('2026-08-27T10:00:00Z', null)).toBeNull();
    expect(formatRunDuration('2026-08-27T10:00:00Z', undefined)).toBeNull();
  });

  it('formats a sub-minute duration in seconds', () => {
    expect(formatRunDuration('2026-08-27T10:00:00Z', '2026-08-27T10:00:42Z')).toBe('42s');
  });

  it('formats a sub-hour duration in minutes and seconds', () => {
    expect(formatRunDuration('2026-08-27T10:00:00Z', '2026-08-27T10:05:09Z')).toBe('5m 09s');
  });

  it('formats an hour-plus duration in hours and minutes', () => {
    expect(formatRunDuration('2026-08-27T10:00:00Z', '2026-08-27T11:30:00Z')).toBe('1h 30m');
  });

  it('is stable across repeated calls for the same input (deterministic, not locale-dependent)', () => {
    const a = formatRunDuration('2026-08-27T10:00:00Z', '2026-08-27T10:05:09Z');
    const b = formatRunDuration('2026-08-27T10:00:00Z', '2026-08-27T10:05:09Z');
    expect(a).toBe(b);
  });
});

describe('getDateGroupStarts', () => {
  it('marks the first item as starting a group', () => {
    const items = [{ created_at: '2026-08-27T10:00:00Z' }];
    expect(getDateGroupStarts(items)).toEqual([true]);
  });

  it('does not start a new group for consecutive same-day items', () => {
    const items = [
      { created_at: '2026-08-27T08:00:00' },
      { created_at: '2026-08-27T10:00:00' },
      { created_at: '2026-08-27T22:00:00' },
    ];
    expect(getDateGroupStarts(items)).toEqual([true, false, false]);
  });

  it('starts a new group at a day boundary mid-array', () => {
    const items = [
      { created_at: '2026-08-27T10:00:00' },
      { created_at: '2026-08-27T12:00:00' },
      { created_at: '2026-08-26T09:00:00' },
      { created_at: '2026-08-26T11:00:00' },
      { created_at: '2026-08-25T14:00:00' },
    ];
    expect(getDateGroupStarts(items)).toEqual([true, false, true, false, true]);
  });

  it('supports a custom timestamp field, e.g. grouped rows keyed off latest_activity', () => {
    const groups = [
      { latest_activity: '2026-08-27T10:00:00' },
      { latest_activity: '2026-08-27T12:00:00' },
      { latest_activity: '2026-08-26T09:00:00' },
    ];
    expect(getDateGroupStarts(groups, g => g.latest_activity)).toEqual([true, false, true]);
  });
});

describe('getDateGroupLabel', () => {
  const t = (key: string) => (key === 'dateGroups.today' ? "Aujourd'hui" : 'Hier');

  it('returns the translated "Today" label for the current calendar day', () => {
    const now = new Date();
    expect(getDateGroupLabel(now, t, 'fr')).toBe("Aujourd'hui");
  });

  it('returns the translated "Yesterday" label for the previous calendar day', () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    expect(getDateGroupLabel(yesterday, t, 'fr')).toBe('Hier');
  });

  it('falls back to the full formatted date for an older day, in fr', () => {
    const older = new Date('2020-01-15T12:00:00');
    expect(getDateGroupLabel(older, t, 'fr')).toBe('15/01/2020');
  });

  it('falls back to the full formatted date for an older day, in en', () => {
    const older = new Date('2020-01-15T12:00:00');
    expect(getDateGroupLabel(older, t, 'en')).toBe('01/15/2020');
  });
});
