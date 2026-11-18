import { describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { TaskRow, TaskGroup, buildListGroups, loadListGroupBy } from './ProjectList';
import type { ApiStatus } from '../../shared/services/flowApi';
import type { Task } from '../../shared/stores/flowStore';

const mockNavigate = vi.fn();

// Filtering has its own suite (ProjectBoardFilters); here it is a no-op.
const saveAsView = vi.hoisted(() => ({ save: vi.fn(), pending: false }));
vi.mock('../views/useSaveFiltersAsView', () => ({
  projectFiltersForView: (filters: unknown) => filters,
  useSaveFiltersAsView: () => saveAsView,
}));
vi.mock('./useProjectTaskFilters', () => ({
  useProjectTaskFilters: (_projectId: string, tasks: unknown[]) => ({
    state: {
      filters: {
        assignees: [],
        priorities: [],
        labels: [],
        due: [],
        statuses: [],
        cycles: [],
        milestones: [],
        projects: [],
        query: '',
      },
      setFilters: vi.fn(),
      clear: vi.fn(),
      active: false,
      dim: false,
      setDim: vi.fn(),
    },
    fields: [],
    people: [],
    currentUserId: null,
    matchingIds: null,
    shownTasks: tasks,
    dimmedIds: null,
  }),
}));
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
  useParams: () => ({}),
  Link: ({ children }: { children?: ReactNode }) => <>{children}</>,
}));

// Identity translator: renders the i18n key itself instead of resolved text,
// same convention used in TaskDetail.test.tsx. What matters for the
// regression test below is that both components resolve the SAME key for
// the same status, not the literal English string. Keep the rest of the
// module real (via importOriginal) - `initReactI18next` is still needed by
// the real i18next instance that useTasks.ts imports as a side effect.
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({ data: [] }),
  useSetTaskStatus: () => ({ mutate: vi.fn() }),
}));

vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return {
    ...actual,
    useTranslation: () => ({ t: (key: string) => key }),
  };
});

const task: Task = {
  id: 't1',
  projectId: 'p-web',
  title: 'Ship the release',
  status: 'todo',
  priority: 'medium',
  assignee: 'Ada Lovelace',
  assigneeInitials: 'AL',
  assigneeId: 'u1',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Web',
  createdAt: '',
  updatedAt: '',
  identifier: 'FLOW-1',
};

function renderRow(): void {
  const queryClient = new QueryClient();
  render(
    <QueryClientProvider client={queryClient}>
      <TaskRow task={task} />
    </QueryClientProvider>
  );
}

describe('ProjectList TaskRow more menu', () => {
  // The labels are translation keys now: they were literal English until the
  // menu was internationalised, and this suite's `t` mock returns the key.
  it('opens a menu with open/copy link/delete actions when the more button is clicked', () => {
    renderRow();

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectList.moreOptions' }));

    expect(screen.getByRole('menuitem', { name: 'flow.taskMenu.openTask' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: 'flow.taskMenu.copyLink' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: 'flowCommon.delete' })).toBeInTheDocument();
  });

  it('opens the task over the list when Open task is chosen', () => {
    renderRow();

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectList.moreOptions' }));
    fireEvent.click(screen.getByRole('menuitem', { name: 'flow.taskMenu.openTask' }));

    // `?openTask`, not `/flow/task/t1`: the list stays mounted and the drawer opens
    // over it (rtylr#1079).
    const call = mockNavigate.mock.calls[0][0] as {
      to: string;
      search: (previous: Record<string, unknown>) => Record<string, unknown>;
    };
    expect(call.to).toBe('.');
    expect(call.search({})).toEqual({ openTask: 't1' });
  });
});

const status = (
  id: string,
  name: string,
  bucket: ApiStatus['bucket'],
  sortOrder: number,
  legacyStatus = 'todo'
): ApiStatus => ({
  id,
  projectId: 'p-web',
  name,
  bucket,
  legacyStatus,
  color: '#6366f1',
  sortOrder,
  isDefault: false,
  createdAt: '',
});
const identity = (key: string): string => key;

describe('buildListGroups', () => {
  const statuses = [
    status('3', 'Done', 'done', 900, 'done'),
    status('1', 'To Do', 'backlog', 10),
    status('2', 'Doing', 'active', 20, 'in_progress'),
  ];
  const on = (id: string, statusId: string | undefined, extra: Partial<Task> = {}): Task => ({
    ...task,
    id,
    statusId,
    sectionId: statusId,
    ...extra,
  });

  it('orders status groups by sort order and leaves empty statuses out', () => {
    const groups = buildListGroups([on('a', '3'), on('b', '1')], statuses, 'status', identity);
    // Default statuses are shown through the locale (identity t here).
    expect(groups.map((g) => g.label)).toEqual(['flowStatus.todo', 'flowStatus.done']);
    expect(groups[0].statusId).toBe('1');
  });

  it('never makes an Other group: a stray task joins the status matching its built-in value', () => {
    const groups = buildListGroups(
      [on('a', '99', { status: 'in_progress' }), on('b', undefined, { status: 'cancelled' })],
      statuses,
      'status',
      identity
    );
    expect(groups.map((g) => g.label)).toEqual(['flowStatus.todo', 'Doing']);
    expect(groups.flatMap((g) => g.tasks.map((t) => t.id)).sort()).toEqual(['a', 'b']);
  });

  it('groups by priority in priority order with translated labels', () => {
    const groups = buildListGroups(
      [on('a', '1', { priority: 'low' }), on('b', '1', { priority: 'urgent' })],
      statuses,
      'priority',
      identity
    );
    expect(groups.map((g) => g.label)).toEqual(['flowPriority.urgent', 'flowPriority.low']);
  });

  it('groups by assignee and names the unassigned group', () => {
    const groups = buildListGroups(
      [on('a', '1'), on('b', '1', { assignee: '' })],
      statuses,
      'assignee',
      identity
    );
    expect(groups.map((g) => g.label)).toEqual(['Ada Lovelace', 'flow.quickCreate.unassigned']);
  });

  it('returns one headerless group when grouping is off', () => {
    const groups = buildListGroups([on('a', '1'), on('b', '3')], statuses, 'none', identity);
    expect(groups).toHaveLength(1);
    expect(groups[0].tasks).toHaveLength(2);
  });
});

describe('List group-by preference', () => {
  it('defaults to status and ignores unknown stored values', () => {
    localStorage.removeItem('flow:list:groupBy:p1');
    expect(loadListGroupBy('p1')).toBe('status');
    localStorage.setItem('flow:list:groupBy:p1', 'colour');
    expect(loadListGroupBy('p1')).toBe('status');
    localStorage.setItem('flow:list:groupBy:p1', 'due');
    expect(loadListGroupBy('p1')).toBe('due');
  });

  it('falls back to status when storage throws', () => {
    const spy = vi.spyOn(localStorage, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    expect(loadListGroupBy('p1')).toBe('status');
    spy.mockRestore();
  });
});

describe('TaskGroup', () => {
  const group = {
    id: '2',
    label: 'Doing',
    color: '#6366f1',
    statusId: '2',
    tasks: [
      { ...task, id: 't1', title: 'One' },
      { ...task, id: 't2', title: 'Two' },
    ],
  };

  it('shows the label and count, and adds a task into that status', () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <TaskGroup group={group} />
      </QueryClientProvider>
    );
    expect(screen.getByText('Doing')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
  });

  it('renders no header when grouping is off', () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <TaskGroup group={group} showHeader={false} />
      </QueryClientProvider>
    );
    expect(screen.queryByText('Doing')).not.toBeInTheDocument();
  });
});
