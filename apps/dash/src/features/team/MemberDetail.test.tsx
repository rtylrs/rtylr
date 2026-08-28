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
  useParams: () => ({ id: '1' }),
}));

type Member = {
  id: string;
  name: string;
  fullName: string;
  email: string;
  role: string;
  apps: string[];
  status: string;
};

const memberState = vi.hoisted(() => ({
  data: undefined as Member | undefined,
  isLoading: false,
  refetch: vi.fn(),
}));

const userState = vi.hoisted(() => ({
  user: { role: '', isOwner: false } as {
    role: string;
    isOwner: boolean;
    effectiveAccess?: { dashPermissions: string[] };
  },
}));

const groupsState = vi.hoisted(() => ({
  data: [] as { id: number; name: string; color: string }[],
}));

vi.mock('../../shared/hooks/queries/useTeam', () => ({
  useMember: () => memberState,
  useDeactivateMember: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateMember: () => ({ mutateAsync: vi.fn(), isPending: false }),
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

// Stable references: a mock returning a fresh [] each call changes the
// dependency identity on every render, and MemberDetail's
// `useEffect(..., [memberLocations])` would re-fire forever.
const EMPTY_LOCATIONS = vi.hoisted(() => [] as { id: number; locationId: number; name: string }[]);

vi.mock('../../shared/hooks/queries/usePOSAccess', () => ({
  usePOSLocations: () => ({ data: EMPTY_LOCATIONS }),
  useMemberLocations: () => ({ data: EMPTY_LOCATIONS }),
  useAssignPOSLocation: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemovePOSLocation: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: typeof userState) => unknown) => selector(userState),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: () => ({ data: groupsState.data }),
  useMutation: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}));

import MemberDetail from './MemberDetail';

const makeMember = (overrides: Partial<Member>): Member => ({
  id: '1',
  name: 'Jane Doe',
  fullName: 'Jane Doe',
  email: 'jane@rtylr.com',
  role: 'staff',
  apps: [],
  status: 'active',
  ...overrides,
});

beforeEach(() => {
  vi.clearAllMocks();
  memberState.data = undefined;
  memberState.isLoading = false;
  userState.user = {
    role: 'owner',
    isOwner: true,
    effectiveAccess: { dashPermissions: ['hub:team:edit'] },
  };
  groupsState.data = [];
});

describe('MemberDetail: member-not-found empty state', () => {
  it('shows the illustration, title and a back-to-team call-to-action when the member is missing', () => {
    render(<MemberDetail />);

    expect(screen.getByText('team.memberDetail.notFoundTitle')).toBeInTheDocument();
    expect(screen.getByText('team.memberDetail.notFoundDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'team.memberDetail.backToTeam' }));

    expect(navigateMock).toHaveBeenCalledWith({ to: '/team' });
  });
});

describe('MemberDetail: app-access empty state (permission-gated)', () => {
  it('shows an "assign app" action that opens edit mode when the current user can edit the member', () => {
    userState.user = {
      role: 'owner',
      isOwner: true,
      effectiveAccess: { dashPermissions: ['hub:team:edit'] },
    };
    memberState.data = makeMember({ apps: [] });
    render(<MemberDetail />);

    expect(screen.getByText('team.memberDetail.noAppsAssignedTitle')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'team.memberDetail.assignApp' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'common.refresh' })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'team.memberDetail.assignApp' }));

    // Entering edit mode swaps the header "Edit" button for "Cancel"/"Save".
    expect(screen.getByRole('button', { name: /Save/ })).toBeInTheDocument();
  });

  it('shows a refresh fallback instead when the current user cannot edit the member', () => {
    userState.user = { role: 'staff', isOwner: false, effectiveAccess: { dashPermissions: [] } };
    memberState.data = makeMember({ apps: [] });
    render(<MemberDetail />);

    expect(screen.getByText('team.memberDetail.noAppsAssignedTitle')).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'team.memberDetail.assignApp' })
    ).not.toBeInTheDocument();
    const refreshButton = screen.getByRole('button', { name: 'common.refresh' });
    expect(refreshButton).toBeInTheDocument();

    fireEvent.click(refreshButton);

    expect(memberState.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when the member has assigned apps', () => {
    memberState.data = makeMember({ apps: ['erp'] });
    render(<MemberDetail />);

    expect(screen.queryByText('team.memberDetail.noAppsAssignedTitle')).not.toBeInTheDocument();
  });
});

describe('MemberDetail: groups empty state', () => {
  it('shows the illustration, title and an add-to-group call-to-action when the member is in no groups', () => {
    groupsState.data = [];
    memberState.data = makeMember({});
    render(<MemberDetail />);

    expect(screen.getByText('team.memberDetail.notInGroupTitle')).toBeInTheDocument();
    expect(screen.getByText('team.memberDetail.notInGroupDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'team.memberDetail.addToGroup' }));

    expect(navigateMock).toHaveBeenCalledWith({ to: '/team/groups' });
  });

  it('does not show the empty state when the member belongs to a group', () => {
    groupsState.data = [{ id: 1, name: 'Managers', color: '#4583C4' }];
    memberState.data = makeMember({});
    render(<MemberDetail />);

    expect(screen.queryByText('team.memberDetail.notInGroupTitle')).not.toBeInTheDocument();
    expect(screen.getByText('Managers')).toBeInTheDocument();
  });
});
