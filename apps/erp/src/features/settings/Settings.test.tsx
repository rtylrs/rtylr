import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Settings from './Settings';

const mocks = vi.hoisted(() => ({
  role: 'admin' as string,
  useErpTeam: vi.fn(),
  refetch: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (select: (state: { user: { role: string } }) => unknown) =>
    select({ user: { role: mocks.role } }),
}));

vi.mock('../../shared/hooks/queries/useErpSettings', () => ({
  useErpSettings: () => ({
    data: undefined,
    isError: false,
    isFetching: false,
    isLoading: true,
    refetch: vi.fn(),
  }),
  useUpdateErpSettings: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../warehouses/useWarehouses', () => ({
  useWarehouses: () => ({ data: [] }),
}));

vi.mock('../../shared/hooks/queries/useErpTeam', () => ({
  useErpTeam: mocks.useErpTeam,
  useErpInviteMember: () => ({ mutate: vi.fn(), isPending: false }),
  useErpDeactivateMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useErpSubscription', () => ({
  useErpSubscription: () => ({ data: undefined }),
  centsToDisplay: (cents: number) => `$${(cents / 100).toFixed(2)}`,
  computeSeatProration: () => ({
    daysRemaining: 0,
    daysInMonth: 30,
    proratedCents: 0,
    nextMonthName: 'January',
  }),
}));

function openTeamTab(): void {
  fireEvent.click(screen.getByRole('button', { name: 'Team' }));
}

describe('Settings team empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useErpTeam.mockReturnValue({ data: [], isLoading: false, refetch: mocks.refetch });
  });

  it('shows the invite action and opens the invite form for a role with hub:team:invite', () => {
    mocks.role = 'admin';
    render(<Settings />);
    openTeamTab();

    const title = screen.getByText('settings.team.empty.title');
    expect(title).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'settings.team.invite' }));

    expect(screen.getByText('Invite someone to ERP')).toBeInTheDocument();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('falls back to Refresh for a role without hub:team:invite', () => {
    mocks.role = 'operator';
    render(<Settings />);
    openTeamTab();

    const title = screen.getByText('settings.team.empty.title');
    const emptyState = title.parentElement as HTMLElement;

    expect(
      within(emptyState).queryByRole('button', { name: 'settings.team.invite' })
    ).not.toBeInTheDocument();

    fireEvent.click(within(emptyState).getByRole('button', { name: 'settings.team.refresh' }));

    expect(mocks.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when team members exist', () => {
    mocks.useErpTeam.mockReturnValue({
      data: [{ id: 'u1', name: 'Jane Smith', email: 'jane@example.com', role: 'staff' }],
      isLoading: false,
      refetch: mocks.refetch,
    });
    mocks.role = 'admin';

    render(<Settings />);
    openTeamTab();

    expect(screen.queryByText('settings.team.empty.title')).not.toBeInTheDocument();
  });
});
