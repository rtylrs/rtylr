import { describe, it, expect } from 'vitest';

import {
  findTemplateById,
  sortTemplatesAlpha,
  templatesByPriority,
  taskFromTemplate,
  templateNameExists,
  countTemplatesByPriority,
} from './templateUtils';
import type { TaskTemplate } from '../stores/flowStore';

function makeTemplate(overrides: Partial<TaskTemplate>): TaskTemplate {
  return {
    id: 'tmpl1',
    name: 'Bug Fix',
    title: 'Fix the bug',
    priority: 'high',
    status: 'todo',
    labels: [],
    subtasks: [],
    ...overrides,
  } as TaskTemplate;
}

describe('findTemplateById', () => {
  it('finds template by id', () => {
    const templates = [makeTemplate({ id: 'a' }), makeTemplate({ id: 'b', name: 'Feature' })];
    expect(findTemplateById(templates, 'b')?.name).toBe('Feature');
  });

  it('returns undefined when not found', () => {
    expect(findTemplateById([], 'x')).toBeUndefined();
  });
});

describe('sortTemplatesAlpha', () => {
  it('sorts alphabetically by name', () => {
    const templates = [
      makeTemplate({ name: 'Zebra' }),
      makeTemplate({ name: 'Alpha' }),
      makeTemplate({ name: 'Mango' }),
    ];
    expect(sortTemplatesAlpha(templates).map((t) => t.name)).toEqual(['Alpha', 'Mango', 'Zebra']);
  });

  it('does not mutate original', () => {
    const templates = [makeTemplate({ name: 'Z' }), makeTemplate({ name: 'A' })];
    const original = [...templates];
    sortTemplatesAlpha(templates);
    expect(templates).toEqual(original);
  });
});

describe('templatesByPriority', () => {
  it('filters by priority', () => {
    const templates = [
      makeTemplate({ id: 'a', priority: 'high' }),
      makeTemplate({ id: 'b', priority: 'low' }),
      makeTemplate({ id: 'c', priority: 'high' }),
    ];
    expect(templatesByPriority(templates, 'high').map((t) => t.id)).toEqual(['a', 'c']);
  });
});

describe('taskFromTemplate', () => {
  it('creates partial task from template', () => {
    const template = makeTemplate({
      title: 'Write tests',
      priority: 'medium',
      status: 'todo',
      labels: ['label1'],
      subtasks: ['subtask1', 'subtask2'],
    });
    const task = taskFromTemplate(template, 'p1', 'w1');
    expect(task.title).toBe('Write tests');
    expect(task.priority).toBe('medium');
    expect(task.projectId).toBe('p1');
    expect(task.workspaceId).toBe('w1');
    expect(task.labelIds).toEqual(['label1']);
    expect(task.subtaskCount).toBe(2);
  });
});

describe('templateNameExists', () => {
  it('returns true for existing name (case-insensitive)', () => {
    const templates = [makeTemplate({ name: 'Bug Fix' })];
    expect(templateNameExists(templates, 'bug fix')).toBe(true);
  });

  it('returns false for new name', () => {
    const templates = [makeTemplate({ name: 'Bug Fix' })];
    expect(templateNameExists(templates, 'Feature Request')).toBe(false);
  });
});

describe('countTemplatesByPriority', () => {
  it('counts templates per priority', () => {
    const templates = [
      makeTemplate({ priority: 'high' }),
      makeTemplate({ priority: 'medium' }),
      makeTemplate({ priority: 'high' }),
    ];
    const result = countTemplatesByPriority(templates);
    expect(result.high).toBe(2);
    expect(result.medium).toBe(1);
  });
});
