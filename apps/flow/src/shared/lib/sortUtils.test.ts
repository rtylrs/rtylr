import { describe, it, expect } from 'vitest';

import { sortTasks, sortTasksMultiKey } from './sortUtils';
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

describe('sortTasks by priority', () => {
  it('sorts urgent first in asc direction', () => {
    const tasks = [
      makeTask({ id: 'a', priority: 'low' }),
      makeTask({ id: 'b', priority: 'urgent' }),
    ];
    expect(sortTasks(tasks, 'priority', 'asc')[0].id).toBe('b');
  });

  it('sorts low first in desc direction', () => {
    const tasks = [
      makeTask({ id: 'a', priority: 'low' }),
      makeTask({ id: 'b', priority: 'urgent' }),
    ];
    expect(sortTasks(tasks, 'priority', 'desc')[0].id).toBe('a');
  });
});

describe('sortTasks by status', () => {
  it('sorts backlog first in asc direction', () => {
    const tasks = [makeTask({ id: 'a', status: 'done' }), makeTask({ id: 'b', status: 'backlog' })];
    expect(sortTasks(tasks, 'status', 'asc')[0].id).toBe('b');
  });
});

describe('sortTasks by dueDate', () => {
  it('sorts earlier dates first', () => {
    const tasks = [
      makeTask({ id: 'a', dueDate: '2026-12-31' }),
      makeTask({ id: 'b', dueDate: '2026-01-01' }),
    ];
    expect(sortTasks(tasks, 'dueDate', 'asc')[0].id).toBe('b');
  });

  it('places empty due dates last', () => {
    const tasks = [
      makeTask({ id: 'a', dueDate: '' }),
      makeTask({ id: 'b', dueDate: '2026-01-01' }),
    ];
    expect(sortTasks(tasks, 'dueDate', 'asc')[0].id).toBe('b');
  });
});

describe('sortTasks by title', () => {
  it('sorts alphabetically', () => {
    const tasks = [makeTask({ id: 'a', title: 'Zebra' }), makeTask({ id: 'b', title: 'Alpha' })];
    expect(sortTasks(tasks, 'title', 'asc')[0].id).toBe('b');
  });
});

describe('sortTasks by created', () => {
  it('sorts oldest first', () => {
    const tasks = [
      makeTask({ id: 'a', createdAt: '2026-06-01' }),
      makeTask({ id: 'b', createdAt: '2026-01-01' }),
    ];
    expect(sortTasks(tasks, 'created', 'asc')[0].id).toBe('b');
  });
});

describe('sortTasks by updated', () => {
  it('sorts most recently updated first in desc', () => {
    const tasks = [
      makeTask({ id: 'a', updatedAt: '2026-01-01' }),
      makeTask({ id: 'b', updatedAt: '2026-06-01' }),
    ];
    expect(sortTasks(tasks, 'updated', 'desc')[0].id).toBe('b');
  });
});

describe('sortTasksMultiKey', () => {
  it('applies criteria in order', () => {
    const tasks = [
      makeTask({ id: 'a', priority: 'urgent', title: 'Zebra' }),
      makeTask({ id: 'b', priority: 'urgent', title: 'Alpha' }),
      makeTask({ id: 'c', priority: 'low', title: 'Alpha' }),
    ];
    const sorted = sortTasksMultiKey(tasks, [
      { key: 'priority', dir: 'asc' },
      { key: 'title', dir: 'asc' },
    ]);
    expect(sorted.map((t) => t.id)).toEqual(['b', 'a', 'c']);
  });
});
