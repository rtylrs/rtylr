// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

const { storeState, tasksResult, detailsState, roleQuery, restrictingRole } = vi.hoisted(() => ({
  storeState: {
    selectedTaskIds: [] as string[],
    clearSelection: vi.fn(),
    // The store's own task list. `setTasks` has no call site in the app, so in
    // production this is permanently empty; the tests keep it that way to make
    // sure nothing has quietly started depending on it again.
    tasks: [] as unknown[],
  },
  tasksResult: { data: [] as { id: string; projectId: string }[] },
  // What useTaskDetails returns for the ids absent from tasksResult.data.
  detailsState: {
    ids: [] as string[],
    value: [] as { data?: { task: { projectId: string } } }[],
  },
  roleQuery: { ids: [] as string[] },
  restrictingRole: { value: undefined as string | undefined },
}));

vi.mock('../stores/flowStore', () => ({
  useFlowStore: (selector?: (state: object) => unknown) => {
    const state = storeState;
    return selector ? selector(state) : state;
  },
}));

vi.mock('../../features/tasks/useTasks', () => ({
  useTasks: () => tasksResult,
  useTaskDetails: (ids: string[]) => {
    detailsState.ids = ids;
    return detailsState.value;
  },
  useBulkUpdateStatus: () => ({ isPending: false, mutateAsync: vi.fn() }),
  useBulkUpdatePriority: () => ({ isPending: false, mutateAsync: vi.fn() }),
  useBulkUpdateArchive: () => ({ isPending: false, mutateAsync: vi.fn() }),
  useBulkDelete: () => ({ isPending: false, mutateAsync: vi.fn() }),
}));

vi.mock('../lib/currentUser', () => ({
  useMostRestrictiveProjectRole: (projectIds: string[]) => {
    roleQuery.ids = projectIds;
    return restrictingRole.value;
  },
}));

vi.mock('../services/i18next', () => ({
  default: { t: (key: string) => key },
}));

vi.mock('motion/react', () => ({
  AnimatePresence: ({ children }: { children: React.ReactNode }) => children,
  motion: {
    div: ({ children, ...rest }: { children: React.ReactNode }) =>
      React.createElement('div', rest, children),
  },
}));

import { BulkEditBar } from './BulkEditBar';

const markDone = (): HTMLButtonElement =>
  screen.getByRole('button', { name: /flow.bulkEdit.markDone/ }) as HTMLButtonElement;
const deleteButton = (): HTMLButtonElement =>
  screen.getByRole('button', { name: /flowCommon.delete/ }) as HTMLButtonElement;

describe('BulkEditBar resolves its selection from the task query (#1042)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    storeState.selectedTaskIds = ['1', '2'];
    storeState.tasks = [];
    tasksResult.data = [];
    detailsState.ids = [];
    detailsState.value = [];
    roleQuery.ids = [];
    restrictingRole.value = undefined;
  });

  it('enables the editing actions for an editor when the query resolves the selection', () => {
    tasksResult.data = [
      { id: '1', projectId: '900' },
      { id: '2', projectId: '900' },
    ];
    restrictingRole.value = 'member';

    render(<BulkEditBar />);

    expect(markDone()).not.toBeDisabled();
    // A member may edit but not delete.
    expect(deleteButton()).toBeDisabled();
  });

  it('enables delete for a project admin', () => {
    tasksResult.data = [{ id: '1', projectId: '900' }];
    storeState.selectedTaskIds = ['1'];
    restrictingRole.value = 'admin';

    render(<BulkEditBar />);

    expect(deleteButton()).not.toBeDisabled();
  });

  it('matches ids across the numeric wire values the API actually sends', () => {
    // ApiTask types id and projectId as strings, but int64 arrives as a number.
    tasksResult.data = [{ id: 1, projectId: 900 } as unknown as { id: string; projectId: string }];
    storeState.selectedTaskIds = ['1'];
    restrictingRole.value = 'member';

    render(<BulkEditBar />);

    expect(markDone()).not.toBeDisabled();
  });

  it('resolves a selected task the first page of the list does not contain (#01329)', () => {
    // The business-wide list is one page (the server caps it at 50 rows), so a
    // task selected on its project board past that page is not in
    // tasksResult.data. Its detail is fetched by id instead of leaving every
    // bulk action disabled with the "unresolved" tooltip.
    tasksResult.data = [{ id: '1', projectId: '900' }];
    detailsState.value = [{ data: { task: { projectId: '901' } } }];
    storeState.selectedTaskIds = ['1', '2'];
    restrictingRole.value = 'admin';

    render(<BulkEditBar />);

    expect(markDone()).not.toBeDisabled();
    expect(deleteButton()).not.toBeDisabled();
    expect(detailsState.ids).toEqual(['2']);
    expect(roleQuery.ids).toEqual(['900', '901']);
  });

  it('still fails closed when a selected task resolves neither way', () => {
    tasksResult.data = [{ id: '1', projectId: '900' }];
    detailsState.value = [{ data: undefined }];
    storeState.selectedTaskIds = ['1', '2'];
    restrictingRole.value = 'admin';

    render(<BulkEditBar />);

    expect(markDone()).toBeDisabled();
    expect(markDone().getAttribute('title')).toBe('flow.bulkEdit.unresolved');
  });

  it('blames the data, not the role, when a selected task is not in the list', () => {
    tasksResult.data = [{ id: '1', projectId: '900' }];
    storeState.selectedTaskIds = ['1', 'missing'];
    restrictingRole.value = 'admin';

    render(<BulkEditBar />);

    expect(markDone()).toBeDisabled();
    expect(markDone().getAttribute('title')).toBe('flow.bulkEdit.unresolved');
    expect(deleteButton().getAttribute('title')).toBe('flow.bulkEdit.unresolved');
  });

  it('still says view-only when the selection resolves and the role is the reason', () => {
    tasksResult.data = [{ id: '1', projectId: '900' }];
    storeState.selectedTaskIds = ['1'];
    restrictingRole.value = 'viewer';

    render(<BulkEditBar />);

    expect(markDone()).toBeDisabled();
    expect(markDone().getAttribute('title')).toBe('flowCommon.readOnlyAccess');
    expect(deleteButton().getAttribute('title')).toBe('flowCommon.deleteRestricted');
  });

  it('renders nothing without a selection', () => {
    storeState.selectedTaskIds = [];

    const { container } = render(<BulkEditBar />);

    expect(container).toBeEmptyDOMElement();
  });
});
