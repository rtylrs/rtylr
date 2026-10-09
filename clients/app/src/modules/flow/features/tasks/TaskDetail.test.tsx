import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { chooseFlowOption, flowSelectTrigger } from '../../../../../test/flowSelect';
import { STATUS_LABEL_KEYS } from '../../shared/lib/constantsFlow';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
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
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));
vi.mock('./useTask', () => ({
  useTask: () => taskDetailData,
}));

const updateTaskMutate = vi.hoisted(() => vi.fn());
const deleteTaskMutate = vi.hoisted(() => vi.fn());
vi.mock('./useTasks', () => ({
  useTasks: () => ({ data: [], isLoading: false, isError: false }),
  useAddTaskDependency: () => ({ mutate: vi.fn(), isPending: false }),
  useRemoveTaskDependency: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateTask: () => ({ mutate: updateTaskMutate, isPending: false }),
  useDeleteTask: () => ({ mutate: deleteTaskMutate, isPending: false }),
  useAddComment: () => ({ mutate: vi.fn(), isPending: false }),
  useAddSubtask: () => ({ mutate: vi.fn(), isPending: false }),
  useToggleSubtask: () => ({ mutate: vi.fn(), isPending: false }),
  useTaskAttachments: () => ({ data: [] }),
  useAddAttachment: () => ({ mutate: vi.fn(), isPending: false }),
  useRemoveAttachment: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../cycles/useCycles', () => ({ useCycles: () => ({ data: [] }) }));
vi.mock('../milestones/useMilestones', () => ({ useMilestones: () => ({ data: [] }) }));
vi.mock('../projects/useColumns', () => ({ useProjectColumns: () => ({ data: [] }) }));

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
// Kept about whether deletion requires a confirmation, not about the dialog's
// own markup, which ConfirmDialog's own usage sites already exercise.
vi.mock('../../shared/components/ConfirmDialog', () => ({
  default: ({
    open,
    onConfirm,
    description,
  }: {
    open: boolean;
    onConfirm: () => void;
    description: React.ReactNode;
  }) =>
    open ? (
      <div role="dialog">
        <div>{description}</div>
        <button type="button" onClick={onConfirm}>
          confirm-delete
        </button>
      </div>
    ) : null,
}));

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
  // The query-failure flags belong here, not in one describe's own hook. They
  // were reset inside the failure describe only, so `--sequence.shuffle` could
  // run a failure test before another and leave this stuck (#1194).
  taskDetailData.isError = false;
  taskDetailData.isPaused = false;
  taskDetailData.failureCount = 0;
  useFlowStore.setState({
    projects: [],
  });
  taskDetailData.data = undefined;
  taskDetailData.isLoading = false;
});

describe('TaskDetail business context', () => {
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
      task: apiTask({ projectId: 'p-9' }),
      subtasks: [],
      comments: [],
      activity: [],
    };

    renderTaskDetail();

    expect(useCurrentProjectSpy).toHaveBeenCalledWith('p-9');
  });

  it("falls back to the task's own projectName when the project isn't in the store", () => {
    taskDetailData.data = {
      task: apiTask({
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
    expect(updateTaskMutate).not.toHaveBeenCalled();
    fireEvent.blur(screen.getByPlaceholderText('flow.taskDetail.descriptionPlaceholder'));

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

  // The status control is a FlowSelect: open it, then pick an option by its
  // visible label (the test `t` returns the key, so that is the label too).
  const chooseStatus = (status: keyof typeof STATUS_LABEL_KEYS): Promise<void> =>
    chooseFlowOption('flow.taskDetail.status', STATUS_LABEL_KEYS[status]);

  it('does not celebrate before the write has succeeded', async () => {
    const { toast } = await import('sonner');
    renderWithStatus('todo');

    await chooseStatus('done');

    expect(updateTaskMutate).toHaveBeenCalledTimes(1);
    expect(toast.success).not.toHaveBeenCalled();
  });

  it('celebrates from the mutation success path', async () => {
    const { toast } = await import('sonner');
    renderWithStatus('todo');

    await chooseStatus('done');
    const options = updateTaskMutate.mock.calls[0][1] as { onSuccess?: () => void };
    expect(options?.onSuccess).toBeTypeOf('function');

    options.onSuccess?.();

    expect(toast.success).toHaveBeenCalledTimes(1);
    expect(toast.success).toHaveBeenCalledWith('flowToasts.task.niceWork', expect.any(Object));
  });

  it('passes no success handler for a status change that is not a completion', async () => {
    renderWithStatus('todo');

    await chooseStatus('in_progress');

    expect(updateTaskMutate.mock.calls[0][1]).toBeUndefined();
  });

  it('does not celebrate re-selecting done on an already done task', async () => {
    const { toast } = await import('sonner');
    renderWithStatus('done');

    // Picking the status the task already has is a no-op: nothing is written
    // and so nothing can celebrate a completion that already happened.
    await chooseStatus('done');

    expect(updateTaskMutate).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
  });
});

describe('TaskDetail deletion confirmation (#1056)', () => {
  const renderWithTask = (): void => {
    taskDetailData.data = {
      task: apiTask({ title: 'SENTINEL PROBE' }),
      subtasks: [],
      comments: [],
      activity: [],
    } as unknown as ApiTaskDetail;
    renderTaskDetail();
  };

  const deleteButton = (): HTMLElement =>
    screen.getByRole('button', { name: 'flow.taskDetail.delete' });

  it('does not delete on the first click', () => {
    renderWithTask();

    fireEvent.click(deleteButton());

    expect(deleteTaskMutate).not.toHaveBeenCalled();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('names the task being deleted', () => {
    renderWithTask();

    fireEvent.click(deleteButton());

    expect(screen.getByText('SENTINEL PROBE')).toBeInTheDocument();
  });

  it('deletes once confirmed', () => {
    renderWithTask();

    fireEvent.click(deleteButton());
    fireEvent.click(screen.getByRole('button', { name: 'confirm-delete' }));

    expect(deleteTaskMutate).toHaveBeenCalledWith('t-1', expect.any(Object));
  });
});

describe('TaskDetail when the request fails', () => {
  afterEach(() => {
    taskDetailData.isError = false;
    taskDetailData.isPaused = false;
    taskDetailData.failureCount = 0;
  });

  it('offers a retry instead of reporting the task deleted', () => {
    // A failed request leaves the task undefined exactly as a deleted one
    // does, so a 500 used to say the task was gone and offer no way back.
    taskDetailData.data = undefined;
    taskDetailData.isError = true;
    taskDetailData.failureCount = 1;
    renderTaskDetail();

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.taskDetail.notFound')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(taskDetailData.refetch).toHaveBeenCalled();
  });

  it('reports a paused query that already failed as a failure, not as missing', () => {
    // networkMode 'online' parks a failed query at pending/paused, which never
    // reaches isError and used to fall through to the not-found branch.
    taskDetailData.data = undefined;
    taskDetailData.isPaused = true;
    taskDetailData.failureCount = 1;
    renderTaskDetail();
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
  });

  it('still reports a genuinely missing task as not found', () => {
    // Both directions, or the branch above could swallow a real deletion.
    taskDetailData.data = undefined;
    renderTaskDetail();
    expect(screen.getByText('flow.taskDetail.notFound')).toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });
});

describe('every task detail control has an accessible name (#1079)', () => {
  it('leaves no combobox, dropdown trigger, textbox or spinbutton unnamed', () => {
    // WCAG 4.1.2. Ten controls had no name: the title input, six selects in
    // the properties rail, two custom-field inputs and the dependency picker.
    // Each sat beside visible label text that was never associated with it,
    // so a screen reader announced "combo box" with nothing else.
    //
    // Queried by role rather than by reading the source, so this measures what
    // the accessibility tree actually exposes.
    taskDetailData.data = {
      task: apiTask({}),
      subtasks: [],
      comments: [],
      activity: [],
    };
    renderTaskDetail();

    const controls = [
      ...screen.queryAllByRole('combobox'),
      ...screen
        .queryAllByRole('button')
        .filter((button) => button.getAttribute('aria-haspopup') === 'menu'),
      ...screen.queryAllByRole('textbox'),
      ...screen.queryAllByRole('spinbutton'),
    ];

    expect(controls.length, 'no controls rendered, so nothing was measured').toBeGreaterThan(4);

    const unnamed = controls.filter((el) => {
      const label =
        el.getAttribute('aria-label') ??
        el.getAttribute('title') ??
        el.getAttribute('placeholder') ??
        '';
      return label.trim() === '';
    });

    expect(
      unnamed.map((el) => `${el.tagName.toLowerCase()}[${el.getAttribute('class') ?? ''}]`),
      'these controls expose no accessible name'
    ).toEqual([]);
  });

  it('names the title input, which carried no label at all', () => {
    taskDetailData.data = {
      task: apiTask({}),
      subtasks: [],
      comments: [],
      activity: [],
    };
    renderTaskDetail();
    expect(screen.getByRole('textbox', { name: 'flow.taskDetail.taskTitle' })).toBeInTheDocument();
  });

  it('names the status dropdown after the label shown beside it', () => {
    taskDetailData.data = {
      task: apiTask({}),
      subtasks: [],
      comments: [],
      activity: [],
    };
    renderTaskDetail();
    expect(flowSelectTrigger('flow.taskDetail.status')).toBeInTheDocument();
  });
});
