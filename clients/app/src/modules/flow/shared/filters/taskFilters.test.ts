import { describe, it, expect } from 'vitest';
import type { Task } from '../stores/flowStore';
import {
  EMPTY_TASK_FILTERS,
  UNASSIGNED,
  activeFieldCount,
  hasActiveTaskFilters,
  isOnlyMine,
  matchesTaskFilters,
  parseTaskFilters,
  serializeTaskFilters,
  toggleOnlyMine,
  type TaskFilters,
} from './taskFilters';

const TODAY = '2026-11-18';

function task(overrides: Partial<Task> = {}): Task {
  return {
    id: '1',
    projectId: 'p1',
    title: 'Ship the payment-plan tracker',
    status: 'todo',
    priority: 'medium',
    assignee: 'Ahmed',
    assigneeInitials: 'AH',
    assigneeId: 'u1',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Investment Hub',
    createdAt: '2026-11-01',
    updatedAt: '2026-11-01',
    identifier: 'T-213',
    ...overrides,
  };
}

const f = (overrides: Partial<TaskFilters>): TaskFilters => ({
  ...EMPTY_TASK_FILTERS,
  ...overrides,
});
const ctx = { today: TODAY, labelNames: new Map([['7', 'Design']]) };

describe('matchesTaskFilters', () => {
  it('passes everything with no filters', () => {
    expect(matchesTaskFilters(task(), EMPTY_TASK_FILTERS, ctx)).toBe(true);
  });

  it('ORs values inside a field', () => {
    const filters = f({ assignees: ['u2', 'u1'] });
    expect(matchesTaskFilters(task(), filters, ctx)).toBe(true);
    expect(matchesTaskFilters(task({ assigneeId: 'u3' }), filters, ctx)).toBe(false);
  });

  it('ANDs across fields', () => {
    const filters = f({ assignees: ['u1'], priorities: ['high'] });
    expect(matchesTaskFilters(task({ priority: 'high' }), filters, ctx)).toBe(true);
    expect(matchesTaskFilters(task({ priority: 'low' }), filters, ctx)).toBe(false);
  });

  it('matches Unassigned against an empty assignee', () => {
    const filters = f({ assignees: [UNASSIGNED] });
    expect(matchesTaskFilters(task({ assigneeId: '' }), filters, ctx)).toBe(true);
    expect(matchesTaskFilters(task(), filters, ctx)).toBe(false);
  });

  it('matches any of the chosen labels', () => {
    const filters = f({ labels: ['7', '9'] });
    expect(matchesTaskFilters(task({ labelIds: ['9'] }), filters, ctx)).toBe(true);
    expect(matchesTaskFilters(task({ labelIds: [] }), filters, ctx)).toBe(false);
  });

  it('reads the status through the context when given', () => {
    const filters = f({ statuses: ['s-42'] });
    const byId = { ...ctx, statusOf: (t: Task) => t.statusId ?? '' };
    expect(matchesTaskFilters(task({ statusId: 's-42' } as Partial<Task>), filters, byId)).toBe(
      true
    );
    expect(matchesTaskFilters(task(), filters, byId)).toBe(false);
    expect(matchesTaskFilters(task(), f({ statuses: ['todo'] }), ctx)).toBe(true);
  });

  it.each([
    ['overdue', '2026-11-17', true],
    ['overdue', TODAY, false],
    ['today', TODAY, true],
    ['next_7', '2026-11-25', true],
    ['next_7', '2026-11-26', false],
    ['next_30', '2026-12-18', true],
    ['no_date', '', true],
    ['no_date', TODAY, false],
  ] as const)('due window %s against %s is %s', (window, dueDate, expected) => {
    expect(matchesTaskFilters(task({ dueDate }), f({ due: [window] }), ctx)).toBe(expected);
  });

  it('compares due dates as days, ignoring any time part', () => {
    const filters = f({ due: ['today'] });
    expect(matchesTaskFilters(task({ dueDate: `${TODAY}T23:30:00Z` }), filters, ctx)).toBe(true);
  });

  it('treats a custom range as inclusive and ORs it with the windows', () => {
    const filters = f({ due: ['overdue'], dueFrom: '2026-12-01', dueTo: '2026-12-31' });
    expect(matchesTaskFilters(task({ dueDate: '2026-12-01' }), filters, ctx)).toBe(true);
    expect(matchesTaskFilters(task({ dueDate: '2026-12-31' }), filters, ctx)).toBe(true);
    expect(matchesTaskFilters(task({ dueDate: '2026-11-01' }), filters, ctx)).toBe(true);
    expect(matchesTaskFilters(task({ dueDate: '2026-11-20' }), filters, ctx)).toBe(false);
    expect(matchesTaskFilters(task({ dueDate: '' }), filters, ctx)).toBe(false);
  });

  describe('recent changes', () => {
    // Midday UTC lands on the same calendar day in every timezone the tests run in.
    const at = (day: string): string => `${day}T12:00:00Z`;

    it.each([
      ['created_7d', { createdAt: at('2026-11-12') }, true],
      ['created_7d', { createdAt: at('2026-11-11') }, false],
      ['updated_7d', { updatedAt: at(TODAY) }, true],
      ['completed_7d', { completedAt: at('2026-11-15') }, true],
      ['completed_7d', { completedAt: undefined }, false],
    ] as const)('%s against %o is %s', (window, overrides, expected) => {
      expect(matchesTaskFilters(task(overrides), f({ changed: [window] }), ctx)).toBe(expected);
    });

    it('ORs windows and ANDs them with the other fields', () => {
      const filters = f({ changed: ['created_7d', 'completed_7d'], priorities: ['high'] });
      const recent = task({ createdAt: at('2026-11-01'), completedAt: at(TODAY) });
      expect(matchesTaskFilters({ ...recent, priority: 'high' }, filters, ctx)).toBe(true);
      expect(matchesTaskFilters(recent, filters, ctx)).toBe(false);
    });
  });

  describe('search', () => {
    it('matches the identifier', () => {
      expect(matchesTaskFilters(task(), f({ query: 't-213' }), ctx)).toBe(true);
    });

    it('needs every word, in any field', () => {
      expect(matchesTaskFilters(task(), f({ query: 'tracker payment' }), ctx)).toBe(true);
      expect(matchesTaskFilters(task(), f({ query: 'tracker invoice' }), ctx)).toBe(false);
    });

    it('matches the description and label names', () => {
      const t = task({ description: 'Milestones and due dates', labelIds: ['7'] });
      expect(matchesTaskFilters(t, f({ query: 'milestones' }), ctx)).toBe(true);
      expect(matchesTaskFilters(t, f({ query: 'design' }), ctx)).toBe(true);
    });

    it('ignores accents and Arabic diacritics', () => {
      expect(matchesTaskFilters(task({ title: 'Café menu' }), f({ query: 'cafe' }), ctx)).toBe(
        true
      );
      expect(matchesTaskFilters(task({ title: 'مَشْرُوع' }), f({ query: 'مشروع' }), ctx)).toBe(
        true
      );
    });
  });
});

