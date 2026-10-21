import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const inviteMemberMutate = vi.hoisted(() => vi.fn());

vi.mock('../useProjectMembers', () => ({
  useInviteProjectMember: () => ({ mutate: inviteMemberMutate, isPending: false }),
}));

import AddProjectMemberDialog from './AddProjectMemberDialog';
import { toast } from 'sonner';

function renderDialog(props: Partial<React.ComponentProps<typeof AddProjectMemberDialog>> = {}) {
  const onOpenChange = vi.fn();
  const utils = render(
    <AddProjectMemberDialog
      open
      onOpenChange={onOpenChange}
      projectId="p1"
      availableMembers={[
        {
          id: 'u1',
          name: 'Grace Hopper',
          email: 'grace@example.com',
          memberType: 'member',
          isOwner: false,
          hasFlowAccess: true,
          isActive: true,
        },
        {
          id: 'u2',
          name: 'Inactive Member',
          email: 'inactive@example.com',
          memberType: 'member',
          isOwner: false,
          hasFlowAccess: true,
          isActive: false,
        },
      ]}
      projectMembers={[]}
      {...props}
    />
  );
  return { onOpenChange, ...utils };
}

describe('AddProjectMemberDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders as a real dialog with a backdrop, not an inline panel', () => {
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
      'flow.projectSettings.addMemberDescription'
    );
  });

  it('does not render when closed', () => {
    renderDialog({ open: false });
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('finds workspace members by name and invites the selected member', () => {
    const { onOpenChange } = renderDialog();

    fireEvent.change(screen.getByPlaceholderText('flow.projectSettings.emailPlaceholder'), {
      target: { value: 'Grace' },
    });
    fireEvent.click(screen.getByRole('option', { name: /Grace Hopper grace@example.com/ }));
    fireEvent.change(screen.getByLabelText('flow.projectSettings.roleLabel'), {
      target: { value: 'admin' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.invite' }));

    expect(inviteMemberMutate).toHaveBeenCalledWith(
      { email: 'grace@example.com', role: 'admin' },
      expect.objectContaining({ onSuccess: expect.any(Function) })
    );

    // Simulate the mutation's onSuccess callback closing the dialog.
    const [, options] = inviteMemberMutate.mock.calls[0];
    options.onSuccess();
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('does not offer inactive or already-added members', () => {
    renderDialog({
      projectMembers: [
        {
          id: 'pm1',
          projectId: 'p1',
          userId: 'u1',
          name: 'Grace Hopper',
          email: 'grace@example.com',
          role: 'member',
          addedAt: '2026-01-01T00:00:00Z',
        },
      ],
    });
    const emailInput = screen.getByPlaceholderText('flow.projectSettings.emailPlaceholder');

    fireEvent.change(emailInput, { target: { value: 'grace' } });
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveTextContent('flow.projectSettings.noMemberMatches');
    expect(screen.getByRole('button', { name: 'flow.projectSettings.invite' })).toBeDisabled();

    fireEvent.change(emailInput, { target: { value: 'inactive' } });
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });

  it('requires an exact active workspace member email before enabling invite', () => {
    renderDialog();
    const emailInput = screen.getByPlaceholderText('flow.projectSettings.emailPlaceholder');
    const inviteButton = screen.getByRole('button', { name: 'flow.projectSettings.invite' });

    expect(inviteButton).toBeDisabled();
    fireEvent.change(emailInput, { target: { value: 'grace@' } });
    expect(inviteButton).toBeDisabled();
    fireEvent.change(emailInput, { target: { value: 'grace@example.com' } });
    expect(inviteButton).toBeEnabled();
  });

  it('rejects an address that does not belong to an active workspace member', () => {
    renderDialog();
    const emailInput = screen.getByPlaceholderText('flow.projectSettings.emailPlaceholder');
    fireEvent.change(emailInput, { target: { value: 'unknown@example.com' } });
    fireEvent.keyDown(emailInput, { key: 'Enter' });

    expect(inviteMemberMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('flowToasts.projectMember.selectWorkspaceMember');
  });

  it('labels the email input for screen readers', () => {
    renderDialog();

    expect(screen.getByLabelText('flow.projectSettings.emailPlaceholder')).toBeInTheDocument();
  });
});
