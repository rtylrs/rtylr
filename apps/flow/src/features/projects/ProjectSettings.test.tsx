import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
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
  useUpdateProject: () => ({ mutate: vi.fn(), isPending: false }),
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

function renderProjectSettings() {
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
        health: 'on_track',
        progress: 0,
        startDate: '2026-01-01',
        targetDate: '2026-06-01',
        lead: 'u1',
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
  });

  it('shows a translated reason on each disabled Connect button instead of a silent disabled control', () => {
    renderProjectSettings();
    fireEvent.click(screen.getByRole('button', { name: 'Integrations' }));

    const comingSoonButtons = screen.getAllByRole('button', { name: 'common.comingSoon' });

    expect(comingSoonButtons).toHaveLength(3);
    comingSoonButtons.forEach((button) => expect(button).toBeDisabled());
  });
});
