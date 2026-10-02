import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import ProjectBoard from './ProjectBoard';
import { initTestI18n } from '../../shared/testing/i18n';

const access = vi.hoisted(() => ({ create: false, edit: false, manage: false, open: vi.fn() }));
const mutation = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));
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
vi.mock('../tasks/useLabels', () => ({ useLabels: () => ({ data: [] }) }));
vi.mock('./useColumns', () => ({
  useProjectColumns: () => ({
    isLoading: false,
    isError: false,
    data: [{ id: '9', name: 'Release review', color: '#6366f1', sortOrder: 0 }],
  }),
  useCreateColumn: () => mutation,
  useUpdateColumn: () => mutation,
  useDeleteColumn: () => mutation,
  useMoveTaskToColumn: () => mutation,
}));
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({ data: [], isPending: false, isError: false }),
  useCreateStatus: () => mutation,
  useUpdateStatus: () => mutation,
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
    for (const name of ['New task', 'Add task', 'Add column', 'Delete column', 'Statuses']) {
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
    for (const name of ['New task', 'Add task', 'Add column', 'Delete column', 'Statuses']) {
      expect(screen.getByRole('button', { name, exact: true })).toBeEnabled();
    }
    expect(screen.getByRole('button', { name: 'Release review' })).toBeEnabled();
  });
});
