import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { initTestI18n } from '../../shared/testing/i18n';
import type { ApiTask } from '../../shared/services/flowApi';
import ProjectList from './ProjectList';
import { useFlowStore } from '../../shared/stores/flowStore';

const remote = vi.hoisted(() => ({
  tasks: [] as ApiTask[],
  taskLoading: false,
  columnsFailed: false,
  retryTasks: vi.fn(),
  retryColumns: vi.fn(),
  retryProject: vi.fn(),
  openTask: vi.fn(),
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
vi.mock('./useColumns', () => ({
  useProjectColumns: () => ({
    data: [{ id: '9', name: 'Release review', sortOrder: 0 }],
    isLoading: false,
    isError: remote.columnsFailed,
    refetch: remote.retryColumns,
  }),
}));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({
    data: remote.tasks,
    isLoading: remote.taskLoading,
    isPending: remote.taskLoading,
    isError: false,
    refetch: remote.retryTasks,
  }),
  useUpdateTaskStatus: () => ({ mutate: vi.fn() }),
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
  remote.taskLoading = false;
  remote.columnsFailed = false;
  remote.tasks = [
    {
      id: '1',
      projectId: '7',
      sectionId: '9',
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
  it('groups persisted tasks under the server column', () => {
    renderList();
    expect(screen.getByText('Release review')).toBeInTheDocument();
    expect(screen.getByText('Persisted task')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'In Progress' })).not.toBeInTheDocument();
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
  it('shows loading while tasks are pending', () => {
    remote.taskLoading = true;
    renderList();
    expect(screen.getByText('Loading project…')).toBeInTheDocument();
    expect(screen.queryByText('Release review')).not.toBeInTheDocument();
  });
  it('retries the project, columns and tasks when columns fail', () => {
    remote.columnsFailed = true;
    renderList();
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(remote.retryProject).toHaveBeenCalledOnce();
    expect(remote.retryTasks).toHaveBeenCalledOnce();
    expect(remote.retryColumns).toHaveBeenCalledOnce();
  });
});
