import React from 'react';
import { createInstance } from 'i18next';
import { I18nextProvider } from 'react-i18next';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import flow from '../../locales/en/flow.json';
import { chooseFlowOption, flowSelectTrigger } from '../../../../../test/flowSelect';
import flowCommon from '../../locales/en/flowCommon.json';
import type { ApiStatus, ApiTask, ApiTaskDetail } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';

const state = vi.hoisted(() => ({
  detail: undefined as ApiTaskDetail | undefined,
  tasks: [] as ApiTask[],
  statuses: [] as ApiStatus[],
  canEdit: true,
}));
const setStatus = vi.hoisted(() => vi.fn());
const createStatus = vi.hoisted(() => vi.fn());
const update = vi.hoisted(() => vi.fn());
const addDependency = vi.hoisted(() => vi.fn());
const removeDependency = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: '1' }),
  useNavigate: () => vi.fn(),
  Link: ({ children }: { children: React.ReactNode }) => <a>{children}</a>,
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('./useTask', () => ({ useTask: () => ({ data: state.detail, isLoading: false }) }));
vi.mock('./useTasks', () => ({
  useTasks: () => ({ data: state.tasks, isLoading: false }),
  useUpdateTask: () => ({ mutate: update }),
  useAddTaskDependency: () => ({ mutate: addDependency }),
  useRemoveTaskDependency: () => ({ mutate: removeDependency }),
  useSetTaskParent: () => ({ mutate: vi.fn(), isPending: false }),
  useEpicRollups: () => ({ data: [] }),
  useDeleteTask: () => ({ mutate: vi.fn() }),
  useAddComment: () => ({ mutate: vi.fn() }),
  useUpdateComment: () => ({ mutate: vi.fn() }),
  useAddSubtask: () => ({ mutate: vi.fn() }),
  useToggleSubtask: () => ({ mutate: vi.fn() }),
  useTaskAttachments: () => ({ data: { attachments: [] } }),
  useAddAttachment: () => ({ mutate: vi.fn() }),
  useRemoveAttachment: () => ({ mutate: vi.fn() }),
}));
vi.mock('./useLabels', () => ({
  useLabels: () => ({ data: [] }),
  useSetTaskLabels: () => ({ mutate: vi.fn() }),
}));
vi.mock('./useTaskActivity', () => ({ useTaskActivity: () => ({ data: [] }) }));
vi.mock('./useTimeLogs', () => ({
  useTimeLogs: () => ({ data: { timeLogs: [] }, isPending: false, isError: false }),
}));
vi.mock('../projects/useStatuses', () => ({
  useProjectStatuses: () => ({ data: state.statuses }),
  useSetTaskStatus: () => ({ mutate: setStatus }),
  useCreateStatus: () => ({ mutate: createStatus, isPending: false }),
}));
vi.mock('../projects-v2/useDeveloperMode', () => ({ useIsDeveloperProject: () => false }));
vi.mock('./components/LinksSection', () => ({ default: () => null }));
vi.mock('../projects/useCurrentProject', () => ({
  useCurrentProject: () => ({ project: undefined }),
}));
vi.mock('../../shared/components/AttachmentUpload', () => ({ default: () => null }));
vi.mock('../projects/useProjectMembers', () => ({ useProjectMembers: () => ({ data: [] }) }));
vi.mock('../cycles/useCycles', () => ({ useCycles: () => ({ data: [] }) }));
vi.mock('../milestones/useMilestones', () => ({ useMilestones: () => ({ data: [] }) }));
vi.mock('../../shared/hooks/queries/useApprovedLeaveInRange', () => ({
  useApprovedLeaveInRange: () => ({ data: [] }),
}));
vi.mock('../../shared/lib/useTaskRealtime', () => ({ useTaskRealtime: () => {} }));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanEditTask: () => state.canEdit,
  useCanDeleteResource: () => state.canEdit,
}));
vi.mock('./components/TimePanel', () => ({ default: () => null }));

import TaskDetail from './TaskDetail';
const i18n = createInstance();
void i18n.init({
  lng: 'en',
  resources: { en: { translation: { flow, flowCommon } } },
  interpolation: { escapeValue: false },
});
const task = (id: string, title: string): ApiTask => ({
  id,
  projectId: '2',
  title,
  status: 'todo',
  priority: 'high',
  assignee: '',
  assigneeInitials: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Project',
});
const renderDetail = () =>
  render(
    <I18nextProvider i18n={i18n}>
      <TaskDetail />
    </I18nextProvider>
  );

beforeEach(() => {
  vi.clearAllMocks();
  state.canEdit = true;
  useFlowStore.setState({ members: [] });
  state.detail = {
    task: {
      ...task('1', 'Initial title'),
      waitingOn: 'Approval',
      decisionNote: 'Original plan',
      effortMinutes: 60,
      estimatedHours: 1,
    },
    subtasks: [],
    comments: [],
    activity: [],
  };
  state.tasks = [task('1', 'Initial title'), task('2', 'Remote blocker')];
  state.statuses = [];
});

