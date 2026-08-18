import React from 'react';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiTask, ApiTaskDetail } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: 't-1' }),
  useNavigate: () => vi.fn(),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

function apiTask(overrides: Partial<ApiTask> = {}): ApiTask {
  return {
    id: 't-1',
    projectId: 'p-9',
    workspaceId: 'ws-9',
    title: 'Cross-workspace task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Other workspace project',
    ...overrides,
  };
}

const taskDetailData = vi.hoisted(() => ({
  data: undefined as ApiTaskDetail | undefined,
  isLoading: false,
}));
vi.mock('./useTask', () => ({
  useTask: () => taskDetailData,
}));

vi.mock('./useTasks', () => ({
  useUpdateTask: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteTask: () => ({ mutate: vi.fn(), isPending: false }),
  useAddComment: () => ({ mutate: vi.fn(), isPending: false }),
  useAddSubtask: () => ({ mutate: vi.fn(), isPending: false }),
  useToggleSubtask: () => ({ mutate: vi.fn(), isPending: false }),
  useTaskAttachments: () => ({ data: [] }),
  useAddAttachment: () => ({ mutate: vi.fn(), isPending: false }),
  useRemoveAttachment: () => ({ mutate: vi.fn(), isPending: false }),
}));

const useLabelsSpy = vi.hoisted(() => vi.fn(() => ({ data: [] })));
vi.mock('./useLabels', () => ({
  useLabels: useLabelsSpy,
  useSetTaskLabels: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('./useTaskActivity', () => ({
  useTaskActivity: () => ({ data: [] }),
}));

vi.mock('../../shared/lib/useTaskRealtime', () => ({
  useTaskRealtime: () => {},
}));

const useCurrentProjectSpy = vi.hoisted(() =>
  vi.fn(() => ({ project: undefined, isLoading: false }))
);
vi.mock('../projects/useCurrentProject', () => ({
  useCurrentProject: useCurrentProjectSpy,
}));

vi.mock('./components/TimePanel', () => ({
  default: () => null,
}));

vi.mock('../../shared/lib/currentUser', () => ({
  useCanEditTask: () => true,
  useCanDeleteResource: () => true,
}));

vi.mock('../../shared/components/Breadcrumbs', () => ({
  default: ({ items }: { items: { label: string }[] }) => (
    <>{items.map((item) => item.label).join(' / ')}</>
  ),
}));
vi.mock('../../shared/components/CommentList', () => ({ default: () => null }));
vi.mock('../../shared/components/CommentInput', () => ({ default: () => null }));
vi.mock('../../shared/components/SubtaskList', () => ({ default: () => null }));
vi.mock('../../shared/components/AttachmentUpload', () => ({ default: () => null }));
vi.mock('../../shared/components/AssigneeAvatar', () => ({ default: () => null }));
vi.mock('../../shared/components/LabelChip', () => ({ default: () => null }));
vi.mock('../../shared/components/DueDatePicker', () => ({ default: () => null }));

import TaskDetail from './TaskDetail';
import { useFlowStore } from '../../shared/stores/flowStore';

function renderTaskDetail() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <TaskDetail />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useLabelsSpy.mockReturnValue({ data: [] });
  useCurrentProjectSpy.mockReturnValue({ project: undefined, isLoading: false });
  useFlowStore.setState({
    activeWorkspaceId: 'ws-active',
    workspaces: [{ id: 'ws-active', name: 'Active workspace', color: '#000' }],
    projects: [],
  });
  taskDetailData.data = undefined;
  taskDetailData.isLoading = false;
});

describe('TaskDetail cross-workspace sync', () => {
  it("syncs the shared store's active workspace to the task's own workspace", () => {
    taskDetailData.data = {
      task: apiTask({ workspaceId: 'ws-9' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    expect(useFlowStore.getState().activeWorkspaceId).toBe('ws-9');
  });

  it('does not overwrite the active workspace for a personal-sentinel task', () => {
    taskDetailData.data = {
      task: apiTask({ workspaceId: 'personal' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    expect(useFlowStore.getState().activeWorkspaceId).toBe('ws-active');
  });

  it("scopes the labels query to the task's own workspace, not the globally-active one", () => {
    taskDetailData.data = {
      task: apiTask({ workspaceId: 'ws-9' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    expect(useLabelsSpy).toHaveBeenCalledWith('ws-9');
    expect(useLabelsSpy).not.toHaveBeenCalledWith('ws-active');
  });

  it("scopes the project lookup to the task's own workspace, not the globally-active one", () => {
    taskDetailData.data = {
      task: apiTask({ workspaceId: 'ws-9', projectId: 'p-9' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    expect(useCurrentProjectSpy).toHaveBeenCalledWith('ws-9', 'p-9');
  });

  it("falls back to the task's own projectName when the project isn't in the store", () => {
    taskDetailData.data = {
      task: apiTask({
        workspaceId: 'ws-9',
        projectId: 'p-9',
        projectName: 'Other workspace project',
      }),
      subtasks: [],
      comments: [],
      activity: [],
    };
    useCurrentProjectSpy.mockReturnValue({ project: undefined, isLoading: false });

    renderTaskDetail();

    expect(
      screen.getByText((_, element) => element?.textContent === 'Other workspace project / t-1')
    ).toBeTruthy();
  });
});
