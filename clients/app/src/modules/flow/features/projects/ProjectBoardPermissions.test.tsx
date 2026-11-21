import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import ProjectBoard from './ProjectBoard';
import { initTestI18n } from '../../shared/testing/i18n';

const access = vi.hoisted(() => ({ create: false, edit: false, manage: false, open: vi.fn() }));
const mutation = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));
const status = vi.hoisted(() => (id: string, name: string, sortOrder: number) => ({
  id,
  projectId: '2',
  name,
  bucket: 'active',
  legacyStatus: 'todo',
  color: '#6366f1',
  sortOrder,
  isDefault: sortOrder === 0,
  createdAt: '',
}));
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
  useCanCreateTask: () => access.create,
  useCanEditTask: () => access.edit,
  useCanDeleteResource: () => access.manage,
  useCanAdministerProject: () => access.manage,
}));
vi.mock('../../shared/lib/useOpenTask', () => ({ useOpenTask: () => access.open }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: '2', name: 'Release project', color: '#6366f1' },
    isLoading: false,
    loadFailed: false,
  }),
}));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({
    isLoading: false,
    isError: false,
    data: [
      {
        id: '3',
        projectId: '2',
        sectionId: '9',
        title: 'Review the release',
        status: 'todo',
        priority: 'medium',
        assignee: '',
        assigneeInitials: '',
        dueDate: '',
        subtaskCount: 0,
        projectName: 'Release project',
      },
    ],
  }),
}));
// Cards render the blocked badge, whose relations hook reads the Flow API.
vi.mock('../tasks/useTaskRelations', () => ({
  useTaskRelations: () => ({ data: undefined, isPending: true }),
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

vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({
    isPending: false,
    isError: false,
    data: [status('9', 'Release review', 0)],
  }),
  useCreateStatus: () => mutation,
  useUpdateStatus: () => mutation,
  useReorderStatuses: () => mutation,
  useDeleteStatus: () => mutation,
}));

const renderBoard = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProjectBoard />
    </QueryClientProvider>
  );
beforeEach(async () => {
  await initTestI18n();
  vi.clearAllMocks();
  access.create = access.edit = access.manage = false;
});

describe('Project board permissions', () => {
  it('lets a viewer open tasks without offering task or board mutations', () => {
    renderBoard();
    for (const name of [
      'New task',
      'Add task',
      'Add status',
      'Actions for Release review',
      'Statuses',
    ]) {
      expect(screen.queryByRole('button', { name, exact: true })).not.toBeInTheDocument();
    }
    expect(screen.getByRole('button', { name: 'Release review' })).toBeDisabled();
    fireEvent.click(screen.getByRole('button', { name: /Open task.*Review the release/ }));
    expect(access.open).toHaveBeenCalledWith('3');
    expect(mutation.mutate).not.toHaveBeenCalled();
  });

  it('offers task and board mutations to an authorized project administrator', () => {
    access.create = access.edit = access.manage = true;
    renderBoard();
    for (const name of [
      'New task',
      'Add task',
      'Add status',
      'Actions for Release review',
      'Statuses',
    ]) {
      expect(screen.getByRole('button', { name, exact: true })).toBeEnabled();
    }
    expect(screen.getByRole('button', { name: 'Release review' })).toBeEnabled();
  });
});