const status = (id: string, name: string, bucket: ApiStatus['bucket'], sortOrder: number) => ({
  id,
  projectId: '2',
  name,
  bucket,
  legacyStatus: bucket === 'done' ? 'done' : 'in_review',
  color: '#8b5cf6',
  sortOrder,
  isDefault: false,
  createdAt: '',
});

describe('Task detail persistence controls', () => {
  it('shows the saved assignee to a viewer who cannot load the full people list', () => {
    state.canEdit = false;
    Object.assign(state.detail!.task, {
      assigneeId: '1001',
      assignee: '',
      assigneeName: 'Taylor Reed',
      assigneeInitials: 'TR',
    });
    renderDetail();
    const assignee = flowSelectTrigger(flow.taskDetail.assignee);
    expect(assignee).toBeDisabled();
    expect(assignee).toHaveAccessibleName(`${flow.taskDetail.assignee}: Taylor Reed`);
    expect(update).not.toHaveBeenCalled();
  });

  it('keeps typing local until blur and can clear blocker text', () => {
    renderDetail();
    const title = screen.getByRole('textbox', { name: 'Task title' });
    fireEvent.focus(title);
    fireEvent.change(title, { target: { value: 'A complete replacement title' } });
    expect(title).toHaveValue('A complete replacement title');
    expect(update).not.toHaveBeenCalled();
    fireEvent.blur(title);
    expect(update).toHaveBeenCalledWith({
      id: '1',
      data: { title: 'A complete replacement title' },
    });
    const waiting = screen.getByPlaceholderText(flow.taskDetail.waitingPlaceholder);
    fireEvent.change(waiting, { target: { value: '' } });
    fireEvent.blur(waiting);
    expect(update).toHaveBeenCalledWith({ id: '1', data: { waitingOn: '' } });
  });

  it('sends explicit zero when clearing effort', () => {
    renderDetail();
    fireEvent.click(screen.getByRole('button', { name: flow.quickCreate.noEffort }));
    expect(update).toHaveBeenCalledWith({ id: '1', data: { effortMinutes: 0, estimate: 0 } });
  });

  it('links a remote task and removes existing links through the dependency endpoints', async () => {
    const rendered = renderDetail();
    await chooseFlowOption(flow.taskDetail.addDependency, /Remote blocker/);
    fireEvent.click(screen.getByRole('button', { name: flow.taskDetail.add, exact: true }));
    expect(addDependency).toHaveBeenCalledWith(
      { taskId: '1', dependsOnTaskId: '2' },
      expect.any(Object)
    );
    state.detail!.task.dependencies = ['2'];
    rendered.rerender(
      <I18nextProvider i18n={i18n}>
        <TaskDetail />
      </I18nextProvider>
    );
    expect(screen.getByText(/Remote blocker/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: flow.taskDetail.remove, exact: true }));
    expect(removeDependency).toHaveBeenCalledWith({ taskId: '1', dependsOnTaskId: '2' });
  });

  it('picks the status from the project statuses and has no separate section field', async () => {
    state.statuses = [status('6', 'Review', 'active', 30), status('7', 'Shipped', 'done', 900)];
    state.detail!.task.statusId = '6';
    state.detail!.task.sectionId = '6';
    renderDetail();
    expect(flowSelectTrigger(flow.taskDetail.status)).toHaveAccessibleName(
      `${flow.taskDetail.status}: Review`
    );
    expect(screen.queryByText('Section')).not.toBeInTheDocument();
    await chooseFlowOption(flow.taskDetail.status, 'Shipped');
    expect(setStatus).toHaveBeenCalledWith({ taskId: '1', statusId: '7' }, expect.any(Object));
  });

  it('creates a status from the status menu and applies it to the task', async () => {
    state.statuses = [status('6', 'Review', 'active', 30), status('7', 'Shipped', 'done', 900)];
    state.detail!.task.statusId = '6';
    createStatus.mockImplementation((_vars, options?: { onSuccess?: (s: ApiStatus) => void }) =>
      options?.onSuccess?.(status('9', 'Blocked', 'active', 910))
    );
    renderDetail();
    await chooseFlowOption(flow.taskDetail.status, flow.status.newStatusAction);
    fireEvent.change(await screen.findByLabelText(flow.status.statusName), {
      target: { value: 'Blocked' },
    });
    fireEvent.click(screen.getByRole('button', { name: flow.status.addStatus }));
    expect(createStatus).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'Blocked', bucket: 'active', sortOrder: 910 }),
      expect.any(Object)
    );
    expect(setStatus).toHaveBeenCalledWith({ taskId: '1', statusId: '9' });
  });

  it('offers no new-status action to someone who cannot manage the project', async () => {
    state.canEdit = false;
    state.statuses = [status('6', 'Review', 'active', 30), status('7', 'Shipped', 'done', 900)];
    renderDetail();
    expect(flowSelectTrigger(flow.taskDetail.status)).toBeDisabled();
  });
});
