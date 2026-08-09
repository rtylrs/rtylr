import { describe, it, expect } from 'vitest';

import {
  taskPassesFilter,
  applyFilter,
  mergeFilters,
  hasActiveFilter,
  activeFilterCount,
  emptyFilter,
} from './filterUtils';
import type { Task } from '../stores/flowStore';

function makeTask(overrides: Partial<Task>): Task {
  return {
    id: 't1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    assigneeId: 'u1',
    projectId: 'p1',
    workspaceId: 'w1',
    dueDate: '2026-06-30',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    labelIds: ['l1'],
    estimate: 2,
    ...overrides,
  } as Task;
}

describe('taskPassesFilter', () => {
  it('passes an empty filter', () => {
    expect(taskPassesFilter(makeTask({}), {})).toBe(true);
  });

  it('filters by status', () => {
    expect(taskPassesFilter(makeTask({ status: 'done' }), { status: ['todo'] })).toBe(false);
    expect(taskPassesFilter(makeTask({ status: 'todo' }), { status: ['todo'] })).toBe(true);
  });

  it('filters by priority', () => {
    expect(taskPassesFilter(makeTask({ priority: 'urgent' }), { priority: ['low'] })).toBe(false);
  });

  it('filters by assigneeId', () => {
    expect(taskPassesFilter(makeTask({ assigneeId: 'u2' }), { assigneeId: 'u1' })).toBe(false);
    expect(taskPassesFilter(makeTask({ assigneeId: 'u1' }), { assigneeId: 'u1' })).toBe(true);
  });

  it('filters by projectId', () => {
    expect(taskPassesFilter(makeTask({ projectId: 'p2' }), { projectId: 'p1' })).toBe(false);
  });

  it('filters by labelIds matching any', () => {
    expect(taskPassesFilter(makeTask({ labelIds: ['l2'] }), { labelIds: ['l1'] })).toBe(false);
    expect(taskPassesFilter(makeTask({ labelIds: ['l1', 'l2'] }), { labelIds: ['l1'] })).toBe(true);
  });

  it('filters by hasDueDate', () => {
    expect(taskPassesFilter(makeTask({ dueDate: '' }), { hasDueDate: true })).toBe(false);
    expect(taskPassesFilter(makeTask({ dueDate: '2026-06-30' }), { hasDueDate: true })).toBe(true);
  });

  it('filters by hasEstimate', () => {
    expect(taskPassesFilter(makeTask({ estimate: undefined }), { hasEstimate: true })).toBe(false);
    expect(taskPassesFilter(makeTask({ estimate: 2 }), { hasEstimate: true })).toBe(true);
  });
});

describe('applyFilter', () => {
  it('returns matching tasks', () => {
    const tasks = [makeTask({ id: 'a', status: 'done' }), makeTask({ id: 'b', status: 'todo' })];
    expect(applyFilter(tasks, { status: ['done'] }).map((t) => t.id)).toEqual(['a']);
  });
});

describe('mergeFilters', () => {
  it('overrides base with second filter', () => {
    const merged = mergeFilters({ projectId: 'p1' }, { projectId: 'p2' });
    expect(merged.projectId).toBe('p2');
  });

  it('keeps base fields not in override', () => {
    const merged = mergeFilters({ assigneeId: 'u1' }, { projectId: 'p2' });
    expect(merged.assigneeId).toBe('u1');
  });
});

describe('hasActiveFilter', () => {
  it('returns false for empty filter', () => {
    expect(hasActiveFilter(emptyFilter())).toBe(false);
  });

  it('returns true when a criterion is set', () => {
    expect(hasActiveFilter({ assigneeId: 'u1' })).toBe(true);
    expect(hasActiveFilter({ status: ['done'] })).toBe(true);
  });
});

describe('activeFilterCount', () => {
  it('counts each distinct criterion', () => {
    expect(activeFilterCount({ status: ['done'], priority: ['urgent'], assigneeId: 'u1' })).toBe(3);
  });

  it('returns 0 for empty filter', () => {
    expect(activeFilterCount({})).toBe(0);
  });
});
