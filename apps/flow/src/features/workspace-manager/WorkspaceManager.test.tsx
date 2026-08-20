import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiWorkspaceMember } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${Object.values(params).join(',')}` : key,
  }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ workspaceId: 'ws-1' }),
}));

const membersState = vi.hoisted(() => ({ data: [] as ApiWorkspaceMember[], isLoading: false }));
vi.mock('./useWorkspaceManager', () => ({
  useWorkspaceMembers: () => membersState,
  useRemoveWorkspaceMember: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateWorkspaceMemberRole: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: [] }),
}));

// Rendered inline in WorkspaceManager but tested on its own - stub it here so
// this file only exercises the empty-state wiring, not the dialog internals.
vi.mock('./AddWorkspaceMemberDialog', () => ({
  default: ({ open }: { open: boolean }) => (open ? <div data-testid="add-member-dialog" /> : null),
}));

import WorkspaceManager from './WorkspaceManager';

beforeEach(() => {
  vi.clearAllMocks();
  membersState.data = [];
  membersState.isLoading = false;
});

describe('WorkspaceManager empty state', () => {
  it('shows the empty title with an Add member CTA that opens the dialog', () => {
    render(<WorkspaceManager />);

    expect(screen.getByText('workspaceManager.noMembersEmpty')).toBeInTheDocument();
    expect(screen.queryByTestId('add-member-dialog')).not.toBeInTheDocument();

    const addButtons = screen.getAllByRole('button', { name: 'workspaceManager.addMember' });
    expect(addButtons.length).toBeGreaterThan(0);
    fireEvent.click(addButtons[addButtons.length - 1]);

    expect(screen.getByTestId('add-member-dialog')).toBeInTheDocument();
  });
});
