import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { chooseFlowOption, flowSelectTrigger } from '../../../../../test/flowSelect';
import { STATUS_LABEL_KEYS } from '../../shared/lib/constantsFlow';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { ApiTask, ApiTaskDetail } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { language: 'en', dir: () => 'ltr' },
  }),
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
  useUpdateComment: () => ({ mutate: vi.fn(), isPending: false }),
  useAddSubtask: () => ({ mutate: vi.fn(), isPending: false }),
  useToggleSubtask: () => ({ mutate: vi.fn(), isPending: false }),
  useTaskAttachments: () => ({ data: [] }),
  useAddAttachment: () => ({ mutate: vi.fn(), isPending: false }),
  useRemoveAttachment: () => ({ mutate: vi.fn(), isPending: false }),
}));

const cyclesData = vi.hoisted(() => ({ data: [] as Array<Record<string, unknown>> }));
vi.mock('../cycles/useCycles', () => ({ useCycles: () => cyclesData }));
vi.mock('../milestones/useMilestones', () => ({ useMilestones: () => ({ data: [] }) }));
vi.mock('../projects/useColumns', () => ({ useProjectColumns: () => ({ data: [] }) }));
vi.mock('../projects/useProjectMembers', () => ({ useProjectMembers: () => ({ data: [] }) }));

const useLabelsSpy = vi.hoisted(() => vi.fn(() => ({ data: [] })));
vi.mock('./useLabels', () => ({
  useLabels: useLabelsSpy,
  useSetTaskLabels: () => ({ mutate: vi.fn(), isPending: false }),
}));

const taskActivity = vi.hoisted(() => ({ data: [] as unknown[] }));
vi.mock('./useTaskActivity', () => ({
  useTaskActivity: () => taskActivity,
}));
vi.mock('./useTimeLogs', () => ({
  useTimeLogs: () => ({ data: { timeLogs: [] }, isPending: false, isError: false }),
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

const developerProjectMock = vi.hoisted(() => vi.fn(() => false));
vi.mock('../projects-v2/useDeveloperMode', () => ({
  useIsDeveloperProject: () => developerProjectMock(),
}));
vi.mock('./components/LinksSection', () => ({ default: () => null }));
vi.mock('./useTaskLinks', () => ({ useTaskLinks: () => ({ data: [] }) }));
vi.mock('../projects/useCurrentProject', () => ({
  useCurrentProject: useCurrentProjectSpy,
}));

// The real editor is a contenteditable the DOM shim cannot type into; its own
// behaviour is covered in DescriptionEditor.test.tsx.
vi.mock('./components/DescriptionEditor', () => ({
  default: ({
    value,
    onSave,
    placeholder,
  }: {
    value: string;
    onSave: (markdown: string) => void;
    placeholder: string;
  }) => (
    <textarea
      placeholder={placeholder}
      defaultValue={value}
      onBlur={(event) => {
        if (event.target.value !== value) onSave(event.target.value);
      }}
    />
  ),
}));
vi.mock('./components/TimePanel', () => ({
  default: () => null,
}));

const canEditMock = vi.hoisted(() => vi.fn(() => true));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanEditTask: () => canEditMock(),
  useCanDeleteResource: () => canEditMock(),
}));

