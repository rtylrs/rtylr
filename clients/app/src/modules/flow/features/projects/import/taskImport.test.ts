import { describe, expect, it } from 'vitest';
import { buildImportPreview, parseEffort } from './taskImport';

const ctx = {
  people: [{ id: '7', name: 'Mhmd Issa', email: 'mhmd@example.test' }],
  statuses: [
    { id: '30', name: 'To Do' },
    { id: '31', name: 'In Review' },
  ],
};

describe('parseEffort', () => {
  it('reads minutes, hours and mixed values', () => {
    expect(parseEffort('90')).toBe(90);
    expect(parseEffort('2', true)).toBe(120);
    expect(parseEffort('1h')).toBe(60);
    expect(parseEffort('1.5h')).toBe(90);
    expect(parseEffort('1h 30m')).toBe(90);
    expect(parseEffort('45 min')).toBe(45);
    expect(parseEffort('soon')).toBeNull();
  });
});

describe('buildImportPreview', () => {
  it('maps headers, people, sections and subtasks', () => {
    const csv = [
      'Task name,Details,Assignee email,Due date,Priority,Effort,Status,Subtasks',
      'Login page,Build it,MHMD@example.test,2026-12-01,High,1h,in review,Form; Errors | Copy',
    ].join('\n');
    const { rows, missingTitle } = buildImportPreview(csv, ctx);
    expect(missingTitle).toBe(false);
    expect(rows[0]).toMatchObject({
      line: 1,
      title: 'Login page',
      description: 'Build it',
      assigneeId: '7',
      dueDate: '2026-12-01',
      priority: 'high',
      effortMinutes: 60,
      sectionId: '31',
      subtasks: ['Form', 'Errors', 'Copy'],
      issues: [],
    });
  });

  it('flags what it cannot match instead of guessing', () => {
    const csv = 'Title,Email,Due,Priority,Section\n,nobody@x.test,01/12/2026,asap,Shipped';
    const [row] = buildImportPreview(csv, ctx).rows;
    expect(row.issues.map((i) => i.kind)).toEqual([
      'noTitle',
      'unknownAssignee',
      'badDate',
      'badPriority',
      'unknownSection',
    ]);
  });

  it('says when the file has no title column', () => {
    expect(buildImportPreview('Name of thing\nx', ctx).missingTitle).toBe(true);
  });
});
