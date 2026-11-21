import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import ProjectBoard from './ProjectBoard';
import { initTestI18n } from '../../shared/testing/i18n';

const mutation = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));
const status = vi.hoisted(() => (id: string, name: string, bucket: string, sortOrder: number) => ({
  id,
  projectId: '2',
  name,
  bucket,
  legacyStatus: bucket === 'done' ? 'done' : 'todo',
  color: '#8b5cf6',
  sortOrder,
  isDefault: sortOrder === 0,
  createdAt: '',
}));
const taskData = vi.hoisted(() => ({ rows: [] as unknown[] }));
const saveAsView = vi.hoisted(() => ({ save: vi.fn(), pending: false }));
// Custom fields come from their own query; none are defined here.
vi.mock('./useFields', () => ({
  useProjectFields: () => ({ data: [] }),
  useSetTaskFieldValues: () => ({ mutate: vi.fn(), isPending: false }),
}));
vi.mock('../views/useSaveFiltersAsView', () => ({
  projectFiltersForView: (filters: unknown) => filters,
  useSaveFiltersAsView: () => saveAsView,
}));
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: '2' }),
  useLocation: () => ({ pathname: '/flow/p/2' }),
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@voxire/config', () => ({
  useFeatureFlag: () => false,
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
}));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanCreateTask: () => true,
  useCanEditTask: () => true,
  useCanDeleteResource: () => true,
  useCanAdministerProject: () => true,
}));
vi.mock('../../shared/lib/useOpenTask', () => ({ useOpenTask: () => vi.fn() }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: '2', name: 'Release project', color: '#6366f1' },
    isLoading: false,
    loadFailed: false,
  }),
}));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({ isLoading: false, isError: false, data: taskData.rows }),
}));
vi.mock('../tasks/useLabels', () => ({ useLabels: () => ({ data: [] }) }));
vi.mock('./useColumns', () => ({ useMoveTaskToColumn: () => mutation }));
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

vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({
    isPending: false,
    isError: false,
    data: [status('9', 'Review', 'active', 0), status('11', 'Done', 'done', 900)],
  }),
  useCreateStatus: () => mutation,
  useUpdateStatus: () => mutation,
  useReorderStatuses: () => mutation,
  useDeleteStatus: () => mutation,
}));

const daysAgo = (days: number): string => new Date(Date.now() - days * 86_400_000).toISOString();
const task = (id: string, title: string, statusId: string, closedDaysAgo: number) => ({
  id,
  projectId: '2',
  statusId,
  sectionId: statusId,
  title,
  status: statusId === '11' ? 'done' : 'todo',
  priority: 'medium',
  assignee: '',
  assigneeInitials: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Release project',
  completedAt: statusId === '11' ? daysAgo(closedDaysAgo) : undefined,
  updatedAt: daysAgo(closedDaysAgo),
});

const renderBoard = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProjectBoard />
    </QueryClientProvider>
  );

const openColumnMenu = async (name: string): Promise<void> => {
  // Radix opens a dropdown on pointerdown, not click.
  fireEvent.pointerDown(screen.getByRole('button', { name: `Actions for ${name}` }), {
    button: 0,
    ctrlKey: false,
  });
  await screen.findByRole('menu');
};

beforeEach(async () => {
  await initTestI18n();
  vi.clearAllMocks();
  taskData.rows = [
    task('1', 'Still in review', '9', 40),
    task('2', 'Shipped this week', '11', 3),
    task('3', 'Shipped last month', '11', 30),
  ];
});

describe('Board done column window', () => {
  it('shows only cards closed in the last 14 days until Show older is chosen', () => {
    renderBoard();
    expect(screen.getByText('Shipped this week')).toBeInTheDocument();
    expect(screen.getByText('Still in review')).toBeInTheDocument();
    expect(screen.queryByText('Shipped last month')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Show older (1)' }));

    expect(screen.getByText('Shipped last month')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Show older/ })).not.toBeInTheDocument();
  });
});

describe('Board column menu', () => {
  it('opens the delete dialog only from the explicit menu item', async () => {
    renderBoard();
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    await openColumnMenu('Review');
    fireEvent.click(screen.getByRole('menuitem', { name: 'Delete status…' }));

    expect(await screen.findByRole('dialog', { name: 'Delete Review' })).toBeInTheDocument();
    expect(
      screen.getByText('1 task is in this status. Choose the status it moves to.')
    ).toBeInTheDocument();
  });

  it('disables deleting the only done status and explains why', async () => {
    renderBoard();
    await openColumnMenu('Done');
    const remove = screen.getByRole('menuitem', { name: 'Delete status…' });

    expect(remove).toHaveAttribute('data-disabled');
    expect(remove).toHaveAccessibleDescription(
      'A project needs at least one done status, so this one cannot be deleted.'
    );
  });

  it('moves a column with one reorder of the whole set', async () => {
    renderBoard();
    await openColumnMenu('Review');
    fireEvent.click(screen.getByRole('menuitem', { name: 'Move right' }));

    expect(mutation.mutate).toHaveBeenCalledTimes(1);
    expect(mutation.mutate).toHaveBeenCalledWith(['11', '9']);
  });

  it('inserts with no gap by creating, then placing with one reorder', async () => {
    renderBoard();
    await openColumnMenu('Review');
    fireEvent.click(screen.getByRole('menuitem', { name: 'Add status to the left' }));
    const dialog = await screen.findByRole('dialog');
    fireEvent.change(within(dialog).getByLabelText('Status name'), {
      target: { value: 'Triage' },
    });
    fireEvent.click(within(dialog).getByRole('button', { name: 'Add status' }));

    expect(mutation.mutate).toHaveBeenCalledTimes(1);
    expect(mutation.mutate).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Triage',
        placeAmong: { ids: ['9', '11'], index: 0 },
      }),
      expect.anything()
    );
    expect(mutation.mutate.mock.calls[0][0]).not.toHaveProperty('sortOrder');
  });
});
