import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

import { bucketForMyTask } from './myTasksBuckets';
import type { Task } from '../stores/flowStore';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function baseTask(overrides: Partial<Task> = {}): Task {
  return {
    id: '1',
    projectId: 'p1',
    workspaceId: 'w1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    assigneeId: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'P',
    identifier: '1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    ...overrides,
  } as Task;
}

// ---------------------------------------------------------------------------
// Tests - pin "now" so relative bucket logic is deterministic
// ---------------------------------------------------------------------------

describe('bucketForMyTask', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // Saturday 2026-05-02, noon UTC
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('no dueDate + created this week → recently_assigned', () => {
    const task = baseTask({
      dueDate: '',
      createdAt: new Date('2026-04-28T10:00:00.000Z').toISOString(), // 4 days ago
    });
    expect(bucketForMyTask(task)).toBe('recently_assigned');
  });

  it('no dueDate + created more than 7 days ago → later', () => {
    const task = baseTask({
      dueDate: '',
      createdAt: new Date('2026-04-20T10:00:00.000Z').toISOString(), // 12 days ago
    });
    expect(bucketForMyTask(task)).toBe('later');
  });

  it('overdue dueDate (in the past) → today', () => {
    const task = baseTask({ dueDate: '2026-04-25' }); // 7 days ago
    expect(bucketForMyTask(task)).toBe('today');
  });

  it('dueDate is today → today', () => {
    const task = baseTask({ dueDate: '2026-05-02' });
    expect(bucketForMyTask(task)).toBe('today');
  });

  it('dueDate within 7 days (not today) → upcoming', () => {
    const task = baseTask({ dueDate: '2026-05-06' }); // 4 days from now
    expect(bucketForMyTask(task)).toBe('upcoming');
  });

  it('dueDate beyond 7 days → later', () => {
    const task = baseTask({ dueDate: '2026-05-30' });
    expect(bucketForMyTask(task)).toBe('later');
  });
});
