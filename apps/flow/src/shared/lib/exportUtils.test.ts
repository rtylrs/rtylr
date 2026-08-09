import { describe, it, expect } from 'vitest';

import {
  taskToCsvRow,
  taskCsvHeader,
  tasksToCsv,
  projectToCsvRow,
  projectCsvHeader,
  projectsToCsv,
  csvToDataUri,
  exportFilename,
} from './exportUtils';
import type { Task, Project } from '../stores/flowStore';

function makeTask(overrides: Partial<Task>): Task {
  return {
    id: 't1',
    title: 'Fix bug',
    status: 'todo',
    priority: 'medium',
    assignee: 'Alice',
    assigneeInitials: 'A',
    assigneeId: 'u1',
    projectId: 'p1',
    workspaceId: 'w1',
    dueDate: '2026-06-30',
    subtaskCount: 0,
    projectName: 'Auth',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    estimate: 2,
    ...overrides,
  } as Task;
}

function makeProject(overrides: Partial<Project>): Project {
  return {
    id: 'p1',
    workspaceId: 'w1',
    name: 'Auth Project',
    color: '#3b82f6',
    taskCount: 10,
    completedCount: 5,
    memberAvatars: [],
    health: 'on_track',
    progress: 50,
    startDate: '2026-01-01',
    targetDate: '2026-12-31',
    lead: '',
    ...overrides,
  } as Project;
}

describe('taskToCsvRow', () => {
  it('includes identifier and title', () => {
    const row = taskToCsvRow(makeTask({}));
    expect(row).toContain('FLOW-1');
    expect(row).toContain('Fix bug');
  });

  it('escapes double quotes in title', () => {
    const row = taskToCsvRow(makeTask({ title: 'He said "hello"' }));
    expect(row).toContain('""hello""');
  });
});

describe('taskCsvHeader', () => {
  it('returns comma-separated header', () => {
    expect(taskCsvHeader()).toContain('Identifier');
    expect(taskCsvHeader()).toContain('Status');
  });
});

describe('tasksToCsv', () => {
  it('includes header and rows', () => {
    const csv = tasksToCsv([makeTask({})]);
    const lines = csv.split('\n');
    expect(lines[0]).toContain('Identifier');
    expect(lines[1]).toContain('FLOW-1');
  });
});

describe('projectToCsvRow', () => {
  it('includes name and health', () => {
    const row = projectToCsvRow(makeProject({}));
    expect(row).toContain('Auth Project');
    expect(row).toContain('on_track');
  });
});

describe('projectCsvHeader', () => {
  it('includes Name and Health columns', () => {
    expect(projectCsvHeader()).toContain('Name');
    expect(projectCsvHeader()).toContain('Health');
  });
});

describe('projectsToCsv', () => {
  it('includes header and project row', () => {
    const csv = projectsToCsv([makeProject({})]);
    const lines = csv.split('\n');
    expect(lines[0]).toContain('Name');
    expect(lines[1]).toContain('Auth Project');
  });
});

describe('csvToDataUri', () => {
  it('returns a data URI', () => {
    const uri = csvToDataUri('a,b\n1,2');
    expect(uri.startsWith('data:text/csv')).toBe(true);
  });
});

describe('exportFilename', () => {
  it('includes prefix and date', () => {
    const name = exportFilename('tasks');
    expect(name.startsWith('tasks-')).toBe(true);
    expect(name.endsWith('.csv')).toBe(true);
  });

  it('supports custom extension', () => {
    expect(exportFilename('report', 'json').endsWith('.json')).toBe(true);
  });
});
