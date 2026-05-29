import { describe, it, expect } from 'vitest';

import {
  tasksDueToday,
  tasksDueWithin,
  overdueTaskList,
  tasksNoDueDate,
  tasksDueThisWeek,
  dueDateLabel,
} from './taskDueUtils';
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

const TODAY = '2026-06-15';

describe('tasksDueToday', () => {
  it('returns tasks due exactly today', () => {
    const tasks = [makeTask({ dueDate: TODAY }), makeTask({ dueDate: '2026-06-16' })];
    expect(tasksDueToday(tasks, TODAY).length).toBe(1);
  });
});

describe('tasksDueWithin', () => {
  it('returns tasks due within N days', () => {
    const tasks = [
      makeTask({ dueDate: '2026-06-17', status: 'todo' }),
      makeTask({ dueDate: '2026-06-30', status: 'todo' }),
      makeTask({ dueDate: '2026-06-10', status: 'todo' }),
    ];
    expect(tasksDueWithin(tasks, TODAY, 5).length).toBe(1);
  });

  it('excludes done and cancelled tasks', () => {
    const tasks = [makeTask({ dueDate: '2026-06-16', status: 'done' })];
    expect(tasksDueWithin(tasks, TODAY, 5).length).toBe(0);
  });
});

describe('overdueTaskList', () => {
  it('returns past-due open tasks', () => {
    const tasks = [
      makeTask({ dueDate: '2026-01-01', status: 'todo' }),
      makeTask({ dueDate: '2026-01-01', status: 'done' }),
    ];
    expect(overdueTaskList(tasks, TODAY).length).toBe(1);
  });
});

describe('tasksNoDueDate', () => {
  it('returns open tasks with no due date', () => {
    const tasks = [makeTask({ dueDate: '' }), makeTask({ dueDate: TODAY })];
    expect(tasksNoDueDate(tasks).length).toBe(1);
  });

  it('excludes done tasks', () => {
    const tasks = [makeTask({ dueDate: '', status: 'done' })];
    expect(tasksNoDueDate(tasks).length).toBe(0);
  });
});

describe('tasksDueThisWeek', () => {
  it('returns tasks in the current week', () => {
    // 2026-06-15 is Monday, so this week is 2026-06-15 to 2026-06-21
    const tasks = [makeTask({ dueDate: '2026-06-17' }), makeTask({ dueDate: '2026-06-22' })];
    expect(tasksDueThisWeek(tasks, TODAY).length).toBe(1);
  });
});

describe('dueDateLabel', () => {
  it('returns No due date for empty', () => {
    expect(dueDateLabel(makeTask({ dueDate: '' }), TODAY)).toBe('No due date');
  });

  it('returns Due today', () => {
    expect(dueDateLabel(makeTask({ dueDate: TODAY }), TODAY)).toBe('Due today');
  });

  it('returns Due tomorrow', () => {
    expect(dueDateLabel(makeTask({ dueDate: '2026-06-16' }), TODAY)).toBe('Due tomorrow');
  });

  it('returns Overdue', () => {
    expect(dueDateLabel(makeTask({ dueDate: '2026-01-01' }), TODAY)).toBe('Overdue');
  });

  it('returns days label for near future', () => {
    expect(dueDateLabel(makeTask({ dueDate: '2026-06-20' }), TODAY)).toBe('Due in 5 days');
  });
});
