import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { toast } from 'sonner';
import { initTestI18n } from '../../shared/testing/i18n';
import type { ApiStatus, ApiTask } from '../../shared/services/flowApi';
import ProjectList from './ProjectList';
import { useFlowStore } from '../../shared/stores/flowStore';
import { forgetPreviousStatuses } from '../tasks/useToggleDone';
import { chooseFlowOption, flowSelectTrigger } from '../../../../../test/flowSelect';

const remote = vi.hoisted(() => ({
  tasks: [] as ApiTask[],
  taskLoading: false,
  columnsFailed: false,
  setStatus: vi.fn(),
  retryTasks: vi.fn(),
  retryColumns: vi.fn(),
  retryProject: vi.fn(),
  openTask: vi.fn(),
}));
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
        epics: [],
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
vi.mock('../../shared/lib/useOpenTask', () => ({ useOpenTask: () => remote.openTask }));
vi.mock('@tanstack/react-router', () => ({
  useLocation: () => ({ pathname: '/flow/p/7/list' }),
  useParams: () => ({ projectId: '7' }),
  useNavigate: () => vi.fn(),
  Link: ({ children }: React.PropsWithChildren) => <>{children}</>,
}));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanEditTask: () => true,
  useCanCreateTask: () => true,
  useCanAdministerProject: () => true,
}));
vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: unknown) => unknown) =>
    selector({ user: { effectiveAccess: { isOwner: true, subscribedApps: ['flow'] } } }),
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: '7', name: 'Remote project', color: '#6366f1' },
    isLoading: false,
    loadFailed: false,
    refetch: remote.retryProject,
  }),
}));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
const status = (
  id: string,
  name: string,
  bucket: ApiStatus['bucket'],
  sortOrder: number,
  isDefault = false
): ApiStatus => ({
  id,
  projectId: '7',
  name,
  bucket,
  legacyStatus: bucket === 'done' ? 'done' : 'todo',
  color: '#6366f1',
  sortOrder,
  isDefault,
  createdAt: '',
});
const STATUSES = [
  status('8', 'Ideas', 'backlog', 0, true),
  status('9', 'Release review', 'active', 10),
  status('11', 'Shipped', 'done', 900),
];
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({
    data: STATUSES,
    isPending: false,
    isError: remote.columnsFailed,
    refetch: remote.retryColumns,
  }),
  useSetTaskStatus: () => ({ mutate: remote.setStatus }),
}));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({
    data: remote.tasks,
    isLoading: remote.taskLoading,
    isPending: remote.taskLoading,
    isError: false,
    refetch: remote.retryTasks,
  }),
}));

const renderList = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProjectList />
    </QueryClientProvider>
  );
beforeEach(async () => {
  await initTestI18n();
  vi.clearAllMocks();
  forgetPreviousStatuses();
  localStorage.clear();
  remote.setStatus.mockImplementation((_vars, options?: { onSuccess?: () => void }) =>
    options?.onSuccess?.()
  );
  remote.taskLoading = false;
  remote.columnsFailed = false;
  remote.tasks = [
    {
      id: '1',
      projectId: '7',
      sectionId: '9',
      statusId: '9',
      title: 'Persisted task',
      status: 'todo',
      priority: 'medium',
      assignee: '',
      assigneeInitials: '',
      dueDate: '',
      subtaskCount: 0,
      projectName: 'Remote project',
    },
  ];
});
describe('Project list server data', () => {
  it('groups tasks by their status and shows only statuses that hold tasks', () => {
    renderList();
    expect(screen.getByText('Release review')).toBeInTheDocument();
    expect(screen.getByText('Persisted task')).toBeInTheDocument();
    expect(screen.queryByText('Ideas')).not.toBeInTheDocument();
    expect(screen.queryByText('Shipped')).not.toBeInTheDocument();
    expect(screen.queryByText('Other')).not.toBeInTheDocument();
    // The last "New task" is the one in the section, not the header's, and its
    // column id must be the section's - stringified, because quick-create and
    // the board's grouping both key columns by string id.
    fireEvent.click(screen.getAllByRole('button', { name: 'New task' }).at(-1)!);
    expect(useFlowStore.getState().quickCreateColumnId).toBe('9');
  });
  it('opens a row with the keyboard without opening from its completion control', () => {
    renderList();
    const row = screen.getByRole('button', { name: /Open task.*Persisted task/ });
    fireEvent.keyDown(row, { key: 'Enter' });
    expect(remote.openTask).toHaveBeenCalledWith('1');
    remote.openTask.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'Mark done' }));
    expect(remote.openTask).not.toHaveBeenCalled();
  });
  it('puts a task whose status is not in the set under its matching status, not Other', () => {
    remote.tasks = [{ ...remote.tasks[0], statusId: '404', sectionId: '404', status: 'done' }];
    renderList();
    expect(screen.getByText('Shipped')).toBeInTheDocument();
    expect(screen.queryByText('Other')).not.toBeInTheDocument();
  });
  it('ticks a task into the first done status and Undo puts it back', () => {
    renderList();
    fireEvent.click(screen.getByRole('button', { name: 'Mark done' }));
    expect(remote.setStatus).toHaveBeenCalledWith(
      { taskId: '1', statusId: '11' },
      expect.any(Object)
    );
    const [message, options] = vi.mocked(toast.success).mock.calls[0] as [
      string,
      { duration: number; action: { label: string; onClick: () => void } },
    ];
    expect(message).toBe('Persisted task marked done');
    expect(options.duration).toBe(5000);
    expect(options.action.label).toBe('Undo');
    options.action.onClick();
    expect(remote.setStatus).toHaveBeenLastCalledWith(
      { taskId: '1', statusId: '9' },
      expect.any(Object)
    );
  });
  it('unticks a done task back to the project default when its earlier status is unknown', () => {
    remote.tasks = [{ ...remote.tasks[0], statusId: '11', sectionId: '11', status: 'done' }];
    renderList();
    fireEvent.click(screen.getByRole('button', { name: 'Mark done' }));
    expect(remote.setStatus).toHaveBeenCalledWith(
      { taskId: '1', statusId: '8' },
      expect.any(Object)
    );
  });
  it('regroups by the chosen field and remembers the choice for the project', async () => {
    remote.tasks = [
      { ...remote.tasks[0], priority: 'high' },
      { ...remote.tasks[0], id: '2', title: 'Second task', priority: 'low' },
    ];
    const { unmount } = renderList();
    expect(flowSelectTrigger('Group by')).toHaveAccessibleName('Group by: Status');
    await chooseFlowOption('Group by', 'Priority');
    expect(screen.getByText('High')).toBeInTheDocument();
    expect(screen.getByText('Low')).toBeInTheDocument();
    expect(screen.queryByText('Release review')).not.toBeInTheDocument();
    expect(localStorage.getItem('flow:list:groupBy:7')).toBe('priority');
    unmount();
    renderList();
    expect(flowSelectTrigger('Group by')).toHaveAccessibleName('Group by: Priority');
  });
  it('shows loading while tasks are pending', () => {
    remote.taskLoading = true;
    renderList();
    expect(screen.getByText('Loading project…')).toBeInTheDocument();
    expect(screen.queryByText('Release review')).not.toBeInTheDocument();
  });
  it('retries the project, statuses and tasks when statuses fail', () => {
    remote.columnsFailed = true;
    renderList();
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(remote.retryProject).toHaveBeenCalledOnce();
    expect(remote.retryTasks).toHaveBeenCalledOnce();
    expect(remote.retryColumns).toHaveBeenCalledOnce();
  });
});
