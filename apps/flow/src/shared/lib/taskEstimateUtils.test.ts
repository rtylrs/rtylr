import { describe, it, expect } from 'vitest';

import {
  hasEstimate,
  estimatedTasks,
  unestimatedTaskList,
  estimateHours,
  totalEstimate,
  averageEstimate,
  sortByEstimateDesc,
  tasksWithEstimateRange,
  estimateCoverage,
} from './taskEstimateUtils';
import type { Task } from '../stores/flowStore';

function makeTask(overrides: Partial<Task>): Task {
  return {
    id: 't1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    assigneeId: '',
    projectId: 'p1',
    workspaceId: 'w1',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    ...overrides,
  } as Task;
}

describe('hasEstimate', () => {
  it('returns true for positive estimate', () => {
    expect(hasEstimate(makeTask({ estimate: 3 }))).toBe(true);
  });

  it('returns false for undefined estimate', () => {
    expect(hasEstimate(makeTask({ estimate: undefined }))).toBe(false);
  });

  it('returns false for 0 estimate', () => {
    expect(hasEstimate(makeTask({ estimate: 0 }))).toBe(false);
  });
});

describe('estimatedTasks', () => {
  it('filters tasks with estimates', () => {
    const tasks = [makeTask({ estimate: 2 }), makeTask({ estimate: undefined })];
    expect(estimatedTasks(tasks).length).toBe(1);
  });
});

describe('unestimatedTaskList', () => {
  it('filters tasks without estimates', () => {
    const tasks = [makeTask({ estimate: 2 }), makeTask({ estimate: undefined })];
    expect(unestimatedTaskList(tasks).length).toBe(1);
  });
});

describe('estimateHours', () => {
  it('returns 0 for undefined', () => {
    expect(estimateHours(makeTask({ estimate: undefined }))).toBe(0);
  });

  it('returns estimate when set', () => {
    expect(estimateHours(makeTask({ estimate: 4 }))).toBe(4);
  });
});

describe('totalEstimate', () => {
  it('sums all estimates', () => {
    const tasks = [makeTask({ estimate: 3 }), makeTask({ estimate: 5 }), makeTask({})];
    expect(totalEstimate(tasks)).toBe(8);
  });
});

describe('averageEstimate', () => {
  it('averages tasks with estimates', () => {
    const tasks = [makeTask({ estimate: 2 }), makeTask({ estimate: 4 }), makeTask({})];
    expect(averageEstimate(tasks)).toBe(3);
  });

  it('returns 0 for no estimates', () => {
    expect(averageEstimate([])).toBe(0);
  });
});

describe('sortByEstimateDesc', () => {
  it('sorts largest estimate first', () => {
    const tasks = [makeTask({ id: 'a', estimate: 2 }), makeTask({ id: 'b', estimate: 8 })];
    expect(sortByEstimateDesc(tasks)[0].id).toBe('b');
  });
});

describe('tasksWithEstimateRange', () => {
  it('returns tasks within range', () => {
    const tasks = [makeTask({ estimate: 1 }), makeTask({ estimate: 3 }), makeTask({ estimate: 8 })];
    expect(tasksWithEstimateRange(tasks, 2, 5).length).toBe(1);
  });
});

describe('estimateCoverage', () => {
  it('returns percentage with estimates', () => {
    const tasks = [makeTask({ estimate: 2 }), makeTask({ estimate: undefined })];
    expect(estimateCoverage(tasks)).toBe(50);
  });

  it('returns 0 for empty list', () => {
    expect(estimateCoverage([])).toBe(0);
  });
});
