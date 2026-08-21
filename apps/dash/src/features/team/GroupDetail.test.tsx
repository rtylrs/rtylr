import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
  },
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
  useParams: () => ({ id: '5' }),
}));

type GroupMember = { userId: number; fullName: string; email: string };
type Group = {
  id: number;
  name: string;
  description: string;
  color: string;
  memberCount: number;
  members: GroupMember[];
  resourcePermissions: { resourceKey: string; permissionLevel: string }[];
};
type TeamMember = { id: string; name: string; email: string; role: string };

const groupState = vi.hoisted(() => ({
  data: undefined as Group | undefined,
  isLoading: false,
}));

const teamState = vi.hoisted(() => ({
  data: { members: [] as TeamMember[] } as { members: TeamMember[] } | undefined,
}));

vi.mock('../../shared/hooks/queries/useGroups', () => ({
  useGroup: () => groupState,
  useCreateGroup: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateGroup: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useAddGroupMembers: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemoveGroupMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useTeam', () => ({
  useTeam: () => teamState,
}));

import GroupDetail from './GroupDetail';

const makeGroup = (overrides: Partial<Group>): Group => ({
  id: 5,
  name: 'Team A',
  description: '',
  color: '#4583C4',
  memberCount: 0,
  members: [],
  resourcePermissions: [],
  ...overrides,
});

beforeEach(() => {
  vi.clearAllMocks();
  groupState.data = makeGroup({});
  groupState.isLoading = false;
  teamState.data = { members: [] };
});

describe('GroupDetail: no-members empty state', () => {
  it('shows the illustration, title and an add-members call-to-action when the group has no members', () => {
    teamState.data = {
      members: [{ id: '1', name: 'Jane Doe', email: 'jane@rtylr.com', role: 'staff' }],
    };
    groupState.data = makeGroup({ members: [] });
    render(<GroupDetail />);

    fireEvent.click(screen.getByRole('button', { name: /team.groupDetail.tabMembers/ }));

    expect(screen.getByText('team.groupDetail.noMembersTitle')).toBeInTheDocument();
    expect(screen.getByText('team.groupDetail.noMembersDescription')).toBeInTheDocument();

    // The panel header also has an "add members" button; click the one inside the empty-state card.
    const addMembersButtons = screen.getAllByRole('button', {
      name: /team.groupDetail.addMembersButton/,
    });
    fireEvent.click(addMembersButtons[addMembersButtons.length - 1]);

    expect(screen.getByText('team.groupDetail.addMembersTitle')).toBeInTheDocument();
  });
});

describe('GroupDetail: all-members-in-group empty state', () => {
  it('shows the "everyone already added" state and closes the panel when its call-to-action is clicked', () => {
    teamState.data = {
      members: [{ id: '1', name: 'Jane Doe', email: 'jane@rtylr.com', role: 'staff' }],
    };
    groupState.data = makeGroup({
      members: [{ userId: 1, fullName: 'Jane Doe', email: 'jane@rtylr.com' }],
    });
    render(<GroupDetail />);

    fireEvent.click(screen.getByRole('button', { name: /team.groupDetail.tabMembers/ }));
    fireEvent.click(screen.getByRole('button', { name: /team.groupDetail.addMembersButton/ }));

    expect(screen.getByText('team.groupDetail.allMembersInGroup')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.close' }));

    expect(screen.queryByText('team.groupDetail.addMembersTitle')).not.toBeInTheDocument();
  });
});
