import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Settings from './Settings';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  role: 'admin' as string,
  useErpTeam: vi.fn(),
  refetch: vi.fn(),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

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

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Settings team empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useErpTeam.mockReturnValue({ data: [], isLoading: false, refetch: mocks.refetch });
  });

  it('shows the invite action and opens the invite form for a role with hub:team:invite', () => {
    mocks.role = 'admin';
    render(<Settings />);
    openTeamTab();

    const title = screen.getByText('No team members yet.');
    expect(title).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Invite member' }));

    expect(screen.getByText('Invite someone to ERP')).toBeInTheDocument();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('falls back to Refresh for a role without hub:team:invite', () => {
    mocks.role = 'operator';
    render(<Settings />);
    openTeamTab();

    const title = screen.getByText('No team members yet.');
    const emptyState = title.parentElement as HTMLElement;

    expect(
      within(emptyState).queryByRole('button', { name: 'Invite member' })
    ).not.toBeInTheDocument();

    fireEvent.click(within(emptyState).getByRole('button', { name: 'Refresh' }));

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

    expect(screen.queryByText('No team members yet.')).not.toBeInTheDocument();
  });
});
