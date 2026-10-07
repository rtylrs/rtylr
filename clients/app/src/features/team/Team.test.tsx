import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';

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

describe('Team member filters', () => {
  const member = (id: string, name: string, memberType: string, status: string): TeamMember => ({
    id,
    name,
    email: `${id}@example.test`,
    memberType,
    status,
    licensedAppIds: [],
    permissionProfileAssignments: [],
    effectiveAccess: { apps: [] },
  });

  it('defaults to active members and composes type and status filters', () => {
    teamData.data = {
      members: [
        member('1', 'Taylor Owner', 'owner', 'active'),
        member('2', 'Alex Admin', 'admin', 'active'),
        member('3', 'Sam Member', 'member', 'active'),
        member('4', 'Casey Admin', 'admin', 'inactive'),
        member('5', 'Jordan Member', 'member', 'inactive'),
      ],
      hasMore: false,
    };
    render(<Team />);
    expect(screen.getByRole('button', { name: 'Taylor Owner' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Jordan Member' })).not.toBeInTheDocument();

    fireEvent.change(screen.getByRole('combobox', { name: 'Member type' }), {
      target: { value: 'admin' },
    });
    expect(screen.getByRole('button', { name: 'Alex Admin' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Sam Member' })).not.toBeInTheDocument();

    fireEvent.change(screen.getByRole('combobox', { name: 'Member status' }), {
      target: { value: 'inactive' },
    });
    expect(screen.getByRole('button', { name: 'Casey Admin' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Alex Admin' })).not.toBeInTheDocument();

    fireEvent.change(screen.getByRole('combobox', { name: 'Member type' }), {
      target: { value: 'all' },
    });
    expect(screen.getByRole('button', { name: 'Jordan Member' })).toBeInTheDocument();
  });

  it('can reveal inactive members from a filtered empty state without claiming the team is empty', () => {
    teamData.data = {
      members: [member('1', 'Jordan Member', 'member', 'inactive')],
      hasMore: false,
    };
    render(<Team />);
    expect(screen.queryByText('No team members yet')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Clear filters' }));
    expect(screen.getByRole('button', { name: 'Jordan Member' })).toBeInTheDocument();
  });

  it('renders the member actions menu outside the clipped member table', () => {
    teamData.data = {
      members: [member('1', 'Taylor Owner', 'admin', 'active')],
      hasMore: false,
    };
    render(<Team />);

    const memberTable = screen
      .getByRole('button', { name: 'Taylor Owner' })
      .closest('.rounded-xl.overflow-hidden');
    fireEvent.pointerDown(
      screen.getByRole('button', { name: i18next.t('team.list.memberActions') })
    );

    const menu = screen.getByRole('menu');
    expect(memberTable).not.toContainElement(menu);
    expect(
      screen.getByRole('menuitem', { name: i18next.t('team.list.editMember') })
    ).toBeInTheDocument();
  });
});
