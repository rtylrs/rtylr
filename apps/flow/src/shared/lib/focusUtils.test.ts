import { describe, it, expect } from 'vitest';

import {
  focusPinnedTasks,
  tasksDueToday,
  focusTodayTasks,
  waitingTasks,
  activeFocusTasks,
  tasksWithinEffortBudget,
  tasksWithDecisionNote,
} from './focusUtils';
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

describe('focusPinnedTasks', () => {
  it('returns tasks with focusPin=true', () => {
    const tasks = [makeTask({ id: 'a', focusPin: true }), makeTask({ id: 'b', focusPin: false })];
    expect(focusPinnedTasks(tasks).map((t) => t.id)).toEqual(['a']);
  });
});

describe('tasksDueToday', () => {
  it('returns tasks due on the given date', () => {
    const tasks = [
      makeTask({ id: 'a', dueDate: '2026-05-29' }),
      makeTask({ id: 'b', dueDate: '2026-06-01' }),
    ];
    expect(tasksDueToday(tasks, '2026-05-29').map((t) => t.id)).toEqual(['a']);
  });
});

describe('focusTodayTasks', () => {
  it('includes tasks due today and pinned tasks without duplicates', () => {
    const tasks = [
      makeTask({ id: 'a', dueDate: '2026-05-29', focusPin: false }),
      makeTask({ id: 'b', dueDate: '2026-06-01', focusPin: true }),
      makeTask({ id: 'c', dueDate: '2026-05-29', focusPin: true }),
    ];
    const result = focusTodayTasks(tasks, '2026-05-29');
    const ids = result.map((t) => t.id);
    expect(ids).toContain('a');
    expect(ids).toContain('b');
    expect(ids).toContain('c');
    expect(new Set(ids).size).toBe(ids.length);
  });
});

describe('waitingTasks', () => {
  it('returns tasks with a waitingOn value', () => {
    const tasks = [makeTask({ id: 'a', waitingOn: 'Alice' }), makeTask({ id: 'b' })];
    expect(waitingTasks(tasks).map((t) => t.id)).toEqual(['a']);
  });
});

describe('activeFocusTasks', () => {
  it('returns in_progress tasks not waiting on anything', () => {
    const tasks = [
      makeTask({ id: 'a', status: 'in_progress', waitingOn: undefined }),
      makeTask({ id: 'b', status: 'in_progress', waitingOn: 'Someone' }),
      makeTask({ id: 'c', status: 'todo' }),
    ];
    expect(activeFocusTasks(tasks).map((t) => t.id)).toEqual(['a']);
  });
});

describe('tasksWithinEffortBudget', () => {
  it('returns incomplete tasks within effort budget', () => {
    const tasks = [
      makeTask({ id: 'a', effortMinutes: 15, status: 'todo' }),
      makeTask({ id: 'b', effortMinutes: 45, status: 'todo' }),
      makeTask({ id: 'c', effortMinutes: 15, status: 'done' }),
    ];
    expect(tasksWithinEffortBudget(tasks, 30).map((t) => t.id)).toEqual(['a']);
  });
});

describe('tasksWithDecisionNote', () => {
  it('returns tasks with a decision note', () => {
    const tasks = [makeTask({ id: 'a', decisionNote: 'Use approach X' }), makeTask({ id: 'b' })];
    expect(tasksWithDecisionNote(tasks).map((t) => t.id)).toEqual(['a']);
  });
});
