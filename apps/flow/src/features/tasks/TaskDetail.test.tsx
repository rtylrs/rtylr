import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
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

const updateTaskMutate = vi.hoisted(() => vi.fn());
vi.mock('./useTasks', () => ({
  useUpdateTask: () => ({ mutate: updateTaskMutate, isPending: false }),
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
vi.mock('../../shared/hooks/queries/useApprovedLeaveInRange', () => ({
  useApprovedLeaveInRange: vi.fn(() => ({ data: null })),
}));

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
// Renders the label name, so the tests can assert which labels reached it.
vi.mock('../../shared/components/LabelChip', () => ({
  default: ({ label }: { label: { name: string } }) => <span>{label.name}</span>,
}));
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
  updateTaskMutate.mockClear();
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

  // Labels are one business-wide pool now, so there is nothing to scope the
  // query by. This used to assert the opposite: that the query followed the
  // task's own workspace rather than the active one.
  it('asks for labels without scoping them to a workspace', () => {
    taskDetailData.data = {
      task: apiTask({ workspaceId: 'ws-9' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    expect(useLabelsSpy).toHaveBeenCalledWith();
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

describe('TaskDetail description sentinel display', () => {
  it('renders a "-" sentinel description as an empty field with placeholder text, not a literal dash', () => {
    taskDetailData.data = {
      task: apiTask({ description: '-' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    const descriptionField = screen.getByPlaceholderText('flow.taskDetail.descriptionPlaceholder');
    expect(descriptionField).toHaveValue('');
  });

  it('renders a real description normally', () => {
    taskDetailData.data = {
      task: apiTask({ description: 'Investigate the flaky checkout test' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    const descriptionField = screen.getByPlaceholderText('flow.taskDetail.descriptionPlaceholder');
    expect(descriptionField).toHaveValue('Investigate the flaky checkout test');
  });

  it('does not send an update when the sentinel-stripped description is only displayed, never edited', () => {
    taskDetailData.data = {
      task: apiTask({ description: '-' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    expect(updateTaskMutate).not.toHaveBeenCalled();
  });

  it('sends a typed description as-is when editing a task that had no real description', () => {
    taskDetailData.data = {
      task: apiTask({ description: '-' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    fireEvent.change(screen.getByPlaceholderText('flow.taskDetail.descriptionPlaceholder'), {
      target: { value: 'New real description' },
    });

    expect(updateTaskMutate).toHaveBeenCalledWith({
      id: 't-1',
      data: { description: 'New real description' },
    });
  });
});

describe('TaskDetail activity empty state', () => {
  it('shows the empty title with a CTA back to Comments, which switches the tab', () => {
    taskDetailData.data = {
      task: apiTask(),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    fireEvent.click(screen.getByText('flow.taskDetail.activity'));

    expect(screen.getByText('flow.taskDetail.noActivity')).toBeInTheDocument();
    const goToComments = screen.getByRole('button', { name: 'flow.taskDetail.goToComments' });
    expect(goToComments).toBeEnabled();

    fireEvent.click(goToComments);
    expect(screen.queryByText('flow.taskDetail.noActivity')).not.toBeInTheDocument();
  });
});

describe('TaskDetail label rendering (#1054)', () => {
  // ApiLabel declares id as a number, but useLabels casts the response instead
  // of converting it, so at runtime these are int64 wire strings. The mocks use
  // strings deliberately: with numbers the old code would have worked and the
  // test would prove nothing.
  const labels = [
    { id: '994343', name: 'qa-temp', color: '#ff0000' },
    { id: '994344', name: 'blocked', color: '#00ff00' },
  ];

  it('renders a label whose id arrives as a wire string', () => {
    useLabelsSpy.mockReturnValue({ data: labels });
    taskDetailData.data = {
      task: apiTask({ labels: ['994343'] }) as ApiTask,
      subtasks: [],
      comments: [],
      activity: [],
    } as unknown as ApiTaskDetail;

    renderTaskDetail();

    expect(screen.getByText('qa-temp')).toBeInTheDocument();
    expect(screen.queryByText('blocked')).not.toBeInTheDocument();
  });

  it('renders labels arriving under the snake_case label_ids shape', () => {
    useLabelsSpy.mockReturnValue({ data: labels });
    taskDetailData.data = {
      task: { ...apiTask(), label_ids: ['994344'] },
      subtasks: [],
      comments: [],
      activity: [],
    } as unknown as ApiTaskDetail;

    renderTaskDetail();

    expect(screen.getByText('blocked')).toBeInTheDocument();
  });

  it('renders no label chip when the task has none', () => {
    useLabelsSpy.mockReturnValue({ data: labels });
    taskDetailData.data = {
      task: apiTask({ labels: [] }) as ApiTask,
      subtasks: [],
      comments: [],
      activity: [],
    } as unknown as ApiTaskDetail;

    renderTaskDetail();

    expect(screen.queryByText('qa-temp')).not.toBeInTheDocument();
    expect(screen.queryByText('blocked')).not.toBeInTheDocument();
  });
});

describe('TaskDetail effort control (#1053)', () => {
  it('labels the no-effort option instead of rendering a bare comma', () => {
    taskDetailData.data = {
      task: apiTask(),
      subtasks: [],
      comments: [],
      activity: [],
    } as unknown as ApiTaskDetail;

    renderTaskDetail();

    expect(screen.getByRole('button', { name: 'flow.quickCreate.noEffort' })).toBeInTheDocument();
    // The comma was the whole visible label of that button.
    expect(screen.queryByRole('button', { name: ',' })).not.toBeInTheDocument();
  });
});

describe('TaskDetail completion toast (#1040)', () => {
  const renderWithStatus = (status: ApiTask['status']): void => {
    taskDetailData.data = {
      task: apiTask({ status }),
      subtasks: [],
      comments: [],
      activity: [],
    } as unknown as ApiTaskDetail;
    renderTaskDetail();
  };

  // The selects in TaskDetail carry no accessible name: each label lives in a
  // sibling element that is not associated with the control. That is a real gap,
  // tracked by #1096 (replace the native selects with the shared Select), and
  // out of scope here. Identify the status select by an option only it has.
  const statusSelect = (): HTMLSelectElement => {
    const found = (screen.getAllByRole('combobox') as HTMLSelectElement[]).find((select) =>
      [...select.options].some((option) => option.value === 'in_review')
    );
    if (!found) throw new Error('status select not found');
    return found;
  };

  it('does not celebrate before the write has succeeded', async () => {
    const { toast } = await import('sonner');
    renderWithStatus('todo');

    fireEvent.change(statusSelect(), { target: { value: 'done' } });

    expect(updateTaskMutate).toHaveBeenCalledTimes(1);
    expect(toast.success).not.toHaveBeenCalled();
  });

  it('celebrates from the mutation success path', async () => {
    const { toast } = await import('sonner');
    renderWithStatus('todo');

    fireEvent.change(statusSelect(), { target: { value: 'done' } });
    const options = updateTaskMutate.mock.calls[0][1] as { onSuccess?: () => void };
    expect(options?.onSuccess).toBeTypeOf('function');

    options.onSuccess?.();

    expect(toast.success).toHaveBeenCalledTimes(1);
    expect(toast.success).toHaveBeenCalledWith('toasts.task.niceWork', expect.any(Object));
  });

  it('passes no success handler for a status change that is not a completion', async () => {
    renderWithStatus('todo');

    fireEvent.change(statusSelect(), { target: { value: 'in_progress' } });

    expect(updateTaskMutate.mock.calls[0][1]).toBeUndefined();
  });

  it('does not celebrate re-selecting done on an already done task', async () => {
    const { toast } = await import('sonner');
    renderWithStatus('done');

    // fireEvent.change dispatches even for an unchanged value, so the handler
    // does run; what matters is that it carries no success handler and so
    // cannot celebrate a completion that already happened.
    fireEvent.change(statusSelect(), { target: { value: 'done' } });

    expect(updateTaskMutate.mock.calls[0][1]).toBeUndefined();
    expect(toast.success).not.toHaveBeenCalled();
  });
});
