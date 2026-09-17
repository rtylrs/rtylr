import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  role: 'operator' as string,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (selector: (state: { user: { role: string } }) => unknown) =>
    selector({ user: { role: mocks.role } }),
}));

const refetchTeam = vi.fn();
const teamMembers: unknown[] = [];

vi.mock('../../shared/hooks/queries/useCrmSettings', () => ({
  useCrmSettings: () => ({ data: undefined }),
  useUpdateCrmSettings: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useCrmTeam', () => ({
  useCrmTeam: () => ({ data: teamMembers, isLoading: false, refetch: refetchTeam }),
  useCrmInviteMember: () => ({ mutate: vi.fn(), isPending: false }),
  useCrmDeactivateMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useCrmSubscription', () => ({
  useCrmSubscription: () => ({ data: undefined }),
  centsToDisplay: (cents: number) => `$${(cents / 100).toFixed(2)}`,
  computeSeatProration: () => ({
    daysRemaining: 0,
    daysInMonth: 30,
    proratedCents: 0,
    nextMonthName: 'January',
  }),
}));

import SettingsScreen from './Settings';

function renderSettings(): void {
  render(<SettingsScreen />);
  fireEvent.click(screen.getByRole('button', { name: 'crm.settings.tabTeam' }));
}

beforeEach(() => {
  refetchTeam.mockClear();
  teamMembers.length = 0;
});

describe('Settings team empty state', () => {
  it('shows an "invite member" call-to-action for a role that can invite', () => {
    mocks.role = 'admin';
    renderSettings();

    expect(screen.getByText('crm.team.emptyTitle')).toBeInTheDocument();

    // Header and empty-state actions share one key now, so both carry the same
    // accessible name. Assert both exist rather than matching one by luck.
    const buttons = screen.getAllByRole('button', { name: 'crm.team.inviteMember' });
    expect(buttons).toHaveLength(2);
    fireEvent.click(buttons[1]);

    // Opens the inline invite form rather than mutating anything itself.
    expect(screen.getByText('crm.settings.inviteSomeone')).toBeInTheDocument();
  });

  it('falls back to a "refresh" call-to-action for a role that cannot invite', () => {
    mocks.role = 'operator';
    renderSettings();

    expect(screen.getByText('crm.team.emptyTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));

    expect(refetchTeam).toHaveBeenCalledTimes(1);
  });
});
