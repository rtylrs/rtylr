import { describe, it, expect } from 'vitest';

import {
  archivedTasks,
  activeTasks,
  archivedTaskCount,
  archivedProjectList,
  activeProjectList,
  hasArchivedTasks,
  archivedTasksForProject,
} from './taskArchiveUtils';
import type { Task, Project } from '../stores/flowStore';

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
    archived: false,
    ...overrides,
  } as Task;
}

function makeProject(overrides: Partial<Project>): Project {
  return {
    id: 'p1',
    name: 'Project',
    color: '#3b82f6',
    taskCount: 0,
    completedCount: 0,
    memberAvatars: [],
    health: 'on_track',
    progress: 0,
    startDate: '',
    targetDate: '',
    lead: '',
    archived: false,
    ...overrides,
  } as Project;
}

describe('archivedTasks', () => {
  it('returns archived tasks', () => {
    const tasks = [makeTask({ archived: true }), makeTask({ archived: false })];
    expect(archivedTasks(tasks).length).toBe(1);
  });
});

describe('activeTasks', () => {
  it('returns non-archived tasks', () => {
    const tasks = [makeTask({ archived: true }), makeTask({ archived: false })];
    expect(activeTasks(tasks).length).toBe(1);
  });
});

describe('archivedTaskCount', () => {
  it('counts archived tasks', () => {
    const tasks = [makeTask({ archived: true }), makeTask({ archived: true }), makeTask({})];
    expect(archivedTaskCount(tasks)).toBe(2);
  });
});

describe('archivedProjectList', () => {
  it('returns archived projects', () => {
    const projects = [makeProject({ archived: true }), makeProject({ id: 'p2' })];
    expect(archivedProjectList(projects).length).toBe(1);
  });
});

describe('activeProjectList', () => {
  it('returns non-archived projects', () => {
    const projects = [makeProject({ archived: true }), makeProject({ id: 'p2', archived: false })];
    expect(activeProjectList(projects).length).toBe(1);
  });
});

describe('hasArchivedTasks', () => {
  it('returns true when archived tasks exist', () => {
    expect(hasArchivedTasks([makeTask({ archived: true })])).toBe(true);
  });

  it('returns false when none archived', () => {
    expect(hasArchivedTasks([makeTask({ archived: false })])).toBe(false);
  });
});

describe('archivedTasksForProject', () => {
  it('returns archived tasks for a project', () => {
    const tasks = [
      makeTask({ archived: true, projectId: 'p1' }),
      makeTask({ id: 't2', archived: true, projectId: 'p2' }),
      makeTask({ id: 't3', archived: false, projectId: 'p1' }),
    ];
    expect(archivedTasksForProject(tasks, 'p1').length).toBe(1);
  });
});
