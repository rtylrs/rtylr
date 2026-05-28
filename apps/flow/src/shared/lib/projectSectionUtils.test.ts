import { describe, it, expect } from 'vitest';

import {
  sectionsForProject,
  findSectionById,
  tasksInSection,
  unsectionedTasks,
  nextSectionOrder,
  sectionNameExists,
} from './projectSectionUtils';
import type { ProjectSection, Task } from '../stores/flowStore';

function makeSection(overrides: Partial<ProjectSection>): ProjectSection {
  return {
    id: 's1',
    projectId: 'p1',
    name: 'To Do',
    order: 0,
    ...overrides,
  } as ProjectSection;
}

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

const sections = [
  makeSection({ id: 'a', projectId: 'p1', order: 2 }),
  makeSection({ id: 'b', projectId: 'p1', order: 0 }),
  makeSection({ id: 'c', projectId: 'p2', order: 0 }),
];

describe('sectionsForProject', () => {
  it('returns sections for the given project sorted by order', () => {
    expect(sectionsForProject(sections, 'p1').map((s) => s.id)).toEqual(['b', 'a']);
  });

  it('returns empty for unknown project', () => {
    expect(sectionsForProject(sections, 'unknown')).toHaveLength(0);
  });
});

describe('findSectionById', () => {
  it('finds section by id', () => {
    expect(findSectionById(sections, 'c')?.projectId).toBe('p2');
  });

  it('returns undefined when not found', () => {
    expect(findSectionById(sections, 'z')).toBeUndefined();
  });
});

describe('tasksInSection', () => {
  it('returns tasks belonging to the section', () => {
    const tasks = [
      makeTask({ id: 'a', sectionId: 's1' }),
      makeTask({ id: 'b', sectionId: 's2' }),
      makeTask({ id: 'c', sectionId: 's1' }),
    ];
    expect(tasksInSection(tasks, 's1').map((t) => t.id)).toEqual(['a', 'c']);
  });
});

describe('unsectionedTasks', () => {
  it('returns tasks without a section for the project', () => {
    const tasks = [
      makeTask({ id: 'a', projectId: 'p1', sectionId: 's1' }),
      makeTask({ id: 'b', projectId: 'p1' }),
      makeTask({ id: 'c', projectId: 'p2' }),
    ];
    expect(unsectionedTasks(tasks, 'p1').map((t) => t.id)).toEqual(['b']);
  });
});

describe('nextSectionOrder', () => {
  it('returns max order + 1', () => {
    expect(nextSectionOrder(sections, 'p1')).toBe(3);
  });

  it('returns 0 for empty project', () => {
    expect(nextSectionOrder(sections, 'empty')).toBe(0);
  });
});

describe('sectionNameExists', () => {
  it('returns true for existing name (case insensitive)', () => {
    const secs = [makeSection({ name: 'To Do', projectId: 'p1' })];
    expect(sectionNameExists(secs, 'p1', 'to do')).toBe(true);
  });

  it('returns false for new name', () => {
    const secs = [makeSection({ name: 'To Do', projectId: 'p1' })];
    expect(sectionNameExists(secs, 'p1', 'In Progress')).toBe(false);
  });
});
