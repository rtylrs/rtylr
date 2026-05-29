import { describe, it, expect } from 'vitest';

import {
  searchTasks,
  searchProjects,
  searchMembers,
  taskMatchesTokens,
  tokenizeQuery,
  searchTasksMultiToken,
  tasksByIdentifierPrefix,
} from './searchUtils';
import type { Task, Project, Member } from '../stores/flowStore';

function makeTask(overrides: Partial<Task>): Task {
  return {
    id: 't1',
    title: 'Fix login bug',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    assigneeId: '',
    projectId: 'p1',
    workspaceId: 'w1',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Auth',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    ...overrides,
  } as Task;
}

function makeProject(overrides: Partial<Project>): Project {
  return {
    id: 'p1',
    workspaceId: 'w1',
    name: 'Dashboard',
    color: '#3b82f6',
    taskCount: 0,
    completedCount: 0,
    memberAvatars: [],
    health: 'on_track',
    progress: 0,
    startDate: '',
    targetDate: '',
    lead: '',
    ...overrides,
  } as Project;
}

function makeMember(overrides: Partial<Member>): Member {
  return {
    id: 'm1',
    name: 'Alice Smith',
    email: 'alice@example.com',
    initials: 'AS',
    role: 'member',
    capacity: 40,
    ...overrides,
  } as Member;
}

describe('searchTasks', () => {
  it('finds tasks by title substring', () => {
    const tasks = [makeTask({ title: 'Fix login bug' }), makeTask({ title: 'Add user settings' })];
    expect(searchTasks(tasks, 'login').length).toBe(1);
  });

  it('finds tasks by identifier', () => {
    const tasks = [makeTask({ identifier: 'FLOW-42' })];
    expect(searchTasks(tasks, 'FLOW-42').length).toBe(1);
  });

  it('returns all tasks for empty query', () => {
    const tasks = [makeTask({}), makeTask({})];
    expect(searchTasks(tasks, '').length).toBe(2);
  });

  it('is case-insensitive', () => {
    const tasks = [makeTask({ title: 'Fix Login Bug' })];
    expect(searchTasks(tasks, 'login').length).toBe(1);
  });
});

describe('searchProjects', () => {
  it('finds projects by name', () => {
    const projects = [makeProject({ name: 'Dashboard' }), makeProject({ name: 'Settings' })];
    expect(searchProjects(projects, 'dash').length).toBe(1);
  });

  it('returns all for empty query', () => {
    expect(searchProjects([makeProject({})], '').length).toBe(1);
  });
});

describe('searchMembers', () => {
  it('finds by name', () => {
    const members = [
      makeMember({ name: 'Alice' }),
      makeMember({ name: 'Bob', email: 'bob@x.com' }),
    ];
    expect(searchMembers(members, 'alice').length).toBe(1);
  });

  it('finds by email', () => {
    const members = [makeMember({ email: 'alice@example.com' })];
    expect(searchMembers(members, 'example').length).toBe(1);
  });
});

describe('tokenizeQuery', () => {
  it('splits on whitespace and lowercases', () => {
    expect(tokenizeQuery('Fix Login Bug')).toEqual(['fix', 'login', 'bug']);
  });

  it('returns empty for blank string', () => {
    expect(tokenizeQuery('   ')).toEqual([]);
  });
});

describe('taskMatchesTokens', () => {
  it('returns true when all tokens match', () => {
    const task = makeTask({ title: 'Fix login bug', projectName: 'Auth' });
    expect(taskMatchesTokens(task, ['fix', 'auth'])).toBe(true);
  });

  it('returns false when a token does not match', () => {
    const task = makeTask({ title: 'Fix login bug' });
    expect(taskMatchesTokens(task, ['fix', 'payment'])).toBe(false);
  });
});

describe('searchTasksMultiToken', () => {
  it('filters tasks matching all tokens', () => {
    const tasks = [
      makeTask({ title: 'Fix login bug', projectName: 'Auth' }),
      makeTask({ title: 'Add payment flow', projectName: 'Billing' }),
    ];
    expect(searchTasksMultiToken(tasks, 'fix auth').length).toBe(1);
  });
});

describe('tasksByIdentifierPrefix', () => {
  it('returns tasks with matching prefix', () => {
    const tasks = [makeTask({ identifier: 'FLOW-1' }), makeTask({ identifier: 'CRM-2' })];
    expect(tasksByIdentifierPrefix(tasks, 'FLOW-').map((t) => t.identifier)).toEqual(['FLOW-1']);
  });
});
