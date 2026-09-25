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
    <AddProjectMemberDialog open onOpenChange={onOpenChange} projectId="p1" {...props} />
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

  it('invites by email and role, and closes the dialog on success', () => {
    const { onOpenChange } = renderDialog();

    fireEvent.change(screen.getByPlaceholderText('flow.projectSettings.emailPlaceholder'), {
      target: { value: 'new.hire@example.com' },
    });
    fireEvent.change(screen.getByLabelText('flow.projectSettings.roleLabel'), {
      target: { value: 'admin' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.invite' }));

    expect(inviteMemberMutate).toHaveBeenCalledWith(
      { email: 'new.hire@example.com', role: 'admin' },
      expect.objectContaining({ onSuccess: expect.any(Function) })
    );

    // Simulate the mutation's onSuccess callback closing the dialog.
    const [, options] = inviteMemberMutate.mock.calls[0];
    options.onSuccess();
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('shows an error toast and does not submit when the email is blank', () => {
    renderDialog();

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectSettings.invite' }));

    expect(inviteMemberMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('toasts.projectMember.enterEmail');
  });

  it('labels the email input for screen readers', () => {
    renderDialog();

    expect(screen.getByLabelText('flow.projectSettings.emailPlaceholder')).toBeInTheDocument();
  });
});
