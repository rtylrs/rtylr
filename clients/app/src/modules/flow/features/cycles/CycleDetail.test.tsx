import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiCycle, ApiTask } from '../../shared/services/flowApi';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  deleteMutate: vi.fn(),
  canDelete: false,
  cycle: undefined as unknown,
  tasks: [] as unknown[],
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts && 'count' in opts ? `${key}:${String(opts.count)}` : key,
    i18n: { language: 'en-US' },
  }),
}));
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ cycleId: '7' }),
  useNavigate: () => mocks.navigate,
  Link: ({ children }: { children: React.ReactNode }) => <a>{children}</a>,
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('./useCycles', () => ({
  useCycle: () => ({ data: mocks.cycle, isPending: false, isError: false }),
  useCycles: () => ({ data: [] }),
  useUpdateCycle: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteCycle: () => ({ mutate: mocks.deleteMutate, isPending: false }),
}));
vi.mock('../tasks/useTasks', () => ({ useTasks: () => ({ data: mocks.tasks }) }));
vi.mock('../projects/useStatuses', () => ({ useProjectStatuses: () => ({ data: [] }) }));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanEditTask: () => true,
  useCanAdministerProject: () => mocks.canDelete,
}));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('../../shared/components/TaskRow', () => ({
  default: ({ task }: { task: { title: string } }) => <div>{task.title}</div>,
}));

import CycleDetail from './CycleDetail';

function task(over: Partial<ApiTask>): ApiTask {
  return {
    id: '1',
    projectId: '3',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'RTYLR',
    ...over,
  };
}

const sprint: ApiCycle = {
  id: '7',
  projectId: '3',
  name: 'C1',
  startDate: '2026-11-16',
  endDate: '2026-11-29',
  status: 'active',
  scope: 2,
  completed: 1,
};

beforeEach(() => {
  vi.clearAllMocks();
  mocks.canDelete = false;
  mocks.cycle = sprint;
  mocks.tasks = [
    task({ id: '1', title: 'Open one', cycleId: '7' }),
    task({
      id: '2',
      title: 'Shipped',
      cycleId: '7',
      status: 'done',
      completedAt: '2026-11-17T10:00:00Z',
    }),
    task({ id: '3', title: 'Other cycle', cycleId: '8' }),
    task({ id: '4', title: 'Shelved', cycleId: '7', archived: true }),
    task({ id: '5', title: 'Dropped', cycleId: '7', status: 'cancelled' }),
  ];
});

describe('CycleDetail', () => {
  it("lists the cycle's open and done tasks, leaving out other cycles, archived and cancelled", () => {
    render(<CycleDetail />);
    expect(screen.getByText('C1')).toBeInTheDocument();
    expect(screen.getByText('flow.backlog.status.active')).toBeInTheDocument();
    expect(screen.getByText('Open one')).toBeInTheDocument();
    expect(screen.getByText('Shipped')).toBeInTheDocument();
    expect(screen.queryByText('Other cycle')).not.toBeInTheDocument();
    expect(screen.queryByText('Shelved')).not.toBeInTheDocument();
    expect(screen.queryByText('Dropped')).not.toBeInTheDocument();
    expect(screen.getByText('flow.cycles.detail.openTasks:1')).toBeInTheDocument();
    expect(screen.getByText('50%')).toBeInTheDocument();
  });

  it('offers delete to project admins only, and leaves for the list after deleting', () => {
    const { unmount } = render(<CycleDetail />);
    expect(screen.queryByText('flow.cycles.detail.delete')).not.toBeInTheDocument();
    unmount();

    mocks.canDelete = true;
    mocks.deleteMutate.mockImplementation((_id: string, opts: { onSuccess: () => void }) =>
      opts.onSuccess()
    );
    render(<CycleDetail />);
    fireEvent.click(screen.getByText('flow.cycles.detail.delete'));
    // Every member goes back to the backlog, archived and cancelled ones too.
    expect(screen.getByText('flow.cycles.detail.deleteDescription:4')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flowCommon.delete' }));
    expect(mocks.deleteMutate).toHaveBeenCalledWith('7', expect.anything());
    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/flow/cycles' });
  });
});
