import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';

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

type TeamMember = {
  id: string;
  name: string;
  email: string;
  memberType: string;
  licensedAppIds: string[];
  permissionProfileAssignments: unknown[];
  effectiveAccess: { apps: { appId: string; licensed: boolean }[] };
  status: string;
};

const teamData = vi.hoisted(() => ({
  data: { members: [] as TeamMember[], hasMore: false } as
    { members: TeamMember[]; hasMore: boolean } | undefined,
  isLoading: false,
  isError: false,
  refetch: vi.fn(),
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

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(() => {
  vi.clearAllMocks();
  teamData.data = { members: [], hasMore: false };
  teamData.isLoading = false;
  teamData.isError = false;
});

describe('Team list empty state', () => {
  it('shows a retryable load error instead of an empty team or member count', () => {
    teamData.data = undefined;
    teamData.isError = true;
    render(<Team />);

    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load team members');
    expect(screen.queryByText('No team members yet')).not.toBeInTheDocument();
    expect(screen.queryByText('0 members')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(teamData.refetch).toHaveBeenCalledOnce();
  });

  it('shows the true-empty title, description and add-member call-to-action when there are no members', () => {
    render(<Team />);

    expect(screen.getByText(i18next.t('team.list.noMembersTitle'))).toBeInTheDocument();
    expect(screen.getByText(i18next.t('team.list.noMembersDescription'))).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: i18next.t('common.nav.addMember') })).toHaveLength(
      2
    );
  });

  it('navigates to the invite screen when the true-empty call-to-action is clicked', () => {
    render(<Team />);

    fireEvent.click(screen.getAllByRole('button', { name: i18next.t('common.nav.addMember') })[1]);

    expect(navigateMock).toHaveBeenCalledWith({ to: '/team/invite' });
  });

  it('shows the search-empty title, description and clear-search action when a search matches nothing', () => {
    teamData.data = {
      members: [
        {
          id: '1',
          name: 'Jane Doe',
          email: 'jane@rtylr.com',
          memberType: 'admin',
          licensedAppIds: [],
          permissionProfileAssignments: [],
          effectiveAccess: { apps: [] },
          status: 'active',
        },
      ],
      hasMore: false,
    };
    render(<Team />);

    fireEvent.change(screen.getByPlaceholderText(i18next.t('team.list.searchPlaceholder')), {
      target: { value: 'no such person' },
    });

    expect(screen.getByText(i18next.t('team.list.noMembersMatchTitle'))).toBeInTheDocument();
    expect(screen.getByText(i18next.t('team.list.noMembersMatchDescription'))).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: i18next.t('common.clearSearch') })
    ).toBeInTheDocument();
  });

  it('clears the search query when the search-empty call-to-action is clicked', () => {
    teamData.data = {
      members: [
        {
          id: '1',
          name: 'Jane Doe',
          email: 'jane@rtylr.com',
          memberType: 'admin',
          licensedAppIds: [],
          permissionProfileAssignments: [],
          effectiveAccess: { apps: [] },
          status: 'active',
        },
      ],
      hasMore: false,
    };
    render(<Team />);

    const searchInput = screen.getByPlaceholderText(i18next.t('team.list.searchPlaceholder'));
    fireEvent.change(searchInput, { target: { value: 'no such person' } });
    expect(screen.getByText(i18next.t('team.list.noMembersMatchTitle'))).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: i18next.t('common.clearSearch') }));

    expect(screen.queryByText(i18next.t('team.list.noMembersMatchTitle'))).not.toBeInTheDocument();
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
  });

  it('does not show the empty state when members match', () => {
    teamData.data = {
      members: [
        {
          id: '1',
          name: 'Jane Doe',
          email: 'jane@rtylr.com',
          memberType: 'admin',
          licensedAppIds: [],
          permissionProfileAssignments: [],
          effectiveAccess: { apps: [] },
          status: 'active',
        },
      ],
      hasMore: false,
    };
    render(<Team />);

    expect(screen.queryByText(i18next.t('team.list.noMembersTitle'))).not.toBeInTheDocument();
  });
});
