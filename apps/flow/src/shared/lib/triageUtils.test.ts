import { describe, it, expect } from 'vitest';

import {
  backlogTasks,
  unassignedTasks,
  tasksMissingDueDate,
  unprioritizedTasks,
  unestimatedTasks,
  triageNeeded,
  triageScore,
  sortByTriageScore,
} from './triageUtils';
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
    dueDate: '2026-12-31',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    estimate: 2,
    ...overrides,
  } as Task;
}

describe('backlogTasks', () => {
  it('returns backlog tasks only', () => {
    const tasks = [makeTask({ status: 'backlog' }), makeTask({ status: 'todo' })];
    expect(backlogTasks(tasks).length).toBe(1);
  });
});

describe('unassignedTasks', () => {
  it('returns tasks without an assignee', () => {
    const tasks = [makeTask({ assigneeId: '' }), makeTask({ assigneeId: 'u1' })];
    expect(unassignedTasks(tasks).length).toBe(1);
  });
});

describe('tasksMissingDueDate', () => {
  it('returns tasks without due date', () => {
    const tasks = [makeTask({ dueDate: '' }), makeTask({ dueDate: '2026-06-30' })];
    expect(tasksMissingDueDate(tasks).length).toBe(1);
  });
});

describe('unprioritizedTasks', () => {
  it('returns tasks with no_priority', () => {
    const tasks = [makeTask({ priority: 'no_priority' }), makeTask({ priority: 'high' })];
    expect(unprioritizedTasks(tasks).length).toBe(1);
  });
});

describe('unestimatedTasks', () => {
  it('returns tasks without estimate', () => {
    const tasks = [makeTask({ estimate: undefined }), makeTask({ estimate: 3 })];
    expect(unestimatedTasks(tasks).length).toBe(1);
  });
});

describe('triageNeeded', () => {
  it('includes backlog tasks', () => {
    expect(triageNeeded([makeTask({ status: 'backlog' })]).length).toBe(1);
  });

  it('excludes done and cancelled tasks', () => {
    expect(triageNeeded([makeTask({ status: 'done', assigneeId: '' })]).length).toBe(0);
    expect(triageNeeded([makeTask({ status: 'cancelled', dueDate: '' })]).length).toBe(0);
  });

  it('includes tasks missing assignee, due date, or priority', () => {
    expect(triageNeeded([makeTask({ assigneeId: '' })]).length).toBe(1);
    expect(triageNeeded([makeTask({ dueDate: '' })]).length).toBe(1);
    expect(triageNeeded([makeTask({ priority: 'no_priority' })]).length).toBe(1);
  });
});

describe('triageScore', () => {
  it('scores 0 for a fully filled task', () => {
    expect(triageScore(makeTask({}))).toBe(0);
  });

  it('scores 4 for a completely empty task', () => {
    expect(
      triageScore(
        makeTask({ status: 'backlog', assigneeId: '', dueDate: '', priority: 'no_priority' })
      )
    ).toBe(4);
  });

  it('scores 1 for each missing field', () => {
    expect(triageScore(makeTask({ assigneeId: '' }))).toBe(1);
    expect(triageScore(makeTask({ dueDate: '' }))).toBe(1);
  });
});

describe('sortByTriageScore', () => {
  it('sorts highest score first', () => {
    const tasks = [
      makeTask({ id: 'a' }),
      makeTask({
        id: 'b',
        status: 'backlog',
        assigneeId: '',
        dueDate: '',
        priority: 'no_priority',
      }),
    ];
    expect(sortByTriageScore(tasks)[0].id).toBe('b');
  });
});
