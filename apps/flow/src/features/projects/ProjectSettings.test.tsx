import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { toast } from 'sonner';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const updateProjectMutate = vi.hoisted(() => vi.fn());
const deleteProjectMutate = vi.hoisted(() => vi.fn());
const removeMemberMutate = vi.hoisted(() => vi.fn());
const updateRoleMutate = vi.hoisted(() => vi.fn());
const transferOwnershipMutate = vi.hoisted(() => vi.fn());
// Mutable so the gating tests can render a members list with a row to
// inspect, while the pre-existing tests above keep the empty-list default.
const projectMembersData = vi.hoisted(() => ({
  data: [] as Array<{ id: string; userId: string; name: string; email: string; role: string }>,
}));
const refetchMembersMock = vi.hoisted(() => vi.fn());

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: 'p1', workspaceId: 'w1' }),
  useNavigate: () => vi.fn(),
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}));

let projectsQueryState = { isError: false, isPaused: false, failureCount: 0, refetch: vi.fn() };

vi.mock('./useProjects', () => ({
  useUpdateProject: () => ({ mutate: updateProjectMutate, isPending: false }),
  useArchiveProject: () => ({ mutate: deleteProjectMutate, isPending: false }),
  // Settings reads the projects query only to tell a failed fetch from a
  // deleted project. A settled, successful query keeps these cases on the
  // normal path; the failure case has its own test below.
  useProjects: () => projectsQueryState,
}));

vi.mock('./useProjectMembers', () => ({
  useProjectMembers: () => ({
    data: projectMembersData.data,
    isLoading: false,
    refetch: refetchMembersMock,
  }),
  useRemoveProjectMember: () => ({ mutate: removeMemberMutate, isPending: false }),
  useUpdateProjectMemberRole: () => ({ mutate: updateRoleMutate, isPending: false }),
  useTransferProjectOwnership: () => ({ mutate: transferOwnershipMutate, isPending: false }),
  useInviteProjectMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({
    data: [{ id: '20', name: 'Grace Hopper', email: 'grace@example.com' }],
  }),
}));

// Defaults to project-admin so every pre-existing test below (which is about
// save/description behavior, not permissions) keeps exercising an unlocked
// screen. The gating describe block below flips this per test.
const canManageProjectMock = vi.hoisted(() => vi.fn(() => true));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanAdministerProject: () => canManageProjectMock(),
  useCanDeleteResource: () => canManageProjectMock(),
}));

import ProjectSettings from './ProjectSettings';
import { useFlowStore } from '../../shared/stores/flowStore';

function renderProjectSettings(description = 'Existing description') {
  useFlowStore.setState({
    projects: [
      {
        id: 'p1',
        workspaceId: 'w1',
        name: 'Demo project',
        color: '#3b82f6',
        taskCount: 0,
        completedCount: 0,
        memberAvatars: [],
        description,
        status: 'archived',
        health: 'on_track',
        progress: 0,
        startDate: '2026-01-01',
        targetDate: '2026-06-01',
        lead: 'u1',
        icon: 'briefcase',
      },
    ],
    workspaces: [{ id: 'w1', name: 'Workspace', color: '#000' }],
  });

  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <ProjectSettings />
    </QueryClientProvider>
  );
}

