import { describe, expect, it } from 'vitest';
import type { ApiStatus } from '../../shared/services/flowApi';
import {
  closedRecently,
  defaultStatus,
  displayStatusName,
  firstDoneStatus,
  insertionPlan,
  isOnlyDoneStatus,
  reorderUpdates,
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

describe('reorderUpdates', () => {
  it('swaps neighbours with two writes', () => {
    const ordered = [status('1', 0), status('2', 10), status('3', 20)];
    expect(reorderUpdates(ordered, 2, 1).map((u) => [u.status.id, u.sortOrder])).toEqual([
      ['3', 10],
      ['2', 20],
    ]);
  });

  it('renumbers tied sort orders so the move takes effect', () => {
    const ordered = [status('1', 0), status('2', 0), status('3', 0)];
    expect(reorderUpdates(ordered, 0, 1).map((u) => [u.status.id, u.sortOrder])).toEqual([
      ['1', 10],
      ['3', 20],
    ]);
  });
});

describe('insertionPlan', () => {
  const ordered = [status('1', 0), status('2', 10), status('3', 900)];

  it('uses the gap between neighbours', () => {
    expect(insertionPlan(ordered, 2)).toEqual({ sortOrder: 455, updates: [] });
  });

  it('appends after the last status', () => {
    expect(insertionPlan(ordered, 3)).toEqual({ sortOrder: 910, updates: [] });
  });

  it('renumbers when there is no room before the first status', () => {
    const plan = insertionPlan(ordered, 0);
    expect(plan.sortOrder).toBe(0);
    expect(plan.updates.map((u) => [u.status.id, u.sortOrder])).toEqual([
      ['1', 10],
      ['2', 20],
      ['3', 30],
    ]);
  });
});

describe('displayStatusName', () => {
  const t = (key: string): string => `<${key}>`;

  it('translates an unrenamed default status, whatever its case or spacing', () => {
    expect(displayStatusName('To Do', t)).toBe('<flowStatus.todo>');
    expect(displayStatusName(' in progress ', t)).toBe('<flowStatus.inProgress>');
    expect(displayStatusName('CANCELLED', t)).toBe('<flowStatus.cancelled>');
  });

  it('shows a renamed or custom status exactly as typed', () => {
    expect(displayStatusName('Design', t)).toBe('Design');
    expect(displayStatusName('Done (client)', t)).toBe('Done (client)');
  });
});
