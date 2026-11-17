import { describe, expect, it } from 'vitest';
import type { ApiStatus } from '../../shared/services/flowApi';
import {
  closedRecently,
  defaultStatus,
  firstDoneStatus,
  insertionPlan,
  isOnlyDoneStatus,
  movedStatusIds,
} from './statusUtils';

const status = (
  id: string,
  sortOrder: number,
  bucket: ApiStatus['bucket'] = 'active',
  isDefault = false
): ApiStatus => ({
  id,
  projectId: '1',
  name: `S${id}`,
  bucket,
  legacyStatus: 'todo',
  color: '#8b5cf6',
  sortOrder,
  isDefault,
  createdAt: '',
});

const NOW = Date.parse('2026-11-16T12:00:00Z');
const daysAgo = (days: number): string => new Date(NOW - days * 86_400_000).toISOString();

describe('closedRecently (the 14-day window on done and cancelled columns)', () => {
  it('keeps a done card completed inside the window and hides an older one', () => {
    expect(closedRecently({ completedAt: daysAgo(13), updatedAt: daysAgo(1) }, 'done', NOW)).toBe(
      true
    );
    expect(closedRecently({ completedAt: daysAgo(15), updatedAt: daysAgo(1) }, 'done', NOW)).toBe(
      false
    );
  });

  it('reads updatedAt for cancelled cards, not completedAt', () => {
    expect(
      closedRecently({ completedAt: daysAgo(1), updatedAt: daysAgo(20) }, 'cancelled', NOW)
    ).toBe(false);
    expect(closedRecently({ updatedAt: daysAgo(2) }, 'cancelled', NOW)).toBe(true);
  });

  it('falls back to updatedAt for a done card without completedAt, and shows unknowns', () => {
    expect(closedRecently({ updatedAt: daysAgo(30) }, 'done', NOW)).toBe(false);
    expect(closedRecently({ updatedAt: 'not a date' }, 'done', NOW)).toBe(true);
  });
});

describe('status picks', () => {
  const set = [
    status('1', 0, 'backlog'),
    status('2', 10, 'active', true),
    status('3', 900, 'done'),
  ];

  it('finds the first done status and the default', () => {
    expect(firstDoneStatus(set)?.id).toBe('3');
    expect(defaultStatus(set)?.id).toBe('2');
  });

  it('protects only the last done status', () => {
    expect(isOnlyDoneStatus(set[2], set)).toBe(true);
    expect(isOnlyDoneStatus(set[2], [...set, status('4', 950, 'done')])).toBe(false);
    expect(isOnlyDoneStatus(set[0], set)).toBe(false);
  });
});

describe('movedStatusIds', () => {
  const ordered = [status('1', 0), status('2', 10), status('3', 20)];

  it('returns the whole set in its new order', () => {
    expect(movedStatusIds(ordered, 2, 1)).toEqual(['1', '3', '2']);
    expect(movedStatusIds(ordered, 0, 2)).toEqual(['2', '3', '1']);
  });

  it('moves tied sort orders too, since the server rewrites every position', () => {
    const tied = [status('1', 0), status('2', 0), status('3', 0)];
    expect(movedStatusIds(tied, 0, 1)).toEqual(['2', '1', '3']);
  });

  it('returns null when nothing would move', () => {
    expect(movedStatusIds(ordered, 1, 1)).toBeNull();
    expect(movedStatusIds(ordered, 0, -1)).toBeNull();
    expect(movedStatusIds(ordered, 2, 3)).toBeNull();
  });
});

describe('insertionPlan', () => {
  const ordered = [status('1', 0), status('2', 10), status('3', 900)];

  it('uses the gap between neighbours', () => {
    expect(insertionPlan(ordered, 2)).toEqual({ sortOrder: 455 });
  });

  it('appends after the last status', () => {
    expect(insertionPlan(ordered, 3)).toEqual({ sortOrder: 910 });
  });

  it('places with one reorder when there is no room before the first status', () => {
    expect(insertionPlan(ordered, 0)).toEqual({ placeAmong: { ids: ['1', '2', '3'], index: 0 } });
  });
});
