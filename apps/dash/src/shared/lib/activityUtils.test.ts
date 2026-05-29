import { describe, it, expect } from 'vitest';
import {
  activityLabel,
  activityDescription,
  unreadActivities,
  unreadActivityCount,
  filterByType,
  sortByDate,
  recentEntries,
} from './activityUtils';
import type { ActivityEntry } from './activityUtils';

function makeEntry(overrides: Partial<ActivityEntry> = {}): ActivityEntry {
  return {
    id: '1',
    type: 'update',
    actorName: 'Alice',
    createdAt: '2026-05-01T10:00:00Z',
    read: false,
    ...overrides,
  };
}

describe('activityLabel', () => {
  it('returns "Created" for create', () => {
    expect(activityLabel('create')).toBe('Created');
  });

  it('returns "Logged in" for login', () => {
    expect(activityLabel('login')).toBe('Logged in');
  });
});

describe('activityDescription', () => {
  it('includes actor and action', () => {
    const desc = activityDescription(makeEntry({ type: 'update', actorName: 'Bob' }));
    expect(desc).toContain('Bob');
    expect(desc).toContain('updated');
  });

  it('includes target name when provided', () => {
    const desc = activityDescription(
      makeEntry({ type: 'create', actorName: 'Alice', targetName: 'Invoice #42' })
    );
    expect(desc).toContain('Invoice #42');
  });

  it('omits target when not provided', () => {
    const desc = activityDescription(makeEntry({ type: 'login', actorName: 'Carol' }));
    expect(desc).toBe('Carol logged in');
  });
});

describe('unreadActivities', () => {
  it('returns only unread entries', () => {
    const entries = [
      makeEntry({ id: '1', read: false }),
      makeEntry({ id: '2', read: true }),
      makeEntry({ id: '3', read: false }),
    ];
    expect(unreadActivities(entries)).toHaveLength(2);
  });
});

describe('unreadActivityCount', () => {
  it('counts unread entries', () => {
    const entries = [
      makeEntry({ read: false }),
      makeEntry({ read: true }),
      makeEntry({ read: false }),
    ];
    expect(unreadActivityCount(entries)).toBe(2);
  });
});

describe('filterByType', () => {
  it('returns only entries of the specified type', () => {
    const entries = [
      makeEntry({ type: 'create' }),
      makeEntry({ type: 'update' }),
      makeEntry({ type: 'create' }),
    ];
    const result = filterByType(entries, 'create');
    expect(result).toHaveLength(2);
  });
});

describe('sortByDate', () => {
  it('sorts newest first', () => {
    const entries = [
      makeEntry({ id: '1', createdAt: '2026-01-01T00:00:00Z' }),
      makeEntry({ id: '2', createdAt: '2026-03-01T00:00:00Z' }),
    ];
    const sorted = sortByDate(entries);
    expect(sorted[0]!.id).toBe('2');
  });
});

describe('recentEntries', () => {
  it('returns entries within the last N days', () => {
    const now = new Date('2026-05-10T00:00:00Z');
    const entries = [
      makeEntry({ id: '1', createdAt: '2026-05-09T00:00:00Z' }),
      makeEntry({ id: '2', createdAt: '2026-04-01T00:00:00Z' }),
    ];
    expect(recentEntries(entries, 7, now)).toHaveLength(1);
  });
});