describe('ProjectSettings integrations tab', () => {
  beforeEach(() => {
    useFlowStore.setState({ projects: [], workspaces: [] });
    updateProjectMutate.mockClear();
    canManageProjectMock.mockReturnValue(true);
    projectMembersData.data = [];
  });

  it('shows a translated reason on each disabled Connect button instead of a silent disabled control', () => {
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.integrations' }));

    const comingSoonButtons = screen.getAllByRole('button', { name: 'common.comingSoon' });

    expect(comingSoonButtons).toHaveLength(3);
    comingSoonButtons.forEach((button) => expect(button).toBeDisabled());
  });

  it('sends the untouched project fields when saving a name-only edit', () => {
    renderProjectSettings();

    fireEvent.change(screen.getByDisplayValue('Demo project'), {
      target: { value: 'Renamed project' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.saveChanges' }));

    expect(updateProjectMutate).toHaveBeenCalledWith(
      {
        id: 'p1',
        data: {
          name: 'Renamed project',
          description: 'Existing description',
          icon: 'briefcase',
          color: '#3b82f6',
          status: 'archived',
          startDate: '2026-01-01',
          dueDate: '2026-06-01',
        },
      },
      expect.objectContaining({
        onSuccess: expect.any(Function),
      })
    );
  });

  it('renders a "-" sentinel description as an empty field with placeholder text, not a literal dash', () => {
    renderProjectSettings('-');

    const descriptionField = screen.getByPlaceholderText(
      'flow.projectSettings.descriptionPlaceholder'
    );

    expect(descriptionField).toHaveValue('');
  });

  it('renders a real description normally', () => {
    renderProjectSettings('A project about onboarding flows');

    const descriptionField = screen.getByPlaceholderText(
      'flow.projectSettings.descriptionPlaceholder'
    );

    expect(descriptionField).toHaveValue('A project about onboarding flows');
  });

  it('saves the "-" sentinel (not an empty string) when the description was never filled in', () => {
    renderProjectSettings('-');
    fireEvent.click(screen.getByRole('button', { name: 'common.saveChanges' }));

    expect(updateProjectMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        data: expect.objectContaining({ description: '-' }),
      }),
      expect.anything()
    );
  });

  it('saves a typed description as-is', () => {
    renderProjectSettings('-');

    fireEvent.change(screen.getByPlaceholderText('flow.projectSettings.descriptionPlaceholder'), {
      target: { value: 'New real description' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.saveChanges' }));

    expect(updateProjectMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        data: expect.objectContaining({ description: 'New real description' }),
      }),
      expect.anything()
    );
  });
});

// #793: none of these controls had any role check at all, so a project
// viewer or member could rename/delete/transfer the project or promote
// themselves to admin via the role select. These tests pin the fix: every
// control is disabled (not merely relying on a hidden mutate call) for a
// non-admin, and stays fully usable for a project admin.
describe('ProjectSettings role gating (#793)', () => {
  beforeEach(() => {
    useFlowStore.setState({ projects: [], workspaces: [] });
    updateProjectMutate.mockClear();
    deleteProjectMutate.mockClear();
    removeMemberMutate.mockClear();
    updateRoleMutate.mockClear();
    transferOwnershipMutate.mockClear();
    projectMembersData.data = [
      { id: 'pm1', userId: '7', name: 'Ada Lovelace', email: 'ada@example.com', role: 'viewer' },
    ];
  });

  it('disables rename/describe and Save changes for a non-admin', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();

    expect(screen.getByDisplayValue('Demo project')).toBeDisabled();
    expect(
      screen.getByPlaceholderText('flow.projectSettings.descriptionPlaceholder')
    ).toBeDisabled();
    expect(screen.getByRole('button', { name: 'common.saveChanges' })).toBeDisabled();
  });

  it('leaves rename/describe and Save changes enabled for a project admin', () => {
    canManageProjectMock.mockReturnValue(true);
    renderProjectSettings();

    expect(screen.getByDisplayValue('Demo project')).toBeEnabled();
    expect(
      screen.getByPlaceholderText('flow.projectSettings.descriptionPlaceholder')
    ).toBeEnabled();
    expect(screen.getByRole('button', { name: 'common.saveChanges' })).toBeEnabled();
  });

  it('disables the member role select and remove button for a non-admin, and hides Add', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));

    expect(
      screen.getByRole('combobox', { name: 'flow.projectSettings.roleSelectLabel' })
    ).toBeDisabled();
    expect(
      screen.getByRole('button', { name: 'flow.projectSettings.removeMemberLabel' })
    ).toBeDisabled();
    expect(screen.queryByRole('button', { name: /Add/ })).not.toBeInTheDocument();
  });

  it('a non-admin cannot fire a role change even if the disabled select is programmatically changed', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));

    fireEvent.change(
      screen.getByRole('combobox', { name: 'flow.projectSettings.roleSelectLabel' }),
      {
        target: { value: 'admin' },
      }
    );

    expect(updateRoleMutate).not.toHaveBeenCalled();
  });

  it('leaves the member role select and remove button enabled for a project admin', () => {
    canManageProjectMock.mockReturnValue(true);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));

    expect(
      screen.getByRole('combobox', { name: 'flow.projectSettings.roleSelectLabel' })
    ).toBeEnabled();
    expect(
      screen.getByRole('button', { name: 'flow.projectSettings.removeMemberLabel' })
    ).toBeEnabled();
    expect(screen.getByRole('button', { name: /Add/ })).toBeInTheDocument();
  });

  it('a project admin can change a member role', () => {
    canManageProjectMock.mockReturnValue(true);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));

    fireEvent.change(
      screen.getByRole('combobox', { name: 'flow.projectSettings.roleSelectLabel' }),
      {
        target: { value: 'admin' },
      }
    );

    expect(updateRoleMutate).toHaveBeenCalledWith(
      { project_id: NaN, user_id: 7, role: 'admin' },
      expect.anything()
    );
  });

  it('disables Archive for a non-admin so the type-to-confirm dialog can never open', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.danger' }));

    const deleteButton = screen.getByRole('button', { name: 'flow.projectSettings.archiveAction' });
    expect(deleteButton).toBeDisabled();

    fireEvent.click(deleteButton);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(deleteProjectMutate).not.toHaveBeenCalled();
  });

  it('lets a project admin open the archive confirmation dialog', () => {
    canManageProjectMock.mockReturnValue(true);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.danger' }));

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.archiveAction' }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('disables the transfer-ownership control for a non-admin', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));

    expect(screen.getByRole('button', { name: 'flow.projectSettings.transferCta' })).toBeDisabled();
  });
});

