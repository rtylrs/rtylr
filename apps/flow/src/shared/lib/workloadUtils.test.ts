import { describe, it, expect } from 'vitest';

import {
  capacityUtilizationPercent,
  loadTier,
  loadTierColor,
  totalEstimatedHoursForMember,
  openTaskCountForMember,
  sortMembersByLoad,
} from './workloadUtils';
import type { Task, Member } from '../stores/flowStore';

function makeTask(overrides: Partial<Task>): Task {
  return {
    id: 't1',
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
    identifier: 't1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    ...overrides,
  } as Task;
}

function makeMember(overrides: Partial<Member>): Member {
  return {
    id: 'm1',
    name: 'Alice',
    email: 'alice@example.com',
    initials: 'AL',
    role: 'member',
    capacity: 40,
    ...overrides,
  } as Member;
}

// ---------------------------------------------------------------------------
// capacityUtilizationPercent
// ---------------------------------------------------------------------------

describe('capacityUtilizationPercent', () => {
  it('returns 0 when capacity is 0', () => {
    expect(capacityUtilizationPercent(10, 0)).toBe(0);
  });

  it('returns 100 when fully loaded', () => {
    expect(capacityUtilizationPercent(40, 40)).toBe(100);
  });

  it('returns 50 when half loaded', () => {
    expect(capacityUtilizationPercent(20, 40)).toBe(50);
  });

  it('returns over 100 when overloaded', () => {
    expect(capacityUtilizationPercent(50, 40)).toBe(125);
  });
});

// ---------------------------------------------------------------------------
// loadTier
// ---------------------------------------------------------------------------

describe('loadTier', () => {
  it('returns "over" for > 100%', () => {
    expect(loadTier(110)).toBe('over');
    expect(loadTier(101)).toBe('over');
  });

  it('returns "near" for 80-100%', () => {
    expect(loadTier(100)).toBe('near');
    expect(loadTier(80)).toBe('near');
  });

  it('returns "ok" for < 80%', () => {
    expect(loadTier(79)).toBe('ok');
    expect(loadTier(0)).toBe('ok');
  });
});

// ---------------------------------------------------------------------------
// loadTierColor
// ---------------------------------------------------------------------------

describe('loadTierColor', () => {
  it('returns a usable colour value for over', () => {
    // Not "is it a hex literal". That was the shape of the old assertion, and
    // it is exactly the constraint #1073 removes: these tones are tokens now,
    // so they can hold a different value per mode. What matters is that the
    // function yields something CSS can paint.
    expect(loadTierColor('over')).toMatch(/^(#[0-9a-f]{3,8}|var\(--[\w-]+\)|rgba?\()/i);
  });

  it('returns different colors for each tier', () => {
    const colors = new Set([loadTierColor('over'), loadTierColor('near'), loadTierColor('ok')]);
    expect(colors.size).toBe(3);
  });
});

// ---------------------------------------------------------------------------
// totalEstimatedHoursForMember
// ---------------------------------------------------------------------------

describe('totalEstimatedHoursForMember', () => {
  it('sums estimate hours for open tasks assigned to a member', () => {
    const tasks = [
      makeTask({ assigneeId: 'u1', status: 'todo', estimate: 3 }),
      makeTask({ assigneeId: 'u1', status: 'in_progress', estimate: 5 }),
      makeTask({ assigneeId: 'u1', status: 'done', estimate: 8 }),
      makeTask({ assigneeId: 'u2', status: 'todo', estimate: 4 }),
    ];
    expect(totalEstimatedHoursForMember(tasks, 'u1')).toBe(8);
  });

  it('returns 0 when no tasks are assigned', () => {
    expect(totalEstimatedHoursForMember([], 'u1')).toBe(0);
  });

  it('handles tasks with no estimate', () => {
    const tasks = [makeTask({ assigneeId: 'u1', status: 'todo', estimate: undefined })];
    expect(totalEstimatedHoursForMember(tasks, 'u1')).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// openTaskCountForMember
// ---------------------------------------------------------------------------

describe('openTaskCountForMember', () => {
  it('counts open tasks for member', () => {
    const tasks = [
      makeTask({ assigneeId: 'u1', status: 'todo' }),
      makeTask({ assigneeId: 'u1', status: 'in_progress' }),
      makeTask({ assigneeId: 'u1', status: 'done' }),
      makeTask({ assigneeId: 'u1', status: 'cancelled' }),
    ];
    expect(openTaskCountForMember(tasks, 'u1')).toBe(2);
  });
});

// ---------------------------------------------------------------------------
// sortMembersByLoad
// ---------------------------------------------------------------------------

describe('sortMembersByLoad', () => {
  it('sorts members with highest load first', () => {
    const members = [
      { ...makeMember({ id: 'a', capacity: 40 }), assignedHours: 10 },
      { ...makeMember({ id: 'b', capacity: 40 }), assignedHours: 40 },
      { ...makeMember({ id: 'c', capacity: 40 }), assignedHours: 30 },
    ];
    const sorted = sortMembersByLoad(members);
    expect(sorted.map((m) => m.id)).toEqual(['b', 'c', 'a']);
  });

  it('does not mutate the input', () => {
    const members = [
      { ...makeMember({ id: 'a' }), assignedHours: 40 },
      { ...makeMember({ id: 'b' }), assignedHours: 0 },
    ];
    const original = [...members];
    sortMembersByLoad(members);
    expect(members).toEqual(original);
  });
});
