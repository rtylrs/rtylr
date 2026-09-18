import { describe, it, expect, beforeEach, afterEach, beforeAll, afterAll, vi } from 'vitest';

import {
  bucketForMyTask,
  MY_TASK_BUCKET_ORDER,
  MY_TASK_BUCKET_COLLAPSED_BY_DEFAULT,
  MY_TASK_BUCKET_LABEL_KEY,
} from './myTasksBuckets';
import type { Task } from '../stores/flowStore';

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
    createdAt: '2026-05-02T09:00:00.000Z',
    updatedAt: '2026-05-02T09:00:00.000Z',
    ...overrides,
  } as Task;
}

describe('bucketForMyTask', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // Saturday 2026-05-02, midday, so the local day matches the UTC day and the
    // buckets below are about the rule rather than about the offset.
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('gives an overdue task its own bucket instead of folding it into Today', () => {
    // The defect rtylr#1077 was raised for. Six overdue tasks were spread
    // through a section headed "Today", so a fifteen-day-overdue item had no
    // promotion of any kind.
    expect(bucketForMyTask(baseTask({ dueDate: '2026-04-17' }))).toBe('late');
    expect(bucketForMyTask(baseTask({ dueDate: '2026-05-01' }))).toBe('late');
  });

  it('separates today from late by a single day', () => {
    expect(bucketForMyTask(baseTask({ dueDate: '2026-05-01' }))).toBe('late');
    expect(bucketForMyTask(baseTask({ dueDate: '2026-05-02' }))).toBe('today');
  });

  it('puts the next seven days in this week and the rest in later', () => {
    expect(bucketForMyTask(baseTask({ dueDate: '2026-05-03' }))).toBe('this_week');
    expect(bucketForMyTask(baseTask({ dueDate: '2026-05-09' }))).toBe('this_week');
    expect(bucketForMyTask(baseTask({ dueDate: '2026-05-10' }))).toBe('later');
  });

  it('buckets a task with no due date by that fact, not by when it arrived', () => {
    // The old bucket keyed off createdAt and answered "what landed on me
    // lately", which is not a position on an urgency scale: a task assigned this
    // morning and due in a month sorted above one due tomorrow.
    expect(bucketForMyTask(baseTask({ dueDate: '' }))).toBe('no_date');
    expect(bucketForMyTask(baseTask({ dueDate: '', createdAt: '2026-01-01T00:00:00Z' }))).toBe(
      'no_date'
    );
    expect(bucketForMyTask(baseTask({ dueDate: undefined }))).toBe('no_date');
  });

  it('reads a timestamped due date as the day it names', () => {
    expect(bucketForMyTask(baseTask({ dueDate: '2026-05-02T23:59:00Z' }))).toBe('today');
  });
});

describe('the buckets as a screen', () => {
  it('orders them by urgency, which is the reading order', () => {
    expect(MY_TASK_BUCKET_ORDER).toEqual(['late', 'today', 'this_week', 'later', 'no_date']);
  });

  it('never opens Late or Today collapsed', () => {
    // The point of the screen is that overdue work is visible without an
    // interaction, so this is a rule and not a default.
    expect(MY_TASK_BUCKET_COLLAPSED_BY_DEFAULT.has('late')).toBe(false);
    expect(MY_TASK_BUCKET_COLLAPSED_BY_DEFAULT.has('today')).toBe(false);
  });

  it('gives every bucket a translation key, not English', () => {
    for (const bucket of MY_TASK_BUCKET_ORDER) {
      expect(MY_TASK_BUCKET_LABEL_KEY[bucket]).toMatch(/^flow\.myIssues\.buckets\./);
    }
  });
});

describe('west of Greenwich, where the previous version lost a day', () => {
  /**
   * The old implementation did `startOfDay(new Date(task.dueDate))`, parsing a
   * bare `YYYY-MM-DD` as UTC midnight and reading it back locally. Pinned to a
   * western zone because from Beirut, where this was written, the broken idiom
   * and the correct one agree.
   */
  const ORIGINAL_TZ = process.env.TZ;
  beforeAll(() => {
    process.env.TZ = 'America/New_York';
    vi.useFakeTimers();
    // 08:00 in New York, so the local day is the 2nd and the UTC day is too.
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });
  afterAll(() => {
    vi.useRealTimers();
    // Assigning undefined sets TZ to the string "undefined", which Node rejects
    // and falls back to UTC, silently moving every later file to another zone.
    if (ORIGINAL_TZ === undefined) delete process.env.TZ;
    else process.env.TZ = ORIGINAL_TZ;
  });

  it('calls a task due today today, not late', () => {
    expect(bucketForMyTask(baseTask({ dueDate: '2026-05-02' }))).toBe('today');
  });

  it('shows the idiom that produced the wrong answer', () => {
    // `new Date('2026-05-02')` is 20:00 on 1 May in New York, so a start-of-day
    // comparison put a task due today in the past.
    expect(new Date('2026-05-02').getDate()).toBe(1);
    expect(new Date(2026, 4, 2).getDate()).toBe(2);
  });
});
