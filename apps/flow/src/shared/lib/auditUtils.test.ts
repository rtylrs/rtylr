import { describe, it, expect } from 'vitest';

import {
  latestActivityEntry,
  recentActivity,
  activityCountByType,
  uniqueActors,
  activityForActor,
  mostActiveActor,
  firstActivityDate,
} from './auditUtils';
import type { ActivityEntry } from '../stores/flowStore';

function makeEntry(overrides: Partial<ActivityEntry>): ActivityEntry {
  return {
    id: 'e1',
    taskId: 't1',
    type: 'status_change',
    actor: 'Alice',
    actorInitials: 'A',
    createdAt: '2026-06-01',
    ...overrides,
  } as ActivityEntry;
}

describe('latestActivityEntry', () => {
  it('returns most recent entry', () => {
    const activity = [
      makeEntry({ id: 'a', createdAt: '2026-01-01' }),
      makeEntry({ id: 'b', createdAt: '2026-06-15' }),
    ];
    expect(latestActivityEntry(activity)?.id).toBe('b');
  });

  it('returns undefined for empty list', () => {
    expect(latestActivityEntry([])).toBeUndefined();
  });
});

describe('recentActivity', () => {
  it('returns entries within N days', () => {
    const now = new Date('2026-06-15');
    const activity = [
      makeEntry({ createdAt: '2026-06-14' }),
      makeEntry({ id: 'e2', createdAt: '2026-01-01' }),
    ];
    expect(recentActivity(activity, 7, now).length).toBe(1);
  });
});

describe('activityCountByType', () => {
  it('counts entries by type', () => {
    const activity = [
      makeEntry({ type: 'comment' }),
      makeEntry({ id: 'e2', type: 'comment' }),
      makeEntry({ id: 'e3', type: 'status_change' }),
    ];
    const counts = activityCountByType(activity);
    expect(counts.comment).toBe(2);
    expect(counts.status_change).toBe(1);
  });
});

describe('uniqueActors', () => {
  it('returns unique actor names', () => {
    const activity = [
      makeEntry({ actor: 'Alice' }),
      makeEntry({ id: 'e2', actor: 'Bob' }),
      makeEntry({ id: 'e3', actor: 'Alice' }),
    ];
    expect(uniqueActors(activity).sort()).toEqual(['Alice', 'Bob']);
  });
});

describe('activityForActor', () => {
  it('returns entries for the actor', () => {
    const activity = [makeEntry({ actor: 'Alice' }), makeEntry({ id: 'e2', actor: 'Bob' })];
    expect(activityForActor(activity, 'Alice').length).toBe(1);
  });
});

describe('mostActiveActor', () => {
  it('returns actor with most entries', () => {
    const activity = [
      makeEntry({ actor: 'Alice' }),
      makeEntry({ id: 'e2', actor: 'Bob' }),
      makeEntry({ id: 'e3', actor: 'Bob' }),
    ];
    expect(mostActiveActor(activity)).toBe('Bob');
  });

  it('returns null for empty list', () => {
    expect(mostActiveActor([])).toBeNull();
  });
});

describe('firstActivityDate', () => {
  it('returns earliest creation date', () => {
    const activity = [
      makeEntry({ createdAt: '2026-06-15' }),
      makeEntry({ id: 'e2', createdAt: '2026-01-01' }),
    ];
    expect(firstActivityDate(activity)).toBe('2026-01-01');
  });

  it('returns null for empty list', () => {
    expect(firstActivityDate([])).toBeNull();
  });
});
