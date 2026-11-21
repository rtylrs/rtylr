import { describe, it, expect } from 'vitest';
import type { Task } from '../stores/flowStore';
import {
  NO_EPIC,
  epicKeyOf,
  epicProgress,
  epicsIn,
  parentCandidates,
  toTaskType,
} from './taskHierarchy';

function task(overrides: Partial<Task> = {}): Task {
  return {
    id: '1',
    projectId: 'p1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    assigneeId: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-11-01',
    updatedAt: '2026-11-01',
    identifier: 'T-1',
    taskType: 'task',
    ...overrides,
  };
}

describe('toTaskType', () => {
  it('reads the four types and treats anything else as a task', () => {
    expect(toTaskType('epic')).toBe('epic');
    expect(toTaskType('bug')).toBe('bug');
    expect(toTaskType(undefined)).toBe('task');
    expect(toTaskType('story')).toBe('task');
  });
});

describe('epicKeyOf', () => {
  it('puts an epic in its own group and a task in its epic', () => {
    expect(epicKeyOf(task({ id: '9', taskType: 'epic' }))).toBe('9');
    expect(epicKeyOf(task({ epicId: '9' }))).toBe('9');
    expect(epicKeyOf(task())).toBe(NO_EPIC);
  });
});

describe('epicsIn', () => {
  it('lists the epics in list order with their colour', () => {
    const tasks = [
      task({ id: '1' }),
      task({ id: '2', taskType: 'epic', title: 'Payments', epicColor: '#22c55e' }),
      task({ id: '3', taskType: 'epic', title: 'Onboarding' }),
    ];
    expect(epicsIn(tasks)).toEqual([
      { id: '2', title: 'Payments', color: '#22c55e' },
      { id: '3', title: 'Onboarding', color: undefined },
    ]);
  });
});

describe('parentCandidates', () => {
  const epic = task({ id: 'e', taskType: 'epic' });
  const other = task({ id: 'o' });
  const subtask = task({ id: 's', parentId: 'o' });
  const elsewhere = task({ id: 'x', projectId: 'p2' });
  const all = [epic, other, subtask, elsewhere];

  it('offers epics and top-level tasks of the same project, never itself', () => {
    const ids = parentCandidates(task({ id: 't' }), [...all, task({ id: 't' })]).map((c) => c.id);
    expect(ids).toEqual(['e', 'o']);
  });

  it('offers nothing to an epic, which takes no parent', () => {
    expect(parentCandidates(task({ id: 'e2', taskType: 'epic' }), all)).toEqual([]);
  });

  it('offers only epics to a task that has subtasks', () => {
    expect(parentCandidates(task({ id: 't', subtaskCount: 2 }), all).map((c) => c.id)).toEqual([
      'e',
    ]);
  });

  it('leaves out where the task already is', () => {
    expect(parentCandidates(task({ id: 't', epicId: 'e' }), all).map((c) => c.id)).toEqual(['o']);
    // A subtask may move straight under its parent's epic.
    expect(
      parentCandidates(task({ id: 't', parentId: 'o', epicId: 'e' }), all).map((c) => c.id)
    ).toEqual(['e']);
  });
});

describe('epicProgress', () => {
  const rollup = {
    epicId: '9',
    totalTasks: 5,
    doneTasks: 2,
    cancelledTasks: 1,
    estimateMinutes: 600,
    doneEstimateMinutes: 150,
    loggedMinutes: 720,
  };

  it('counts done tasks against those still in play', () => {
    expect(epicProgress(rollup, 'count')).toEqual({ done: 2, total: 4, percent: 50, over: false });
  });

  it('weights by estimate', () => {
    expect(epicProgress(rollup, 'estimate').percent).toBe(25);
  });

  it('caps hours at a full bar and says it ran over', () => {
    expect(epicProgress(rollup, 'hours')).toEqual({
      done: 720,
      total: 600,
      percent: 100,
      over: true,
    });
  });

  it('reads an empty epic, whose zero counts are omitted on the wire, as 0%', () => {
    expect(epicProgress({ epicId: '9' }, 'count')).toEqual({
      done: 0,
      total: 0,
      percent: 0,
      over: false,
    });
  });
});
