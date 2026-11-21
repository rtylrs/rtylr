import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { initTestI18n } from '../testing/i18n';
import { useUserStore } from '../stores/userStore';
import { useFlowStore } from '../stores/flowStore';
import QuickCreate from './QuickCreate';
import { chooseFlowOption, flowSelectTrigger } from '../../../../../test/flowSelect';
import { toast } from 'sonner';
import type { ApiField, ApiStatus } from '../services/flowApi';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  create: vi.fn(),
  currentProjectId: undefined as string | undefined,
  labels: [] as Array<{ id: number; name: string }>,
  statuses: [] as ApiStatus[],
  milestones: [] as Array<{ id: string; name: string }>,
  fields: [] as ApiField[],
}));
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
  useParams: () => ({ projectId: mocks.currentProjectId }),
  useRouterState: ({
    select,
  }: {
    select: (state: { location: { pathname: string } }) => unknown;
  }) => select({ location: { pathname: '/flow/my-issues' } }),
}));
vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));
vi.mock('../../features/projects/useProjects', () => ({
  useProjects: () => ({ data: undefined }),
}));
vi.mock('../../features/tasks/useTasks', () => ({
  useCreateTask: () => ({ mutateAsync: mocks.create, isPending: false }),
}));
vi.mock('../../features/tasks/useLabels', () => ({ useLabels: () => ({ data: mocks.labels }) }));
vi.mock('../../features/milestones/useMilestones', () => ({
  useMilestones: (projectId: string) => ({ data: projectId ? mocks.milestones : undefined }),
}));
vi.mock('../../features/projects/useFields', () => ({
  useProjectFields: (projectId: string) => ({ data: projectId ? mocks.fields : undefined }),
}));
vi.mock('../../features/projects/useProjectMembers', () => ({
  useProjectMembers: () => ({ data: [] }),
}));
vi.mock('../../features/projects/useStatuses', () => ({
  useProjectStatuses: (projectId: string) => ({ data: projectId ? mocks.statuses : undefined }),
}));

const status = (id: string, name: string, sortOrder: number, isDefault = false): ApiStatus => ({
  id,
  projectId: '12',
  name,
  bucket: 'active',
  legacyStatus: 'in_progress',
  color: '#8b5cf6',
  sortOrder,
  isDefault,
  createdAt: '',
});

beforeEach(async () => {
  vi.clearAllMocks();
  await initTestI18n();
  mocks.currentProjectId = undefined;
  mocks.labels = [];
  mocks.milestones = [];
  mocks.fields = [];
  mocks.statuses = [status('3', 'To Do', 10, true), status('4', 'Doing', 20)];
  mocks.create.mockResolvedValue({
    task: { id: '101', identifier: 'FLOW-101', title: 'Ship the thing' },
  });
  useUserStore.setState({
    user: {
      id: 7,
      fullName: 'Ada Lovelace',
      email: 'ada@example.test',
      effectiveAccess: createEffectiveAccess('flow', { isOwner: true }),
    },
  });
  useFlowStore.setState({ quickCreateOpen: true, quickCreateColumnId: undefined, projects: [] });
});
function fillTitle() {
  fireEvent.change(screen.getByPlaceholderText('Task name'), {
    target: { value: 'Ship the thing' },
  });
}
function create() {
  fireEvent.click(screen.getByRole('button', { name: 'Create task', exact: true }));
}

