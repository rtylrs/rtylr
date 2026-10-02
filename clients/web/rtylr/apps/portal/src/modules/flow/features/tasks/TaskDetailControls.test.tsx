import React from 'react';
import { createInstance } from 'i18next';
import { I18nextProvider } from 'react-i18next';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import flow from '../../locales/en/flow.json';
import flowCommon from '../../locales/en/flowCommon.json';
import type { ApiTask, ApiTaskDetail } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';

const state = vi.hoisted(() => ({
  detail: undefined as ApiTaskDetail | undefined,
  tasks: [] as ApiTask[],
  canEdit: true,
}));
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
  useDeleteTask: () => ({ mutate: vi.fn() }),
  useAddComment: () => ({ mutate: vi.fn() }),
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
vi.mock('../projects/useStatuses', () => ({
  useProjectStatuses: () => ({ data: [] }),
  useSetTaskStatus: () => ({ mutate: vi.fn() }),
}));
vi.mock('../projects/useCurrentProject', () => ({
  useCurrentProject: () => ({ project: undefined }),
}));
vi.mock('../projects/useColumns', () => ({
  useProjectColumns: () => ({ data: [{ id: '6', name: 'Review', sortOrder: 0 }] }),
}));
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
});

describe('Task detail persistence controls', () => {
  it('shows the saved assignee to a viewer who cannot load the full people list', () => {
    state.canEdit = false;
    Object.assign(state.detail!.task, {
      assigneeId: '1001',
      assignee: 'Taylor Reed',
      assigneeInitials: 'TR',
    });
    renderDetail();
    const assignee = screen.getByRole('combobox', { name: flow.taskDetail.assignee });
    expect(assignee).toBeDisabled();
    expect(assignee).toHaveValue('1001');
    expect(screen.getByRole('option', { name: 'Taylor Reed' })).toBeInTheDocument();
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

  it('links a remote task and removes existing links through the dependency endpoints', () => {
    const rendered = renderDetail();
    fireEvent.change(screen.getByRole('combobox', { name: flow.taskDetail.addDependency }), {
      target: { value: '2' },
    });
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

  it('loads the project columns and can remove the selected column', () => {
    state.detail!.task.sectionId = '6';
    renderDetail();
    const column = screen.getByRole('combobox', { name: flow.taskDetail.section });
    expect(column).toHaveValue('6');
    fireEvent.change(column, { target: { value: '' } });
    expect(update).toHaveBeenCalledWith({ id: '1', data: { sectionId: '' } });
  });
});