describe('ProjectSettings members empty state', () => {
  beforeEach(() => {
    useFlowStore.setState({ projects: [], workspaces: [] });
    refetchMembersMock.mockClear();
    projectMembersData.data = [];
  });

  it('offers Add member as the CTA for a project admin', () => {
    canManageProjectMock.mockReturnValue(true);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));

    expect(screen.getByText('flow.projectSettings.noMembersYet')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.projectSettings.addMemberCta' })).toBeEnabled();
  });

  it('offers Refresh instead for a non-admin', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));

    expect(
      screen.queryByRole('button', { name: 'flow.projectSettings.addMemberCta' })
    ).not.toBeInTheDocument();
    const refresh = screen.getByRole('button', { name: 'flow.projectSettings.refreshMembersCta' });
    fireEvent.click(refresh);
    expect(refetchMembersMock).toHaveBeenCalled();
  });
});

describe('ProjectSettings mutation error toasts (no double-toast on failure)', () => {
  beforeEach(() => {
    useFlowStore.setState({ projects: [], workspaces: [] });
    // A member row has to render for the remove/role/transfer controls to
    // exist at all. Without it these assertions would pass vacuously against
    // a screen that never drew the button they claim to exercise.
    projectMembersData.data = [
      {
        id: 'm1',
        userId: '10',
        name: 'Ada Lovelace',
        email: 'ada@example.com',
        role: 'member',
      },
    ];
    // The role-gating block above leaves this returning false, which would
    // disable every control here and make the assertions pass vacuously.
    canManageProjectMock.mockReturnValue(true);
    updateProjectMutate.mockClear();
    removeMemberMutate.mockClear();
    updateRoleMutate.mockClear();
    transferOwnershipMutate.mockClear();
    deleteProjectMutate.mockClear();
    (toast.error as ReturnType<typeof vi.fn>).mockClear();
    (toast.success as ReturnType<typeof vi.fn>).mockClear();
  });

  it('does not show a call-site toast when removing a member fails - the hook already shows one', () => {
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.removeMemberLabel' }));

    const [, options] = removeMemberMutate.mock.calls[0];
    act(() => options.onError?.(new Error('boom')));

    expect(toast.error).not.toHaveBeenCalled();
  });

  it('does not show a call-site toast when updating a member role fails - the hook already shows one', () => {
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));
    fireEvent.change(
      screen.getByRole('combobox', { name: 'flow.projectSettings.roleSelectLabel' }),
      { target: { value: 'admin' } }
    );

    const [, options] = updateRoleMutate.mock.calls[0];
    act(() => options.onError?.(new Error('boom')));

    expect(toast.error).not.toHaveBeenCalled();
  });

  it('does not show a call-site toast when transferring ownership fails, but still resets the confirm state', () => {
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.members' }));

    fireEvent.change(document.getElementById('transfer-select')!, {
      target: { value: '20' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.transferCta' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.confirm' }));

    const [, options] = transferOwnershipMutate.mock.calls[0];
    act(() => options.onError?.(new Error('boom')));

    expect(toast.error).not.toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: 'Confirm' })).not.toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'flow.projectSettings.transferCta' })
    ).toBeInTheDocument();
  });

  it('does not show a call-site toast when deleting a project fails - the hook already shows one', () => {
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.tabs.danger' }));
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.archiveAction' }));
    fireEvent.change(screen.getByPlaceholderText('Demo project'), {
      target: { value: 'Demo project' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.archiveConfirm' }));

    const [, options] = deleteProjectMutate.mock.calls[0];
    act(() => options.onError?.(new Error('boom')));

    expect(toast.error).not.toHaveBeenCalled();
  });

  it('does not show a call-site toast when saving a project update fails - the hook already shows one', () => {
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'common.saveChanges' }));

    const [, options] = updateProjectMutate.mock.calls[0];
    expect(options.onError).toBeUndefined();
  });
});

describe('ProjectSettings when the projects request fails', () => {
  beforeEach(() => {
    useFlowStore.setState({ projects: [], workspaces: [] });
  });

  afterEach(() => {
    projectsQueryState = { isError: false, isPaused: false, failureCount: 0, refetch: vi.fn() };
  });

  it('offers a retry instead of reporting the project deleted', () => {
    // The store is filled by this same query, so a failed fetch leaves it
    // empty and settings used to say the project was gone.
    const refetch = vi.fn();
    projectsQueryState = { isError: true, isPaused: false, failureCount: 1, refetch };
    renderProjectSettings();

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.project.notFound')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(refetch).toHaveBeenCalled();
  });

  it('reports a paused query that already failed as a failure, not as loading', () => {
    // networkMode 'online' can park a failed query at pending/paused, which
    // never reaches isError and used to fall through to the empty branch.
    projectsQueryState = { isError: false, isPaused: true, failureCount: 1, refetch: vi.fn() };
    renderProjectSettings();
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
  });

  it('still reports a genuinely missing project as not found', () => {
    // Both directions matter: without this the branch above could swallow a
    // real deletion, and the screen would invite a pointless retry forever.
    // Rendered directly rather than through the helper, which seeds a project.
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <ProjectSettings />
      </QueryClientProvider>
    );

    expect(screen.getByText('flow.project.notFound')).toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });
});
