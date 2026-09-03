import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
  },
}));

type TeamMember = { id: string; name: string; email: string; role: string };

const teamData = vi.hoisted(() => ({
  data: { members: [] as TeamMember[], hasMore: false } as
    | { members: TeamMember[]; hasMore: boolean }
    | undefined,
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useTeam', () => ({
  useTeam: () => teamData,
  useDeactivateMember: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useAccessRequests: () => ({ data: [] }),
  useApproveAccessRequest: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useSubscription', () => ({
  useMySubscription: () => ({ data: undefined }),
  centsToDisplay: (cents: number) => `$${(cents / 100).toFixed(2)}`,
  computeSeatProration: () => ({
    daysRemaining: 0,
    daysInMonth: 30,
    proratedCents: 0,
    nextMonthName: 'January',
  }),
}));

vi.mock('../../shared/hooks/queries/useGroups', () => ({
  useGroups: () => ({ data: [] }),
  useAddGroupMembers: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useAddons', () => ({
  useAddons: () => ({ data: [] }),
  isAddonEnabled: () => false,
}));

import Team from './Team';

beforeEach(() => {
  vi.clearAllMocks();
  teamData.data = { members: [], hasMore: false };
  teamData.isLoading = false;
});

describe('Team list empty state', () => {
  it('shows the true-empty title, description and add-member call-to-action when there are no members', () => {
    render(<Team />);

    expect(screen.getByText('team.list.noMembersTitle')).toBeInTheDocument();
    expect(screen.getByText('team.list.noMembersDescription')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'common.nav.addMember' })).toHaveLength(2);
  });

  it('navigates to the invite screen when the true-empty call-to-action is clicked', () => {
    render(<Team />);

    fireEvent.click(screen.getAllByRole('button', { name: 'common.nav.addMember' })[1]);

    expect(navigateMock).toHaveBeenCalledWith({ to: '/team/invite' });
  });

  it('shows the search-empty title, description and clear-search action when a search matches nothing', () => {
    teamData.data = {
      members: [{ id: '1', name: 'Jane Doe', email: 'jane@rtylr.com', role: 'admin' }],
      hasMore: false,
    };
    render(<Team />);

    fireEvent.change(screen.getByPlaceholderText('Search by name or email...'), {
      target: { value: 'no such person' },
    });

    expect(screen.getByText('team.list.noMembersMatchTitle')).toBeInTheDocument();
    expect(screen.getByText('team.list.noMembersMatchDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.clearSearch' })).toBeInTheDocument();
  });

  it('clears the search query when the search-empty call-to-action is clicked', () => {
    teamData.data = {
      members: [{ id: '1', name: 'Jane Doe', email: 'jane@rtylr.com', role: 'admin' }],
      hasMore: false,
    };
    render(<Team />);

    const searchInput = screen.getByPlaceholderText('Search by name or email...');
    fireEvent.change(searchInput, { target: { value: 'no such person' } });
    expect(screen.getByText('team.list.noMembersMatchTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.clearSearch' }));

    expect(screen.queryByText('team.list.noMembersMatchTitle')).not.toBeInTheDocument();
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
  });

  it('does not show the empty state when members match', () => {
    teamData.data = {
      members: [{ id: '1', name: 'Jane Doe', email: 'jane@rtylr.com', role: 'admin' }],
      hasMore: false,
    };
    render(<Team />);

    expect(screen.queryByText('team.list.noMembersTitle')).not.toBeInTheDocument();
  });
});
