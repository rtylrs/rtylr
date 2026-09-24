import { describe, expect, it } from 'vitest';

import { apiTaskToTask } from './taskMapper';
import type { ApiTask } from '../services/flowApi';

/**
 * Regression tests for the wire → store mapper. Most of these exist
 * specifically because the bugs they guard against have bitten before:
 *   - numeric wire IDs leaking into a string-keyed store caused tasks
 *     to disappear after drag-to-move (sectionId mismatch)
 *   - missing display fields (projectName, assigneeName) rendered
 *     `undefined` in the UI
 * We're not testing every field - just the coercion + fallback logic
 * that's actually dangerous to get wrong.
 */

function baseApiTask(overrides: Partial<ApiTask> = {}): ApiTask {
  return {
    id: '1',
    projectId: '10',
    workspaceId: '100',
    title: 'Task',
    description: '',
    status: 'todo',
    priority: 'medium',
    dueDate: '',
    labels: [],
    ...overrides,
  } as ApiTask;
}

describe('apiTaskToTask', () => {
  it('stringifies numeric wire IDs', () => {
    // The wire sends `id: 42` as a JSON number even though the
    // hand-rolled ApiTask lies about it being a string. The mapper
    // must always stringify - regression guard for the "moved tasks
    // vanish from the board" bug.
    const input = baseApiTask({
      id: 42 as unknown as string,
      projectId: 5 as unknown as string,
      workspaceId: 9 as unknown as string,
    });
    const out = apiTaskToTask(input);
    expect(out.id).toBe('42');
    expect(out.projectId).toBe('5');
    expect(out.workspaceId).toBe('9');
  });

  it('preserves string IDs as-is', () => {
    const out = apiTaskToTask(baseApiTask({ id: '7' }));
    expect(out.id).toBe('7');
  });

  it('stringifies numeric sectionId', () => {
    // sectionId is the column the task sits in on the kanban board.
    // The bug: numeric sectionId === string column id was always
    // false, so tasks landed in no column and disappeared.
    const out = apiTaskToTask(baseApiTask({ sectionId: 3 as unknown as string | undefined }));
    expect(out.sectionId).toBe('3');
  });

  it('treats sectionId=0 as undefined (unsectioned)', () => {
    // The backend emits 0 when the task has no section. In the store,
    // an undefined sectionId means "goes in first column"; a string
    // "0" would trigger a column lookup miss.
    const out = apiTaskToTask(baseApiTask({ sectionId: 0 as unknown as string | undefined }));
    expect(out.sectionId).toBeUndefined();
  });

  it('treats missing sectionId as undefined', () => {
    const out = apiTaskToTask(baseApiTask());
    expect(out.sectionId).toBeUndefined();
  });

  it('prefers assigneeName over legacy assignee field', () => {
    // Newer proto sends `assigneeName`; older code wrote `assignee`.
    // When both exist the new one wins.
    const out = apiTaskToTask(
      baseApiTask({
        assignee: 'legacy',
        assigneeName: 'canonical',
      } as Partial<ApiTask>)
    );
    expect(out.assignee).toBe('canonical');
  });

  it('falls back to legacy assignee when assigneeName is missing', () => {
    const out = apiTaskToTask(baseApiTask({ assignee: 'Jane' }));
    expect(out.assignee).toBe('Jane');
  });

  it('returns empty string for assignee when neither field is present', () => {
    // UI binds to assignee directly; `undefined` would render literal
    // "undefined" in some display paths.
    const out = apiTaskToTask(baseApiTask());
    expect(out.assignee).toBe('');
  });

  it('defaults createdAt and updatedAt to now when missing', () => {
    const before = new Date().toISOString();
    const out = apiTaskToTask(baseApiTask());
    expect(out.createdAt >= before).toBe(true);
    expect(out.updatedAt >= before).toBe(true);
  });

  it('uses id as identifier fallback when identifier is missing', () => {
    const out = apiTaskToTask(baseApiTask({ id: '99' }));
    expect(out.identifier).toBe('99');
  });

  it('passes labels through unchanged', () => {
    const out = apiTaskToTask(baseApiTask({ labels: [1, 2, 3] as unknown as string[] }));
    expect(out.labelIds).toEqual([1, 2, 3]);
  });

  it('preserves the archived flag', () => {
    // Dropping this silently defeats every `!t.archived` filter downstream
    // (Triage, FocusToday, FlowLayout sidebar counts all read the mapped
    // Task, not the raw ApiTask) - they'd never actually exclude anything.
    const out = apiTaskToTask(baseApiTask({ archived: true }));
    expect(out.archived).toBe(true);
  });

  it('normalises a missing archived field to false', () => {
    // `archived` is omitempty on the contract, so an unarchived task comes
    // back with the key absent. The mapper resolves that to a real `false`
    // rather than leaking `undefined` into the store.
    const out = apiTaskToTask(baseApiTask());
    expect(out.archived).toBe(false);
  });

  it('derives quick-create effort minutes when only estimated hours are returned', () => {
    const out = apiTaskToTask(baseApiTask({ estimatedHours: 1.5 }));
    expect(out.effortMinutes).toBe(90);
  });

  it('prefers an explicit effort value over the estimated-hours fallback', () => {
    const out = apiTaskToTask(baseApiTask({ effortMinutes: 30, estimatedHours: 1.5 }));
    expect(out.effortMinutes).toBe(30);
  });
});

describe('assignee initials cannot disagree with the assignee name (#1097 D-040)', () => {
  it('derives them from the name the row actually shows', () => {
    // The wire carried two sources for one fact: `assignee` prefers the newer
    // assigneeName, while assigneeInitials only ever carried the legacy one.
    // A task detail showed "FO" beside "Mohammad Al-Homsi".
    const task = apiTaskToTask({
      id: '1',
      projectId: '1',
      workspaceId: '1',
      title: 'Anything',
      status: 'todo',
      priority: 'medium',
      assignee: 'Fatima Osman',
      assigneeName: 'Mohammad Al-Homsi',
      assigneeInitials: 'FO',
    } as unknown as Parameters<typeof apiTaskToTask>[0]);

    expect(task.assignee).toBe('Mohammad Al-Homsi');
    expect(task.assigneeInitials).toBe('MA');
  });

  it('still uses the legacy field when there is no name to derive from', () => {
    const task = apiTaskToTask({
      id: '2',
      projectId: '1',
      workspaceId: '1',
      title: 'Anything',
      status: 'todo',
      priority: 'medium',
      assignee: '',
      assigneeInitials: 'ZZ',
    } as unknown as Parameters<typeof apiTaskToTask>[0]);

    expect(task.assigneeInitials).toBe('ZZ');
  });

  it('agrees with the legacy field when the two sources already matched', () => {
    const task = apiTaskToTask({
      id: '3',
      projectId: '1',
      workspaceId: '1',
      title: 'Anything',
      status: 'todo',
      priority: 'medium',
      assignee: 'Fatima Osman',
      assigneeInitials: 'FO',
    } as unknown as Parameters<typeof apiTaskToTask>[0]);

    expect(task.assigneeInitials).toBe('FO');
  });
});
