import { describe, it, expect } from 'vitest';

import {
  activityForTask,
  sortActivityDesc,
  sortActivityAsc,
  activityByType,
  activityTypeLabel,
  latestActivity,
  activityByActor,
} from './activityUtils';
import type { ActivityEntry } from '../stores/flowStore';

function makeEntry(overrides: Partial<ActivityEntry>): ActivityEntry {
  return {
    id: 'e1',
    taskId: 't1',
    type: 'created',
    actor: 'Alice',
    actorInitials: 'AL',
    createdAt: '2026-05-01T10:00:00Z',
    ...overrides,
  } as ActivityEntry;
}

describe('activityForTask', () => {
  it('returns entries for the given task', () => {
    const activity = [makeEntry({ id: 'a', taskId: 't1' }), makeEntry({ id: 'b', taskId: 't2' })];
    expect(activityForTask(activity, 't1').map((e) => e.id)).toEqual(['a']);
  });
});

describe('sortActivityDesc', () => {
  it('sorts newest first', () => {
    const activity = [
      makeEntry({ id: 'a', createdAt: '2026-01-01T00:00:00Z' }),
      makeEntry({ id: 'b', createdAt: '2026-06-01T00:00:00Z' }),
    ];
    expect(sortActivityDesc(activity).map((e) => e.id)).toEqual(['b', 'a']);
  });

  it('does not mutate original', () => {
    const activity = [
      makeEntry({ createdAt: '2026-06-01T00:00:00Z' }),
      makeEntry({ createdAt: '2026-01-01T00:00:00Z' }),
    ];
    const original = [...activity];
    sortActivityDesc(activity);
    expect(activity).toEqual(original);
  });
});

describe('sortActivityAsc', () => {
  it('sorts oldest first', () => {
    const activity = [
      makeEntry({ id: 'a', createdAt: '2026-06-01T00:00:00Z' }),
      makeEntry({ id: 'b', createdAt: '2026-01-01T00:00:00Z' }),
    ];
    expect(sortActivityAsc(activity).map((e) => e.id)).toEqual(['b', 'a']);
  });
});

describe('activityByType', () => {
  it('filters by type', () => {
    const activity = [
      makeEntry({ id: 'a', type: 'created' }),
      makeEntry({ id: 'b', type: 'comment' }),
      makeEntry({ id: 'c', type: 'created' }),
    ];
    expect(activityByType(activity, 'created').map((e) => e.id)).toEqual(['a', 'c']);
  });
});

describe('activityTypeLabel', () => {
  it('returns label for created', () => {
    expect(activityTypeLabel('created')).toBe('created this task');
  });

  it('returns label for status_change', () => {
    expect(activityTypeLabel('status_change')).toBe('changed status');
  });

  it('returns label for comment', () => {
    expect(activityTypeLabel('comment')).toBe('left a comment');
  });

  it('returns label for attachment', () => {
    expect(activityTypeLabel('attachment')).toBe('added an attachment');
  });
});

describe('latestActivity', () => {
  it('returns most recent entry for task', () => {
    const activity = [
      makeEntry({ id: 'a', taskId: 't1', createdAt: '2026-01-01T00:00:00Z' }),
      makeEntry({ id: 'b', taskId: 't1', createdAt: '2026-06-01T00:00:00Z' }),
    ];
    expect(latestActivity(activity, 't1')?.id).toBe('b');
  });

  it('returns undefined for unknown task', () => {
    expect(latestActivity([], 'x')).toBeUndefined();
  });
});

describe('activityByActor', () => {
  it('returns entries by the given actor', () => {
    const activity = [
      makeEntry({ id: 'a', actor: 'Alice' }),
      makeEntry({ id: 'b', actor: 'Bob' }),
      makeEntry({ id: 'c', actor: 'Alice' }),
    ];
    expect(activityByActor(activity, 'Alice').map((e) => e.id)).toEqual(['a', 'c']);
  });
});
