import React from 'react';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { chooseFlowOption } from '../../../../../../test/flowSelect';
import type { ApiTask, ApiTaskRelationGroup } from '../../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, string>) =>
      opts?.identifier ? `${key}:${opts.identifier}` : key,
    i18n: { language: 'en' },
  }),
}));

const state = vi.hoisted(() => ({
  groups: [] as ApiTaskRelationGroup[],
  results: [] as ApiTask[],
  searchedWith: [] as Array<[string, boolean]>,
}));
const addMutate = vi.hoisted(() => vi.fn());
const removeMutate = vi.hoisted(() => vi.fn());
const closeMutate = vi.hoisted(() => vi.fn());

vi.mock('../useTaskRelations', () => ({
  useTaskRelations: () => ({ data: state.groups, isError: false }),
  useAddTaskRelation: () => ({ mutate: addMutate, isPending: false }),
  useRemoveTaskRelation: () => ({ mutate: removeMutate, isPending: false }),
}));
vi.mock('../useTasks', () => ({
  useUpdateTaskStatus: () => ({ mutate: closeMutate, isPending: false }),
}));
vi.mock('@tanstack/react-query', () => ({
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}));
vi.mock('../../../shared/hooks/queries/useTaskSearch', async () => {
  const actual = await vi.importActual<
    typeof import('../../../shared/hooks/queries/useTaskSearch')
  >('../../../shared/hooks/queries/useTaskSearch');
  return {
    taskIdFromKey: actual.taskIdFromKey,
    useTaskSearch: (query: string, enabled: boolean) => {
      state.searchedWith.push([query, enabled]);
      return { data: enabled ? state.results : undefined, isFetching: false };
    },
  };
});
vi.mock('../../../shared/components/ConfirmDialog', () => ({
  default: ({
    open,
    title,
    confirmLabel,
    onConfirm,
  }: {
    open: boolean;
    title: string;
    confirmLabel: string;
    onConfirm: () => void;
  }) =>
    open ? (
      <div role="dialog" aria-label={title}>
        <button type="button" onClick={onConfirm}>
          {confirmLabel}
        </button>
      </div>
    ) : null,
}));

import RelationsSection from './RelationsSection';

const task = (id: string, title: string): ApiTask => ({
  id,
  identifier: `T-${id}`,
  projectId: '2',
  title,
  status: 'todo',
  priority: 'medium',
  assignee: '',
  assigneeInitials: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Project',
});

const related = (id: string, title: string, statusBucket: 'active' | 'done' = 'active') => ({
  relationId: `r${id}`,
  taskId: id,
  identifier: `T-${id}`,
  title,
  projectId: '2',
  projectName: 'Project',
  status: statusBucket === 'done' ? 'done' : 'todo',
  statusName: '',
  statusColor: '',
  statusBucket,
  assigneeName: '',
});

const renderSection = (canEdit = true) => {
  const onOpenTask = vi.fn();
  render(
    <RelationsSection taskId="1" identifier="T-1" canEdit={canEdit} onOpenTask={onOpenTask} />
  );
  return onOpenTask;
};

