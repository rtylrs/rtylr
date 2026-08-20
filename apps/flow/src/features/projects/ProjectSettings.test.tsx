import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const updateProjectMutate = vi.hoisted(() => vi.fn());
const deleteProjectMutate = vi.hoisted(() => vi.fn());
const moveProjectMutate = vi.hoisted(() => vi.fn());
const removeMemberMutate = vi.hoisted(() => vi.fn());
const updateRoleMutate = vi.hoisted(() => vi.fn());
const transferOwnershipMutate = vi.hoisted(() => vi.fn());
// Mutable so the gating tests can render a members list with a row to
// inspect, while the pre-existing tests above keep the empty-list default.
const projectMembersData = vi.hoisted(() => ({
  data: [] as Array<{ id: string; userId: string; name: string; email: string; role: string }>,
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: 'p1', workspaceId: 'w1' }),
  useNavigate: () => vi.fn(),
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}));

vi.mock('./useProjects', () => ({
  useUpdateProject: () => ({ mutate: updateProjectMutate, isPending: false }),
  useDeleteProject: () => ({ mutate: deleteProjectMutate, isPending: false }),
  useMoveProject: () => ({ mutate: moveProjectMutate, isPending: false }),
}));

vi.mock('./useProjectMembers', () => ({
  useProjectMembers: () => ({ data: projectMembersData.data, isLoading: false }),
  useRemoveProjectMember: () => ({ mutate: removeMemberMutate, isPending: false }),
  useUpdateProjectMemberRole: () => ({ mutate: updateRoleMutate, isPending: false }),
  useTransferProjectOwnership: () => ({ mutate: transferOwnershipMutate, isPending: false }),
  useInviteProjectMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: [] }),
}));

// Defaults to project-admin so every pre-existing test below (which is about
// save/description behavior, not permissions) keeps exercising an unlocked
// screen. The gating describe block below flips this per test.
const canManageProjectMock = vi.hoisted(() => vi.fn(() => true));
vi.mock('../../shared/lib/currentUser', () => ({
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
    fireEvent.click(screen.getByRole('button', { name: 'Integrations' }));

    const comingSoonButtons = screen.getAllByRole('button', { name: 'common.comingSoon' });

    expect(comingSoonButtons).toHaveLength(3);
    comingSoonButtons.forEach((button) => expect(button).toBeDisabled());
  });

  it('sends the untouched project fields when saving a name-only edit', () => {
    renderProjectSettings();

    fireEvent.change(screen.getByDisplayValue('Demo project'), {
      target: { value: 'Renamed project' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

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
        onError: expect.any(Function),
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
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

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
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

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
    moveProjectMutate.mockClear();
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
    expect(screen.getByRole('button', { name: /Save changes/ })).toBeDisabled();
  });

  it('leaves rename/describe and Save changes enabled for a project admin', () => {
    canManageProjectMock.mockReturnValue(true);
    renderProjectSettings();

    expect(screen.getByDisplayValue('Demo project')).toBeEnabled();
    expect(
      screen.getByPlaceholderText('flow.projectSettings.descriptionPlaceholder')
    ).toBeEnabled();
    expect(screen.getByRole('button', { name: /Save changes/ })).toBeEnabled();
  });

  it('disables the member role select and remove button for a non-admin, and hides Add', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'Members' }));

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
    fireEvent.click(screen.getByRole('button', { name: 'Members' }));

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
    fireEvent.click(screen.getByRole('button', { name: 'Members' }));

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
    fireEvent.click(screen.getByRole('button', { name: 'Members' }));

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

  it('disables Delete for a non-admin so the type-to-confirm dialog can never open', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'Danger Zone' }));

    const deleteButton = screen.getByRole('button', { name: 'Delete' });
    expect(deleteButton).toBeDisabled();

    fireEvent.click(deleteButton);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(deleteProjectMutate).not.toHaveBeenCalled();
  });

  it('lets a project admin open the delete confirmation dialog', () => {
    canManageProjectMock.mockReturnValue(true);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'Danger Zone' }));

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('disables the transfer-ownership control for a non-admin', () => {
    canManageProjectMock.mockReturnValue(false);
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'Members' }));

    expect(screen.getByRole('button', { name: /Transfer/ })).toBeDisabled();
  });
});
