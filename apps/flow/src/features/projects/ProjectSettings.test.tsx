import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const updateProjectMutate = vi.hoisted(() => vi.fn());

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
  useDeleteProject: () => ({ mutate: vi.fn(), isPending: false }),
  useMoveProject: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('./useProjectMembers', () => ({
  useProjectMembers: () => ({ data: [], isLoading: false }),
  useRemoveProjectMember: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateProjectMemberRole: () => ({ mutate: vi.fn(), isPending: false }),
  useTransferProjectOwnership: () => ({ mutate: vi.fn(), isPending: false }),
  useInviteProjectMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: [] }),
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