describe('RelationsSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    state.groups = [];
    state.results = [];
    state.searchedWith = [];
  });

  it('lists relations under their kind, from this task side, and opens the other task', () => {
    state.groups = [
      { type: 'blocks', direction: 'inward', relation: 'blocked_by', tasks: [related('7', 'API')] },
      {
        type: 'duplicates',
        direction: 'inward',
        relation: 'duplicated_by',
        tasks: [related('9', 'Old copy', 'done')],
      },
    ];
    const onOpenTask = renderSection();
    // Group headings; the relation type picker repeats the default kind's label.
    expect(
      screen.getByText('flow.taskDetail.relationKinds.blocked_by', { selector: 'p' })
    ).toBeInTheDocument();
    expect(
      screen.getByText('flow.taskDetail.relationKinds.duplicated_by', { selector: 'p' })
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'T-7 · API' }));
    expect(onOpenTask).toHaveBeenCalledWith('7');
    // A closed task stays listed, struck through.
    expect(screen.getByRole('button', { name: 'T-9 · Old copy' })).toHaveStyle({
      textDecoration: 'line-through',
    });
  });

  it('removes a relation by its id and names both ends', () => {
    state.groups = [
      {
        type: 'relates',
        direction: 'outward',
        relation: 'relates_to',
        tasks: [related('7', 'API')],
      },
    ];
    renderSection();
    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.removeRelation:T-7' }));
    expect(removeMutate).toHaveBeenCalledWith({ taskId: '1', relationId: 'r7', otherTaskId: '7' });
  });

  it('adds "blocked by" by default from a title search, leaving this task out', () => {
    state.results = [task('1', 'Login page'), task('4', 'Login API'), task('5', 'Billing')];
    renderSection();
    fireEvent.change(screen.getByLabelText('flow.taskDetail.relationSearch'), {
      target: { value: 'login' },
    });
    expect(screen.queryByRole('button', { name: 'T-1 · Login page' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'T-5 · Billing' })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'T-4 · Login API' }));
    expect(addMutate).toHaveBeenCalledWith(
      { taskId: '1', otherTaskId: '4', type: 'blocks', direction: 'inward' },
      expect.anything()
    );
  });

  it('finds a task by its key', () => {
    state.results = [task('42', 'Something unrelated to the query text')];
    renderSection();
    fireEvent.change(screen.getByLabelText('flow.taskDetail.relationSearch'), {
      target: { value: 'T-42' },
    });
    expect(state.searchedWith.at(-1)).toEqual(['T-42', true]);
    expect(
      screen.getByRole('button', { name: 'T-42 · Something unrelated to the query text' })
    ).toBeInTheDocument();
  });

  it('offers to close the duplicate once it is marked', async () => {
    state.results = [task('4', 'Original')];
    renderSection();
    await chooseFlowOption(
      'flow.taskDetail.relationType',
      'flow.taskDetail.relationKinds.duplicates'
    );
    fireEvent.change(screen.getByLabelText('flow.taskDetail.relationSearch'), {
      target: { value: 'orig' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'T-4 · Original' }));
    expect(addMutate).toHaveBeenCalledWith(
      { taskId: '1', otherTaskId: '4', type: 'duplicates', direction: 'outward' },
      expect.anything()
    );
    // This task duplicates T-4, so this task is the one offered for closing.
    const { onSuccess } = addMutate.mock.calls[0]![1] as { onSuccess: () => void };
    act(() => onSuccess());
    expect(
      screen.getByRole('dialog', { name: 'flow.taskDetail.closeDuplicateTitle:T-1' })
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.closeDuplicateConfirm' }));
    expect(closeMutate).toHaveBeenCalledWith({ id: '1', status: 'cancelled' }, expect.anything());
  });

  it('offers to close the other task when it is the duplicate', async () => {
    state.results = [task('4', 'Copy')];
    renderSection();
    await chooseFlowOption(
      'flow.taskDetail.relationType',
      'flow.taskDetail.relationKinds.duplicated_by'
    );
    fireEvent.change(screen.getByLabelText('flow.taskDetail.relationSearch'), {
      target: { value: 'copy' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'T-4 · Copy' }));
    const { onSuccess } = addMutate.mock.calls[0]![1] as { onSuccess: () => void };
    act(() => onSuccess());
    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.closeDuplicateConfirm' }));
    expect(closeMutate).toHaveBeenCalledWith({ id: '4', status: 'cancelled' }, expect.anything());
  });

  it('is read only without edit rights', () => {
    state.groups = [
      { type: 'blocks', direction: 'outward', relation: 'blocks', tasks: [related('7', 'API')] },
    ];
    renderSection(false);
    expect(screen.queryByLabelText('flow.taskDetail.relationSearch')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.taskDetail.remove')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'T-7 · API' })).toBeInTheDocument();
  });
});