describe('counting', () => {
  it('counts fields, not values, and leaves search out of the badge', () => {
    const filters = f({
      priorities: ['high', 'urgent'],
      due: ['today'],
      dueFrom: TODAY,
      query: 'x',
    });
    expect(activeFieldCount(filters)).toBe(2);
    expect(activeFieldCount(f({ changed: ['created_7d', 'updated_7d'] }))).toBe(1);
    expect(hasActiveTaskFilters(f({ changed: ['updated_7d'] }))).toBe(true);
    expect(hasActiveTaskFilters(f({ query: '  ' }))).toBe(false);
    expect(hasActiveTaskFilters(f({ query: 'x' }))).toBe(true);
  });
});

describe('only mine', () => {
  it('is exactly the current user in the assignee field', () => {
    expect(isOnlyMine(f({ assignees: ['u1'] }), 'u1')).toBe(true);
    expect(isOnlyMine(f({ assignees: ['u1', 'u2'] }), 'u1')).toBe(false);
  });

  it('replaces other people when turned on and clears when turned off', () => {
    const on = toggleOnlyMine(f({ assignees: ['u2'] }), 'u1');
    expect(on.assignees).toEqual(['u1']);
    expect(toggleOnlyMine(on, 'u1').assignees).toEqual([]);
  });
});

describe('serialising', () => {
  it('leaves empty fields out and round-trips', () => {
    const filters = f({ assignees: ['u1'], due: ['overdue'], dueTo: '2026-12-01', query: 'tax' });
    const out = serializeTaskFilters(filters);
    expect(out).toEqual({ assignees: ['u1'], due: ['overdue'], dueTo: '2026-12-01', query: 'tax' });
    expect(parseTaskFilters(out)).toEqual(filters);
    expect(serializeTaskFilters(EMPTY_TASK_FILTERS)).toBeUndefined();
  });

  it('carries recent change windows and drops unknown ones', () => {
    expect(serializeTaskFilters(f({ changed: ['completed_7d'] }))).toEqual({
      changed: ['completed_7d'],
    });
    expect(parseTaskFilters({ changed: ['created_7d', 'created_30d'] }).changed).toEqual([
      'created_7d',
    ]);
  });

  it('drops values it does not know instead of failing the whole filter', () => {
    const parsed = parseTaskFilters({
      priorities: ['high', 'whenever'],
      due: ['yesterday', 'today'],
      dueFrom: 'not-a-date',
      labels: [7, 'x'],
      assignees: 'u1',
    });
    expect(parsed.priorities).toEqual(['high']);
    expect(parsed.due).toEqual(['today']);
    expect(parsed.dueFrom).toBeUndefined();
    expect(parsed.labels).toEqual(['7', 'x']);
    expect(parsed.assignees).toEqual([]);
    expect(parseTaskFilters(null)).toEqual(EMPTY_TASK_FILTERS);
  });
});

describe('epic filter', () => {
  const epic = task({ id: '10', taskType: 'epic', title: 'Payments' });
  const child = task({ id: '11', epicId: '10' });
  const subtaskOfChild = task({ id: '12', parentId: '11', epicId: '10' });
  const loose = task({ id: '13' });

  it('keeps the epic, its tasks and their subtasks', () => {
    const filters = f({ epics: ['10'] });
    expect(
      [epic, child, subtaskOfChild, loose].filter((t) => matchesTaskFilters(t, filters, ctx))
    ).toEqual([epic, child, subtaskOfChild]);
  });

  it('matches tasks under no epic with No epic', () => {
    const filters = f({ epics: ['none'] });
    expect(matchesTaskFilters(loose, filters, ctx)).toBe(true);
    expect(matchesTaskFilters(child, filters, ctx)).toBe(false);
    expect(matchesTaskFilters(epic, filters, ctx)).toBe(false);
  });

  it('counts as a field and survives the URL round trip', () => {
    const filters = f({ epics: ['10'] });
    expect(activeFieldCount(filters)).toBe(1);
    expect(serializeTaskFilters(filters)).toEqual({ epics: ['10'] });
    expect(parseTaskFilters(serializeTaskFilters(filters))).toEqual(filters);
  });
});