describe('Quick task creation in Portal', () => {
  it('restores focus to the command button when the originating action disappears', async () => {
    useFlowStore.setState({ quickCreateOpen: false });
    const user = userEvent.setup();
    function Launcher() {
      const open = useFlowStore((state) => state.quickCreateOpen);
      return (
        <>
          <button data-portal-command-trigger>Search Portal</button>
          {!open && (
            <button onClick={() => useFlowStore.getState().setQuickCreateOpen(true)}>
              Open creation
            </button>
          )}
          <QuickCreate />
        </>
      );
    }
    render(<Launcher />);
    await user.click(screen.getByRole('button', { name: 'Open creation' }));
    await waitFor(() => expect(screen.getByPlaceholderText('Task name')).toHaveFocus());
    await user.keyboard('{Escape}');
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'Search Portal' })).toHaveFocus()
    );
    expect(screen.queryByRole('dialog')).toBeNull();
  });
  it('creates personal tasks even when no projects or labels exist', async () => {
    render(<QuickCreate />);
    fillTitle();
    create();
    await waitFor(() => expect(useFlowStore.getState().quickCreateOpen).toBe(false));
    expect(mocks.create).toHaveBeenCalledWith(
      expect.objectContaining({
        title: 'Ship the thing',
        projectId: undefined,
        labels: undefined,
        assigneeId: '7',
      })
    );
    expect(mocks.navigate).toHaveBeenCalledWith(
      expect.objectContaining({ to: '.', search: expect.any(Function) })
    );
  });
  it('keeps labels optional when the business has labels', async () => {
    mocks.labels = [
      { id: 1, name: 'Bug' },
      { id: 2, name: 'Feature' },
    ];
    render(<QuickCreate />);
    expect(screen.getByRole('combobox', { name: 'Labels' })).toHaveValue('');
    fillTitle();
    create();
    await waitFor(() =>
      expect(mocks.create).toHaveBeenCalledWith(expect.objectContaining({ labels: undefined }))
    );
  });
  it('sends a selected label and the current project', async () => {
    mocks.currentProjectId = '12';
    mocks.labels = [{ id: 8, name: 'Feature' }];
    render(<QuickCreate />);
    fillTitle();
    fireEvent.change(screen.getByRole('combobox', { name: 'Labels' }), { target: { value: '8' } });
    create();
    await waitFor(() =>
      expect(mocks.create).toHaveBeenCalledWith(
        expect.objectContaining({ projectId: '12', labels: ['8'] })
      )
    );
  });
  it('keeps an in-progress draft when the project list refreshes', () => {
    render(<QuickCreate />);
    fillTitle();
    act(() => useFlowStore.setState({ projects: [] }));
    expect(screen.getByPlaceholderText('Task name')).toHaveValue('Ship the thing');
  });
  it('requires a title and write access', () => {
    useUserStore.setState({
      user: {
        id: 7,
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:view'] }),
      },
    });
    render(<QuickCreate />);
    expect(screen.getByRole('button', { name: 'Create task', exact: true })).toBeDisabled();
    fillTitle();
    create();
    expect(mocks.create).not.toHaveBeenCalled();
  });
  it('keeps the dialog open when creation fails', async () => {
    mocks.create.mockRejectedValue(new Error('Unavailable'));
    render(<QuickCreate />);
    fillTitle();
    create();
    await waitFor(() => expect(mocks.create).toHaveBeenCalled());
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Task name')).toHaveValue('Ship the thing');
  });
  it('does not open a read-only detail drawer for a write-only member', async () => {
    useUserStore.setState({
      user: {
        id: 7,
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:create'] }),
      },
    });
    render(<QuickCreate />);
    fillTitle();
    create();
    await waitFor(() => expect(useFlowStore.getState().quickCreateOpen).toBe(false));
    expect(mocks.navigate).not.toHaveBeenCalled();
  });
  it('starts on the project default status and sends the chosen one as the column', async () => {
    mocks.currentProjectId = '12';
    render(<QuickCreate />);
    expect(flowSelectTrigger('Status')).toHaveAccessibleName('Status: To Do');
    await chooseFlowOption('Status', 'Doing');
    fillTitle();
    create();
    await waitFor(() =>
      expect(mocks.create).toHaveBeenCalledWith(
        expect.objectContaining({ projectId: '12', sectionId: '4' })
      )
    );
  });
  it('files the new task under the milestone picked for its project', async () => {
    mocks.currentProjectId = '12';
    mocks.milestones = [{ id: '30', name: 'Phase one live' }];
    render(<QuickCreate />);
    await chooseFlowOption('Milestone', 'Phase one live');
    fillTitle();
    create();
    await waitFor(() =>
      expect(mocks.create).toHaveBeenCalledWith(
        expect.objectContaining({ projectId: '12', milestoneId: '30' })
      )
    );
  });
  it('offers no milestone field when the project has none', () => {
    mocks.currentProjectId = '12';
    render(<QuickCreate />);
    expect(screen.queryByRole('button', { name: /^Milestone(:|$)/ })).toBeNull();
  });
  it('lands a task created from a board column in that column', async () => {
    mocks.currentProjectId = '12';
    useFlowStore.setState({ quickCreateColumnId: '4' });
    render(<QuickCreate />);
    expect(flowSelectTrigger('Status')).toHaveAccessibleName('Status: Doing');
    fillTitle();
    create();
    await waitFor(() =>
      expect(mocks.create).toHaveBeenCalledWith(expect.objectContaining({ sectionId: '4' }))
    );
  });
  it('shows no status field for a personal task', () => {
    render(<QuickCreate />);
    expect(screen.queryByRole('button', { name: /^Status/ })).toBeNull();
  });

  it('blocks create while a required custom field is empty and names it', async () => {
    mocks.currentProjectId = '12';
    mocks.fields = [field({ id: '5', name: 'Severity', type: 'select', required: true })];
    render(<QuickCreate />);
    fillTitle();
    create();
    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith('Fill in the required fields first: Severity')
    );
    expect(screen.getByRole('alert')).toHaveTextContent('This field is required.');
    expect(mocks.create).not.toHaveBeenCalled();

    await chooseFlowOption('Severity', 'High');
    create();
    await waitFor(() =>
      expect(mocks.create).toHaveBeenCalledWith(
        expect.objectContaining({ fieldValues: [{ fieldId: '5', value: 'o2' }] })
      )
    );
  });
  it('starts a custom field on its default and sends it', async () => {
    mocks.currentProjectId = '12';
    mocks.fields = [
      field({ id: '6', name: 'Team', type: 'select', required: true, defaultValue: 'o1' }),
    ];
    render(<QuickCreate />);
    expect(flowSelectTrigger('Team')).toHaveAccessibleName('Team: Low');
    fillTitle();
    create();
    await waitFor(() =>
      expect(mocks.create).toHaveBeenCalledWith(
        expect.objectContaining({ fieldValues: [{ fieldId: '6', value: 'o1' }] })
      )
    );
  });
});

function field(overrides: Partial<ApiField>): ApiField {
  return {
    id: '1',
    projectId: '12',
    name: 'Field',
    type: 'text',
    options: [
      { id: 'o1', label: 'Low', color: '#22c55e' },
      { id: 'o2', label: 'High', color: '#ef4444' },
    ],
    required: false,
    defaultValue: '',
    defaultValues: [],
    sortOrder: 0,
    valueCount: 0,
    ...overrides,
  };
}
