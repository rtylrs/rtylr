import { describe, it, expect } from 'vitest';

import {
  dependenciesOf,
  dependentsOf,
  hasDependencies,
  isBlocking,
  allDependenciesSatisfied,
  readyTasks,
  blockedTasks,
} from './taskDependencyUtils';
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

const taskA = makeTask({ id: 'a', status: 'done', dependencies: [] });
const taskB = makeTask({ id: 'b', status: 'in_progress', dependencies: ['a'] });
const taskC = makeTask({ id: 'c', status: 'todo', dependencies: ['b'] });
const taskD = makeTask({ id: 'd', status: 'todo', dependencies: [] });
const tasks = [taskA, taskB, taskC, taskD];

describe('dependenciesOf', () => {
  it('returns the tasks a task depends on', () => {
    expect(dependenciesOf(tasks, 'b').map((t) => t.id)).toEqual(['a']);
  });

  it('returns empty for no dependencies', () => {
    expect(dependenciesOf(tasks, 'd')).toHaveLength(0);
  });

  it('returns empty for unknown task', () => {
    expect(dependenciesOf(tasks, 'z')).toHaveLength(0);
  });
});

describe('dependentsOf', () => {
  it('returns tasks that depend on the given task', () => {
    expect(dependentsOf(tasks, 'a').map((t) => t.id)).toEqual(['b']);
  });

  it('returns empty when no dependents', () => {
    expect(dependentsOf(tasks, 'd')).toHaveLength(0);
  });
});

describe('hasDependencies', () => {
  it('returns true when task has dependencies', () => {
    expect(hasDependencies(taskB)).toBe(true);
  });

  it('returns false when task has no dependencies', () => {
    expect(hasDependencies(taskA)).toBe(false);
  });
});

describe('isBlocking', () => {
  it('returns true when task is a dependency of another', () => {
    expect(isBlocking(tasks, 'a')).toBe(true);
  });

  it('returns false when task is not a dependency', () => {
    expect(isBlocking(tasks, 'd')).toBe(false);
  });
});

describe('allDependenciesSatisfied', () => {
  it('returns true when dependency is done', () => {
    // taskB depends on taskA which is done
    expect(allDependenciesSatisfied(tasks, 'b')).toBe(true);
  });

  it('returns false when dependency is not done', () => {
    // taskC depends on taskB which is in_progress
    expect(allDependenciesSatisfied(tasks, 'c')).toBe(false);
  });

  it('returns true for tasks with no dependencies', () => {
    expect(allDependenciesSatisfied(tasks, 'd')).toBe(true);
  });
});

describe('readyTasks', () => {
  it('returns tasks that are not done and have satisfied dependencies', () => {
    const ready = readyTasks(tasks);
    // taskB (deps satisfied: a is done), taskD (no deps)
    expect(ready.map((t) => t.id)).toContain('b');
    expect(ready.map((t) => t.id)).toContain('d');
    expect(ready.map((t) => t.id)).not.toContain('c');
  });
});

describe('blockedTasks', () => {
  it('returns tasks with unmet dependencies', () => {
    const blocked = blockedTasks(tasks);
    // taskC depends on taskB which is not done
    expect(blocked.map((t) => t.id)).toEqual(['c']);
  });
});
