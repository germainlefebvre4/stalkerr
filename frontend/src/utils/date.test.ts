import { describe, expect, it } from 'vitest';
import { getDateGroupLabel, getDateGroupStarts } from './date';

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