vi.mock('../../shared/components/Breadcrumbs', () => ({
  default: ({ items }: { items: { label: string }[] }) => (
    <>{items.map((item) => item.label).join(' / ')}</>
  ),
}));
// The thread itself is stubbed; the print view still renders comment text
// through the real renderCommentBody.
vi.mock('../../shared/components/CommentList', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../shared/components/CommentList')>()),
  default: () => null,
}));
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
import { useUserStore } from '../../shared/stores/userStore';

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
  developerProjectMock.mockReturnValue(false);
  canEditMock.mockReturnValue(true);
  taskActivity.data = [];
  // The activity tab a person last chose is remembered in localStorage.
  window.localStorage.clear();
  useUserStore.setState({ user: null });
  useFlowStore.setState({ members: [] });
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

    fireEvent.click(screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.history' }));

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

    expect(within(screen.getByRole('dialog')).getByText('SENTINEL PROBE')).toBeInTheDocument();
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

describe('TaskDetail Development section (T-365)', () => {
  const detail = () =>
    ({
      task: apiTask({}) as ApiTask,
      subtasks: [],
      comments: [],
      activity: [],
    }) as unknown as ApiTaskDetail;

  it('is hidden on a project without developer features', () => {
    taskDetailData.data = detail();
    renderTaskDetail();
    expect(screen.queryByText('flow.taskDetail.development')).not.toBeInTheDocument();
  });

  it('appears with the dev fields on a developer project', () => {
    developerProjectMock.mockReturnValue(true);
    taskDetailData.data = detail();
    renderTaskDetail();
    expect(
      screen.getByRole('heading', { name: 'flow.taskDetail.development' })
    ).toBeInTheDocument();
    expect(screen.getByText('flow.taskDetail.devPoints')).toBeInTheDocument();
  });
});

describe('TaskDetail cycle field (T-336)', () => {
  beforeEach(() => {
    cyclesData.data = [
      { id: 'c-1', name: 'Sprint 1', status: 'active' },
      { id: 'c-2', name: 'Sprint 2', status: 'upcoming' },
      { id: 'c-0', name: 'Sprint 0', status: 'completed' },
    ];
  });
  afterEach(() => {
    cyclesData.data = [];
  });

  const detailWith = (cycleId?: string) =>
    ({
      task: apiTask({ cycleId }),
      subtasks: [],
      comments: [],
      activity: [],
    }) as unknown as ApiTaskDetail;

  it('moves the task into an open cycle', async () => {
    taskDetailData.data = detailWith();
    renderTaskDetail();
    expect(flowSelectTrigger('flow.taskDetail.cycle')).toHaveAccessibleName(
      'flow.taskDetail.cycle: flow.taskDetail.noCycle'
    );
    await chooseFlowOption('flow.taskDetail.cycle', 'Sprint 2');
    expect(updateTaskMutate).toHaveBeenCalledWith({ id: 't-1', data: { cycleId: 'c-2' } });
  });

  it('returns the task to the backlog and never offers a completed cycle as a target', async () => {
    taskDetailData.data = detailWith('c-1');
    renderTaskDetail();
    fireEvent.pointerDown(flowSelectTrigger('flow.taskDetail.cycle'), {
      button: 0,
      ctrlKey: false,
    });
    const backlog = await screen.findByRole('menuitem', { name: 'flow.taskDetail.noCycle' });
    expect(screen.queryByRole('menuitem', { name: 'Sprint 0' })).not.toBeInTheDocument();
    fireEvent.click(backlog);
    expect(updateTaskMutate).toHaveBeenLastCalledWith({ id: 't-1', data: { cycleId: '' } });
  });
});

describe('TaskDetail Assign to me (T-603)', () => {
  const renderAssignedTo = (assigneeId?: string): void => {
    useUserStore.setState({ user: { id: 7, fullName: 'Abed Amouneh' } });
    taskDetailData.data = {
      task: apiTask({ assigneeId }),
      subtasks: [],
      comments: [],
      activity: [],
    };
    renderTaskDetail();
  };

  it('assigns the task to the signed-in person in one click', () => {
    renderAssignedTo('12');
    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.assignToMe' }));
    expect(updateTaskMutate).toHaveBeenCalledWith({
      id: 't-1',
      data: { assigneeId: '7', assignee: 'Abed Amouneh', assigneeInitials: 'AA' },
    });
  });

  it('is hidden when the task is already assigned to them', () => {
    renderAssignedTo('7');
    expect(
      screen.queryByRole('button', { name: 'flow.taskDetail.assignToMe' })
    ).not.toBeInTheDocument();
  });

  it('is hidden from someone who cannot edit the task', () => {
    canEditMock.mockReturnValue(false);
    renderAssignedTo(undefined);
    expect(
      screen.queryByRole('button', { name: 'flow.taskDetail.assignToMe' })
    ).not.toBeInTheDocument();
  });
});

describe('TaskDetail reporter (T-603)', () => {
  it('names the reporter from the people the panel already knows', () => {
    useFlowStore.setState({
      members: [{ id: '31', name: 'Mohamad Homsi', initials: 'MH' }] as never,
    });
    taskDetailData.data = {
      task: apiTask({ reporterId: '31' }),
      subtasks: [],
      comments: [],
      activity: [],
    };
    renderTaskDetail();
    expect(screen.getByText('flow.taskDetail.reporter')).toBeInTheDocument();
    expect(screen.getByText('Mohamad Homsi')).toBeInTheDocument();
  });

  it('leaves the field out when the task carries no reporter', () => {
    taskDetailData.data = {
      task: apiTask(),
      subtasks: [],
      comments: [],
      activity: [],
    };
    renderTaskDetail();
    expect(screen.queryByText('flow.taskDetail.reporter')).not.toBeInTheDocument();
  });
});

describe('TaskDetail print (T-603)', () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('lays the task out for print and opens the browser print dialog', () => {
    vi.useFakeTimers();
    // The DOM shim has no print(); the real browser opens its dialog here.
    const print = vi.fn();
    vi.stubGlobal('print', print);
    taskDetailData.data = {
      task: apiTask({ identifier: 'FLOW-603', title: 'Printable task' }),
      subtasks: [{ id: 's-1', title: 'Write the print view', status: 'done' }],
      comments: [
        {
          id: 'c-1',
          taskId: 'task-1',
          userId: '7',
          actorName: 'Abed Amouneh',
          content: 'Looks **ready** to ship',
          isEdited: false,
          createdAt: '2026-11-20T10:00:00Z',
          updatedAt: '2026-11-20T10:00:00Z',
        },
      ],
      activity: [],
    };
    renderTaskDetail();

    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.print.action' }));

    const page = document.body.querySelector('.flow-task-print');
    expect(page).not.toBeNull();
    expect(document.documentElement).toHaveClass('flow-printing');
    expect(within(page as HTMLElement).getByText('Printable task')).toBeInTheDocument();
    expect(within(page as HTMLElement).getByText('Write the print view')).toBeInTheDocument();
    // The comment prints the way the thread shows it, not as raw Markdown.
    const bold = within(page as HTMLElement).getByText('ready');
    expect(bold.tagName).toBe('STRONG');
    expect(page?.textContent).not.toContain('**ready**');

    vi.runAllTimers();
    expect(print).toHaveBeenCalledTimes(1);

    fireEvent(window, new Event('afterprint'));
    expect(document.body.querySelector('.flow-task-print')).toBeNull();
    expect(document.documentElement).not.toHaveClass('flow-printing');
  });
});
