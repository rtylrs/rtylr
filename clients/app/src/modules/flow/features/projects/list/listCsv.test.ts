import { describe, expect, it } from 'vitest';
import type { Task } from '../../../shared/stores/flowStore';
import { buildCsv, csvField, csvFileName, listTasksToCsv } from './listCsv';

const task = (extra: Partial<Task> = {}): Task => ({
  id: '1',
  projectId: '7',
  title: 'Ship it',
  status: 'todo',
  priority: 'high',
  assignee: 'Ada Lovelace',
  assigneeInitials: 'AL',
  assigneeId: '3',
  dueDate: '2026-11-20T00:00:00Z',
  subtaskCount: 0,
  projectName: 'Web',
  createdAt: '2026-11-01T09:00:00Z',
  updatedAt: '',
  identifier: 'WEB-1',
  ...extra,
});

const BOM = '﻿';
const identity = (key: string): string => key;

describe('csvField', () => {
  it('quotes every field and doubles quotes inside it', () => {
    expect(csvField('plain')).toBe('"plain"');
    expect(csvField('say "hi", then go')).toBe('"say ""hi"", then go"');
    expect(csvField('two\nlines')).toBe('"two\nlines"');
    expect(csvField(null)).toBe('""');
    expect(csvField(2.5)).toBe('"2.5"');
  });

  it.each(['=HYPERLINK("x")', '+1+1', '-2', '@SUM(A1)', '\tTAB', '\rCR'])(
    'keeps %j as text rather than a formula',
    (value) => {
      expect(csvField(value).startsWith(`"'`)).toBe(true);
    }
  );

  it('leaves a value that only contains those characters alone', () => {
    expect(csvField('a=b')).toBe('"a=b"');
  });
});

describe('listTasksToCsv', () => {
  it('starts with a byte order mark and ends rows with CRLF', () => {
    expect(buildCsv([['a'], ['b']])).toBe(`${BOM}"a"\r\n"b"\r\n`);
  });

  it('leads with the key and follows the picked columns in order', () => {
    const csv = listTasksToCsv(
      [task({ labelIds: ['1', '2', '9'], estimate: 3 })],
      ['dueDate', 'title', 'labels', 'estimate', 'key'],
      {
        t: identity,
        statuses: [],
        labelNames: new Map([
          ['1', 'Bug'],
          ['2', 'UI'],
        ]),
      }
    );
    const [header, row] = csv.slice(1).split('\r\n');
    expect(header).toBe(
      [
        'flow.projectList.column.key',
        'flowCommon.dueDate',
        'flow.projectList.columnTitle',
        'flow.projectList.column.labels',
        'flow.projectList.column.estimate',
      ]
        .map((h) => `"${h}"`)
        .join(',')
    );
    // Dates as plain days; a label the list no longer knows is left out.
    expect(row).toBe('"WEB-1","2026-11-20","Ship it","Bug; UI","3"');
  });

  it('neutralises a formula in a task title', () => {
    const csv = listTasksToCsv([task({ title: '=cmd|calc' })], ['title'], {
      t: identity,
      statuses: [],
      labelNames: new Map(),
    });
    expect(csv).toContain(`"'=cmd|calc"`);
  });
});

describe('csvFileName', () => {
  it('strips characters a desktop cannot save', () => {
    expect(csvFileName('R&D: "Q4"/plan', '2026-11-19')).toBe('R&D Q4 plan 2026-11-19.csv');
    expect(csvFileName('  ', '2026-11-19')).toBe('tasks 2026-11-19.csv');
  });
});
