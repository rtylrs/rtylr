import { describe, it, expect } from 'vitest';

import {
  taskPath,
  projectPath,
  projectBoardPath,
  cyclePath,
  milestonePath,
  inboxPath,
  myTasksPath,
  workspaceSettingsPath,
  withQueryParams,
  parseQueryString,
} from './urlUtils';

describe('taskPath', () => {
  it('generates task path', () => {
    expect(taskPath('w1', 't1')).toBe('/w/w1/tasks/t1');
  });
});

describe('projectPath', () => {
  it('generates project path', () => {
    expect(projectPath('w1', 'p1')).toBe('/w/w1/projects/p1');
  });
});

describe('projectBoardPath', () => {
  it('generates board path', () => {
    expect(projectBoardPath('w1', 'p1')).toBe('/w/w1/projects/p1/board');
  });
});

describe('cyclePath', () => {
  it('generates cycle path', () => {
    expect(cyclePath('w1', 'c1')).toBe('/w/w1/cycles/c1');
  });
});

describe('milestonePath', () => {
  it('generates milestone path', () => {
    expect(milestonePath('w1', 'm1')).toBe('/w/w1/milestones/m1');
  });
});

describe('inboxPath', () => {
  it('generates inbox path', () => {
    expect(inboxPath('w1')).toBe('/w/w1/inbox');
  });
});

describe('myTasksPath', () => {
  it('generates my tasks path', () => {
    expect(myTasksPath('w1')).toBe('/w/w1/my-tasks');
  });
});

describe('workspaceSettingsPath', () => {
  it('generates settings path', () => {
    expect(workspaceSettingsPath('w1')).toBe('/w/w1/settings');
  });
});

describe('withQueryParams', () => {
  it('appends query string', () => {
    const result = withQueryParams('/tasks', { status: 'done', page: 1 });
    expect(result).toContain('status=done');
    expect(result).toContain('page=1');
  });

  it('skips undefined params', () => {
    const result = withQueryParams('/tasks', { status: undefined });
    expect(result).toBe('/tasks');
  });

  it('returns path unchanged for empty params', () => {
    expect(withQueryParams('/tasks', {})).toBe('/tasks');
  });
});

describe('parseQueryString', () => {
  it('parses query string into object', () => {
    const result = parseQueryString('?status=done&page=1');
    expect(result.status).toBe('done');
    expect(result.page).toBe('1');
  });

  it('handles string without leading ?', () => {
    const result = parseQueryString('status=todo');
    expect(result.status).toBe('todo');
  });
});
