import { describe, it, expect } from 'vitest';

import {
  buildBoardColumns,
  nonEmptyColumns,
  boardTaskCount,
  sortColumns,
  columnForStatus,
  columnLabel,
} from './boardUtils';
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
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    ...overrides,
  } as Task;
}

describe('buildBoardColumns', () => {
  it('creates a column for each status', () => {
    const tasks = [makeTask({ status: 'todo' }), makeTask({ status: 'done' })];
    const columns = buildBoardColumns(tasks);
    expect(columns.length).toBe(6);
    expect(columns.find((c) => c.status === 'todo')?.tasks.length).toBe(1);
    expect(columns.find((c) => c.status === 'done')?.tasks.length).toBe(1);
  });

  it('returns empty tasks array for columns with no tasks', () => {
    const columns = buildBoardColumns([]);
    expect(columns.every((c) => c.tasks.length === 0)).toBe(true);
  });
});

describe('nonEmptyColumns', () => {
  it('filters out columns with no tasks', () => {
    const columns = buildBoardColumns([makeTask({ status: 'todo' })]);
    expect(nonEmptyColumns(columns).length).toBe(1);
  });
});

describe('boardTaskCount', () => {
  it('sums tasks across all columns', () => {
    const columns = buildBoardColumns([
      makeTask({ status: 'todo' }),
      makeTask({ status: 'done' }),
      makeTask({ status: 'done' }),
    ]);
    expect(boardTaskCount(columns)).toBe(3);
  });
});

describe('sortColumns', () => {
  it('returns columns in status order', () => {
    const columns = buildBoardColumns([]);
    const sorted = sortColumns(columns);
    expect(sorted[0].status).toBe('backlog');
    expect(sorted[sorted.length - 1].status).toBe('cancelled');
  });
});

describe('columnForStatus', () => {
  it('returns the matching column', () => {
    const columns = buildBoardColumns([makeTask({ status: 'in_progress' })]);
    expect(columnForStatus(columns, 'in_progress')?.tasks.length).toBe(1);
  });

  it('returns undefined for missing status', () => {
    expect(columnForStatus([], 'todo')).toBeUndefined();
  });
});

describe('columnLabel', () => {
  it('returns display label', () => {
    expect(columnLabel('in_progress')).toBe('In Progress');
    expect(columnLabel('in_review')).toBe('In Review');
  });
});
