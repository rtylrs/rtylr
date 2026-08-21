import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { FlowTeamMember } from '../../shared/hooks/queries/useFlowTeam';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${Object.values(params).join(',')}` : key,
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const addMemberMutate = vi.hoisted(() => vi.fn());

vi.mock('./useWorkspaceManager', () => ({
  useAddWorkspaceMember: () => ({ mutate: addMemberMutate, isPending: false }),
}));

import AddWorkspaceMemberDialog from './AddWorkspaceMemberDialog';
import { toast } from 'sonner';

function orgMember(over: Partial<FlowTeamMember> = {}): FlowTeamMember {
  return {
    id: '1',
    name: 'Jane Doe',
    email: 'jane@example.com',
    role: 'member',
    apps: [],
    isActive: true,
    ...over,
  };
}

const ORG_MEMBERS = [
  orgMember({ id: '1', name: 'Jane Doe', email: 'jane@example.com' }),
  orgMember({ id: '2', name: 'John Smith', email: 'john@example.com' }),
];

function renderDialog(props: Partial<React.ComponentProps<typeof AddWorkspaceMemberDialog>> = {}) {
  const onOpenChange = vi.fn();
  const utils = render(
    <AddWorkspaceMemberDialog
      open
      onOpenChange={onOpenChange}
      workspaceId="1"
      orgMembers={ORG_MEMBERS}
      existingUserIds={new Set()}
      {...props}
    />
  );
  return { onOpenChange, ...utils };
}

describe('AddWorkspaceMemberDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders as a real dialog with a backdrop, not an inline expand panel', () => {
    renderDialog();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  // Without a DialogDescription, Radix warns "Missing Description or
  // aria-describedby" and a screen reader announces the dialog with a title and
  // no statement of what it does. Asserting the key, not English, also catches a
  // description that was added as a hardcoded string.
  it('gives the dialog an accessible description', () => {
    renderDialog();
    expect(screen.getByRole('dialog')).toHaveAccessibleDescription(
      'workspaceManager.addMemberDescription'
    );
  });

  it('does not render when closed', () => {
    renderDialog({ open: false });
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('excludes members already in the workspace from search results', () => {
    renderDialog({ existingUserIds: new Set(['1']) });

    fireEvent.change(screen.getByPlaceholderText('workspaceManager.searchPlaceholder'), {
      target: { value: 'e' },
    });

    expect(screen.queryByText('Jane Doe')).not.toBeInTheDocument();
    expect(screen.getByText('John Smith')).toBeInTheDocument();
  });

  it('adds the selected member with the chosen role and closes on success', () => {
    const { onOpenChange } = renderDialog();

    fireEvent.change(screen.getByPlaceholderText('workspaceManager.searchPlaceholder'), {
      target: { value: 'John' },
    });
    fireEvent.click(screen.getByText('John Smith'));

    fireEvent.change(screen.getByLabelText('workspaceManager.roleLabel'), {
      target: { value: 'admin' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'workspaceManager.addToWorkspace' }));

    expect(addMemberMutate).toHaveBeenCalledWith(
      { workspace_id: 1, user_id: 2, role: 'admin' },
      expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) })
    );

    const [, options] = addMemberMutate.mock.calls[0];
    options.onSuccess();
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('keeps the add-to-workspace action disabled until a member is selected', () => {
    renderDialog();

    expect(screen.getByRole('button', { name: 'workspaceManager.addToWorkspace' })).toBeDisabled();
    expect(addMemberMutate).not.toHaveBeenCalled();
    expect(toast.error).not.toHaveBeenCalled();
  });

  it('shows the empty title with a Clear search CTA when a search matches no one', () => {
    renderDialog();

    fireEvent.change(screen.getByPlaceholderText('workspaceManager.searchPlaceholder'), {
      target: { value: 'nobody-matches-this' },
    });

    expect(screen.getByText('workspaceManager.noOrgMembersFound')).toBeInTheDocument();
    const clear = screen.getByRole('button', { name: 'workspaceManager.clearSearchCta' });

    fireEvent.click(clear);
    expect(screen.getByPlaceholderText('workspaceManager.searchPlaceholder')).toHaveValue('');
  });
});
