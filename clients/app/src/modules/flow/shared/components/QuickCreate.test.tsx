import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { initTestI18n } from '../testing/i18n';
import { useUserStore } from '../stores/userStore';
import { useFlowStore } from '../stores/flowStore';
import QuickCreate from './QuickCreate';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  create: vi.fn(),
  currentProjectId: undefined as string | undefined,
  labels: [] as Array<{ id: number; name: string }>,
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

beforeEach(async () => {
  vi.clearAllMocks();
  await initTestI18n();
  mocks.currentProjectId = undefined;
  mocks.labels = [];
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
  useFlowStore.setState({ quickCreateOpen: true, projects: [] });
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
});
