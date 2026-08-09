import { describe, it, expect } from 'vitest';

import {
  findFieldById,
  fieldsForProject,
  getFieldValue,
  formatFieldValue,
  isValidDropdownValue,
  nextFieldOrder,
  fieldNameExists,
} from './customFieldUtils';
import type { CustomFieldDefinition, Task } from '../stores/flowStore';

function makeField(overrides: Partial<CustomFieldDefinition>): CustomFieldDefinition {
  return {
    id: 'f1',
    projectId: 'p1',
    name: 'Priority Score',
    type: 'number',
    order: 0,
    ...overrides,
  } as CustomFieldDefinition;
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

const fields = [
  makeField({ id: 'a', projectId: 'p1', order: 2, name: 'Score' }),
  makeField({ id: 'b', projectId: 'p1', order: 0, name: 'Stage' }),
  makeField({ id: 'c', projectId: 'p2', order: 0, name: 'Type' }),
];

describe('findFieldById', () => {
  it('finds field by id', () => {
    expect(findFieldById(fields, 'b')?.name).toBe('Stage');
  });

  it('returns undefined when not found', () => {
    expect(findFieldById(fields, 'z')).toBeUndefined();
  });
});

describe('fieldsForProject', () => {
  it('returns fields for project sorted by order', () => {
    expect(fieldsForProject(fields, 'p1').map((f) => f.id)).toEqual(['b', 'a']);
  });

  it('returns empty for unknown project', () => {
    expect(fieldsForProject(fields, 'unknown')).toHaveLength(0);
  });
});

describe('getFieldValue', () => {
  it('returns the custom field value', () => {
    const task = makeTask({ customFieldValues: { f1: 42 } });
    expect(getFieldValue(task, 'f1')).toBe(42);
  });

  it('returns undefined for missing value', () => {
    const task = makeTask({});
    expect(getFieldValue(task, 'f1')).toBeUndefined();
  });
});

describe('formatFieldValue', () => {
  it('formats number values', () => {
    const field = makeField({ type: 'number' });
    expect(formatFieldValue(field, 42)).toBe('42');
  });

  it('formats text values', () => {
    const field = makeField({ type: 'text' });
    expect(formatFieldValue(field, 'hello')).toBe('hello');
  });

  it('validates dropdown value', () => {
    const field = makeField({ type: 'dropdown', options: ['A', 'B'] });
    expect(formatFieldValue(field, 'A')).toBe('A');
    expect(formatFieldValue(field, 'C')).toBe('');
  });

  it('returns empty string for undefined', () => {
    const field = makeField({ type: 'text' });
    expect(formatFieldValue(field, undefined)).toBe('');
  });
});

describe('isValidDropdownValue', () => {
  it('returns true for valid option', () => {
    const field = makeField({ type: 'dropdown', options: ['Alpha', 'Beta'] });
    expect(isValidDropdownValue(field, 'Alpha')).toBe(true);
  });

  it('returns false for invalid option', () => {
    const field = makeField({ type: 'dropdown', options: ['Alpha'] });
    expect(isValidDropdownValue(field, 'Gamma')).toBe(false);
  });

  it('returns false for non-dropdown field', () => {
    const field = makeField({ type: 'text' });
    expect(isValidDropdownValue(field, 'anything')).toBe(false);
  });
});

describe('nextFieldOrder', () => {
  it('returns max order + 1', () => {
    expect(nextFieldOrder(fields, 'p1')).toBe(3);
  });

  it('returns 0 for empty project', () => {
    expect(nextFieldOrder(fields, 'empty')).toBe(0);
  });
});

describe('fieldNameExists', () => {
  it('returns true for existing name (case-insensitive)', () => {
    const fs = [makeField({ name: 'Score', projectId: 'p1' })];
    expect(fieldNameExists(fs, 'p1', 'score')).toBe(true);
  });

  it('returns false for new name', () => {
    const fs = [makeField({ name: 'Score', projectId: 'p1' })];
    expect(fieldNameExists(fs, 'p1', 'Budget')).toBe(false);
  });
});
