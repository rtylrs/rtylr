import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Task } from '../../../shared/stores/flowStore';
import type { ApiStatus } from '../../../shared/services/flowApi';
import {
  DEFAULT_LIST_COLUMNS,
  DEFAULT_LIST_PREFS,
  MAX_COLUMN_WIDTH,
  listPrefsStorageKey,
  loadListPrefs,
  moveListColumn,
  nextSort,
  normalizeListPrefs,
  rowMinWidth,
  saveListPrefs,
  sortListTasks,
  toggleListColumn,
} from './listColumns';

const task = (id: string, extra: Partial<Task> = {}): Task => ({
  id,
  projectId: '7',
  title: `Task ${id}`,
  status: 'todo',
  priority: 'medium',
  assignee: '',
  assigneeInitials: '',
  assigneeId: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Web',
  createdAt: '',
  updatedAt: '',
  identifier: `T-${id}`,
  ...extra,
});

const status = (id: string, sortOrder: number, legacyStatus = 'todo'): ApiStatus => ({
  id,
  projectId: '7',
  name: `S${id}`,
  bucket: 'active',
  legacyStatus,
  color: '#6366f1',
  sortOrder,
  isDefault: false,
  createdAt: '',
});

const ctx = { statuses: [], labelNames: new Map<string, string>() };
const ids = (tasks: Task[]): string[] => tasks.map((t) => t.id);

describe('List column preferences', () => {
  it('defaults to the columns the List always showed', () => {
    expect(DEFAULT_LIST_COLUMNS).toEqual(['priority', 'title', 'assignee', 'dueDate']);
    expect(normalizeListPrefs(null)).toEqual(DEFAULT_LIST_PREFS);
  });

  it('repairs stored values instead of trusting them', () => {
    const prefs = normalizeListPrefs({
      columns: ['status', 'gone', 'status', 'dueDate'],
      widths: { status: 9999, dueDate: 1, gone: 50, labels: 'wide' },
      sort: { column: 'gone', direction: 'asc' },
    });
    // Unknown and duplicate columns dropped, the title put back.
    expect(prefs.columns).toEqual(['title', 'status', 'dueDate']);
    expect(prefs.widths).toEqual({ status: MAX_COLUMN_WIDTH, dueDate: 72 });
    expect(prefs.sort).toBeNull();
  });

  it('never hides the title and clears a sort on a column it hides', () => {
    const sorted = {
      ...DEFAULT_LIST_PREFS,
      sort: { column: 'dueDate' as const, direction: 'asc' as const },
    };
    expect(toggleListColumn(sorted, 'title')).toBe(sorted);
    const hidden = toggleListColumn(sorted, 'dueDate');
    expect(hidden.columns).not.toContain('dueDate');
    expect(hidden.sort).toBeNull();
    const shown = toggleListColumn(hidden, 'labels').columns;
    expect(shown[shown.length - 1]).toBe('labels');
  });

  it('moves a column one place and stops at the edges', () => {
    const moved = moveListColumn(DEFAULT_LIST_PREFS, 'dueDate', -1);
    expect(moved.columns).toEqual(['priority', 'title', 'dueDate', 'assignee']);
    expect(moveListColumn(DEFAULT_LIST_PREFS, 'priority', -1)).toBe(DEFAULT_LIST_PREFS);
  });

  it('cycles a header through ascending, descending and off', () => {
    const asc = nextSort(null, 'title');
    expect(asc).toEqual({ column: 'title', direction: 'asc' });
    expect(nextSort(asc, 'title')).toEqual({ column: 'title', direction: 'desc' });
    expect(nextSort({ column: 'title', direction: 'desc' }, 'title')).toBeNull();
    expect(nextSort(asc, 'dueDate')).toEqual({ column: 'dueDate', direction: 'asc' });
  });

  it('widens the row by every column it shows', () => {
    const wider = { ...DEFAULT_LIST_PREFS, widths: { dueDate: 196 } };
    expect(rowMinWidth(wider) - rowMinWidth(DEFAULT_LIST_PREFS)).toBe(100);
  });
});

describe('List preference storage', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
  });

  it('keeps one entry per user and project', () => {
    const prefs = toggleListColumn(DEFAULT_LIST_PREFS, 'status');
    saveListPrefs('u1', '7', prefs);
    expect(localStorage.getItem(listPrefsStorageKey('u1', '7'))).not.toBeNull();
    expect(loadListPrefs('u1', '7').columns).toContain('status');
    expect(loadListPrefs('u2', '7').columns).toEqual(DEFAULT_LIST_COLUMNS);
    expect(loadListPrefs('u1', '8').columns).toEqual(DEFAULT_LIST_COLUMNS);
  });

  it('writes nothing without a user and survives blocked storage', () => {
    saveListPrefs(null, '7', DEFAULT_LIST_PREFS);
    expect(localStorage.length).toBe(0);
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    expect(loadListPrefs('u1', '7')).toEqual(DEFAULT_LIST_PREFS);
  });
});

describe('sortListTasks', () => {
  it('keeps the server order with no sort', () => {
    const tasks = [task('b'), task('a')];
    expect(ids(sortListTasks(tasks, null, ctx))).toEqual(['b', 'a']);
  });

  it('puts rows without a value last in both directions', () => {
    const tasks = [
      task('none'),
      task('late', { dueDate: '2026-12-01' }),
      task('soon', { dueDate: '2026-11-20' }),
    ];
    expect(ids(sortListTasks(tasks, { column: 'dueDate', direction: 'asc' }, ctx))).toEqual([
      'soon',
      'late',
      'none',
    ]);
    expect(ids(sortListTasks(tasks, { column: 'dueDate', direction: 'desc' }, ctx))).toEqual([
      'late',
      'soon',
      'none',
    ]);
  });

  it('sorts priority by urgency and status by the project order', () => {
    const tasks = [
      task('low', { priority: 'low', statusId: '2' }),
      task('urgent', { priority: 'urgent', statusId: '1' }),
    ];
    expect(ids(sortListTasks(tasks, { column: 'priority', direction: 'asc' }, ctx))).toEqual([
      'urgent',
      'low',
    ]);
    const statuses = [status('2', 20), status('1', 10)];
    expect(
      ids(sortListTasks(tasks, { column: 'status', direction: 'asc' }, { ...ctx, statuses }))
    ).toEqual(['urgent', 'low']);
  });

  it('compares keys as numbers and labels by name', () => {
    const tasks = [
      task('10', { identifier: 'T-10', labelIds: ['2'] }),
      task('9', { identifier: 'T-9', labelIds: ['1'] }),
    ];
    expect(ids(sortListTasks(tasks, { column: 'key', direction: 'asc' }, ctx))).toEqual([
      '9',
      '10',
    ]);
    const labelNames = new Map([
      ['1', 'Zebra'],
      ['2', 'Alpha'],
    ]);
    expect(
      ids(sortListTasks(tasks, { column: 'labels', direction: 'asc' }, { ...ctx, labelNames }))
    ).toEqual(['10', '9']);
  });
});
