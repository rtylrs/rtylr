import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
  Trans: ({ i18nKey }: { i18nKey: string }) => i18nKey,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

const inviteMock = vi.hoisted(() => vi.fn());
vi.mock('../../shared/hooks/queries/useTeam', () => ({
  useInviteMember: () => ({ mutateAsync: inviteMock, isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useGroups', () => ({
  useGroups: () => ({ data: [] }),
  useBulkInvite: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/usePOSAccess', () => ({
  usePOSLocations: () => ({ data: [] }),
  useAssignPOSLocation: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock('../../shared/hooks/queries/useSubscription', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../shared/hooks/queries/useSubscription')>()),
  useMySubscription: () => ({ data: undefined }),
}));

vi.mock('../../shared/hooks/queries/usePermissions', () => ({
  usePermissionProfiles: () => ({ data: [] }),
}));

vi.mock('./PermissionProfilePicker', () => ({
  PermissionProfilePicker: () => null,
}));

import InviteUser from './InviteUser';

describe('InviteUser single-member wizard', () => {
  beforeEach(() => {
    inviteMock.mockReset().mockResolvedValue({ userId: '1' });
    navigateMock.mockReset();
  });

  it('shows three steps and only invites from the review step', async () => {
    render(<InviteUser />);

    expect(screen.queryByText('team.inviteUser.stepActivate')).toBeNull();

    fireEvent.change(screen.getByLabelText('team.inviteUser.fullNameLabel'), {
      target: { value: 'Jane Doe' },
    });
    fireEvent.change(screen.getByLabelText('team.inviteUser.emailLabel'), {
      target: { value: 'jane@example.com' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'team.inviteUser.nextButton' }));
    fireEvent.click(screen.getByRole('button', { name: 'team.inviteUser.nextButton' }));
    await screen.findByText('team.inviteUser.reviewTitle');

    expect(inviteMock).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: 'team.inviteUser.sendInvitation' }));

    await waitFor(() => expect(inviteMock).toHaveBeenCalledTimes(1));
    expect(inviteMock).toHaveBeenCalledWith(
      expect.objectContaining({
        email: 'jane@example.com',
        fullName: 'Jane Doe',
        provisioningMode: 'invitation',
      })
    );
  });

  it('shows missing identity fields inline instead of advancing', () => {
    render(<InviteUser />);

    fireEvent.click(screen.getByRole('button', { name: 'team.inviteUser.nextButton' }));

    const nameInput = screen.getByLabelText('team.inviteUser.fullNameLabel');
    const emailInput = screen.getByLabelText('team.inviteUser.emailLabel');
    expect(screen.getByText('team.inviteUser.nameRequired')).toBeTruthy();
    expect(screen.getByText('team.inviteUser.emailRequired')).toBeTruthy();
    expect(nameInput.getAttribute('aria-invalid')).toBe('true');
    expect(nameInput.getAttribute('aria-describedby')).toBe('inv-name-error');
    expect(document.activeElement).toBe(nameInput);
    // Still on Identity.
    expect(screen.queryByText('team.inviteUser.reviewTitle')).toBeNull();

    fireEvent.change(nameInput, { target: { value: 'Jane Doe' } });
    expect(screen.queryByText('team.inviteUser.nameRequired')).toBeNull();
    expect(nameInput.getAttribute('aria-invalid')).toBeNull();

    fireEvent.change(emailInput, { target: { value: 'jane@' } });
    fireEvent.click(screen.getByRole('button', { name: 'team.inviteUser.nextButton' }));
    expect(screen.getByText('team.inviteUser.emailInvalid')).toBeTruthy();
    expect(document.activeElement).toBe(emailInput);
  });
});
