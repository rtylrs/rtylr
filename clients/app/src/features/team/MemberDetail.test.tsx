import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
  // MemberDetail's deactivate dialog renders SeatSavingsNote, which uses Trans.
  // A mock supplying only useTranslation throws when that dialog opens.
  Trans: ({ i18nKey }: { i18nKey: string }) => i18nKey,
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
  memberType: string;
  licensedAppIds: string[];
  permissionProfileAssignments: {
    appId: string;
    permissionProfileIds: string[];
    assignmentSource: string;
    status: string;
    groupId?: string;
    scopes: { scopeType: string; scopeId: string }[];
  }[];
  status: string;
  effectiveAccess?: {
    apps: {
      appId: string;
      licensed: boolean;
      hasAccess: boolean;
      permissions: string[];
      permissionScopes: Record<string, { scopeType: string; scopeId: string }[]>;
      sources: string[];
    }[];
  };
  isOwner?: boolean;
  groups?: { id: number; name: string; color: string }[];
};

const memberState = vi.hoisted(() => ({
  data: undefined as Member | undefined,
  isLoading: false,
  refetch: vi.fn(),
}));

// Counts MemberDetail renders across the test; the disabled-locations bug
// re-rendered until React threw "Maximum update depth exceeded".
const renderState = vi.hoisted(() => ({ memberDetailRenders: 0 }));
const updateMemberMock = vi.hoisted(() => vi.fn());

const userState = vi.hoisted(() => ({
  user: { isOwner: false } as {
    isOwner: boolean;
    effectiveAccess?: { dashPermissions: string[] };
  },
}));

const groupsState = vi.hoisted(() => ({
  data: [] as { id: number; name: string; color: string }[],
}));

vi.mock('../../shared/hooks/queries/useTeam', () => ({
  useMember: () => {
    renderState.memberDetailRenders += 1;
    // A render loop would otherwise spin until Vitest kills the worker.
    if (renderState.memberDetailRenders > 25) {
      throw new Error('MemberDetail render loop detected');
    }
    return memberState;
  },
  useDeactivateMember: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateMember: () => ({ mutateAsync: updateMemberMock, isPending: false }),
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

// Stable identities for the same reason: these feed a
// useEffect dependency. `undefined` models the hook while its query is
// disabled and no cached data exists.
const posState = vi.hoisted(() => ({
  locations: [] as { id: number; name: string }[],
  memberLocations: undefined as
    | {
        id: number;
        locationId: number;
        canManagePosUsers: boolean;
        canManageTerminals: boolean;
        canViewReports: boolean;
      }[]
    | undefined,
}));
const assignLocationMock = vi.hoisted(() => vi.fn());
const removeLocationMock = vi.hoisted(() => vi.fn());

vi.mock('../../shared/hooks/queries/usePOSAccess', () => ({
  usePOSLocations: () => ({ data: posState.locations }),
  useMemberLocations: () => ({ data: posState.memberLocations }),
  useAssignPOSLocation: () => ({ mutateAsync: assignLocationMock, isPending: false }),
  useRemovePOSLocation: () => ({ mutateAsync: removeLocationMock, isPending: false }),
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
  memberType: 'member',
  licensedAppIds: [],
  permissionProfileAssignments: [],
  effectiveAccess: { apps: [] },
  status: 'active',
  groups: [],
  ...overrides,
});

beforeEach(() => {
  vi.clearAllMocks();
  memberState.data = undefined;
  memberState.isLoading = false;
  renderState.memberDetailRenders = 0;
  updateMemberMock.mockReset();
  updateMemberMock.mockResolvedValue({});
  posState.locations = [];
  posState.memberLocations = undefined;
  userState.user = {
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
      isOwner: true,
      effectiveAccess: { dashPermissions: ['hub:team:edit'] },
    };
    memberState.data = makeMember({ licensedAppIds: [] });
    render(<MemberDetail />);

    expect(screen.getByText('team.memberDetail.noAppsAssignedTitle')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'team.memberDetail.assignApp' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'common.refresh' })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'team.memberDetail.assignApp' }));

    // Entering edit mode swaps the header "Edit" button for "Cancel"/"Save".
    expect(screen.getByRole('button', { name: 'common.save' })).toBeInTheDocument();
  });

  it('shows a refresh fallback instead when the current user cannot edit the member', () => {
    userState.user = { isOwner: false, effectiveAccess: { dashPermissions: [] } };
    memberState.data = makeMember({ licensedAppIds: [] });
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
    memberState.data = makeMember({
      licensedAppIds: ['erp'],
      effectiveAccess: {
        apps: [
          {
            appId: 'erp',
            licensed: true,
            hasAccess: true,
            permissions: [],
            permissionScopes: {},
            sources: ['explicit'],
          },
        ],
      },
    });
    render(<MemberDetail />);

    expect(screen.queryByText('team.memberDetail.noAppsAssignedTitle')).not.toBeInTheDocument();
  });
});

describe('MemberDetail: group access stays inherited while editing direct assignments', () => {
  it('does not convert a group-only app assignment into a direct assignment on save', async () => {
    groupsState.data = [{ id: 12, name: 'Flow users', color: '#64C5B9' }];
    memberState.data = makeMember({
      licensedAppIds: [],
      permissionProfileAssignments: [
        {
          appId: 'flow',
          permissionProfileIds: ['73'],
          assignmentSource: 'group',
          status: 'active',
          groupId: '12',
          scopes: [],
        },
      ],
      effectiveAccess: {
        apps: [
          {
            appId: 'flow',
            licensed: true,
            hasAccess: true,
            permissions: ['flow:projects:view'],
            permissionScopes: {},
            sources: ['group'],
          },
        ],
      },
    });

    render(<MemberDetail />);
    fireEvent.click(screen.getByRole('button', { name: 'common.edit' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.save' }));

    await waitFor(() => expect(updateMemberMock).toHaveBeenCalledTimes(1));
    expect(updateMemberMock).toHaveBeenCalledWith({
      id: '1',
      patch: expect.objectContaining({
        licensedAppIds: [],
        permissionProfileAssignments: [],
        groupIds: [12],
      }),
    });
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

describe('MemberDetail: POS location capabilities (#1028)', () => {
  const KEPT = {
    id: 501,
    locationId: 9001,
    canManagePosUsers: true,
    canManageTerminals: true,
    canViewReports: true,
  };

  const saveWithPosLocations = async (): Promise<void> => {
    posState.locations = [{ id: 9001, name: 'Downtown' }];
    posState.memberLocations = [KEPT];
    memberState.data = makeMember({ licensedAppIds: ['pos'] });

    render(<MemberDetail />);
    fireEvent.click(screen.getByRole('button', { name: 'common.edit' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.save' }));
    await waitFor(() => expect(assignLocationMock).toHaveBeenCalled());
  };

  it('carries the existing capabilities across a save', async () => {
    await saveWithPosLocations();

    // These were hardcoded to false, so saving any member with POS locations
    // revoked every capability on every location they held, whatever their
    // role, with no sign of it in the UI.
    expect(assignLocationMock).toHaveBeenCalledWith({
      userId: 1,
      locationId: 9001,
      canManagePosUsers: true,
      canManageTerminals: true,
      canViewReports: true,
    });
  });

  it('does not revoke a capability the member already held', async () => {
    await saveWithPosLocations();

    const sent = assignLocationMock.mock.calls[0][0] as Record<string, unknown>;
    expect(sent.canManagePosUsers).not.toBe(false);
    expect(sent.canManageTerminals).not.toBe(false);
    expect(sent.canViewReports).not.toBe(false);
  });

  it('preserves a partially granted assignment exactly', async () => {
    posState.locations = [{ id: 9001, name: 'Downtown' }];
    posState.memberLocations = [
      { ...KEPT, canManagePosUsers: false, canManageTerminals: true, canViewReports: false },
    ];
    memberState.data = makeMember({ licensedAppIds: ['pos'] });

    render(<MemberDetail />);
    fireEvent.click(screen.getByRole('button', { name: 'common.edit' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.save' }));

    await waitFor(() => expect(assignLocationMock).toHaveBeenCalled());
    expect(assignLocationMock).toHaveBeenCalledWith(
      expect.objectContaining({
        canManagePosUsers: false,
        canManageTerminals: true,
        canViewReports: false,
      })
    );
  });
});

describe('MemberDetail: member-locations query with no data', () => {
  it('renders a bounded number of times when the query is disabled and empty', () => {
    posState.memberLocations = undefined;
    memberState.data = makeMember({ licensedAppIds: ['pos'] });

    render(<MemberDetail />);

    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    // Pre-fix, the `= []` destructuring default gave the seeding effect a new
    // dependency on every render, so the page looped until React threw
    // "Maximum update depth exceeded" and never settled.
    expect(renderState.memberDetailRenders).toBeLessThan(10);
  });

  it('keeps location edits made before a background refetch', async () => {
    posState.locations = [
      { id: 9001, name: 'Downtown' },
      { id: 9002, name: 'Uptown' },
    ];
    const assigned = {
      id: 501,
      locationId: 9001,
      canManagePosUsers: false,
      canManageTerminals: false,
      canViewReports: false,
    };
    posState.memberLocations = [assigned];
    memberState.data = makeMember({ licensedAppIds: ['pos'] });

    const { rerender } = render(<MemberDetail />);
    fireEvent.click(screen.getByRole('button', { name: 'common.edit' }));
    fireEvent.click(screen.getByRole('button', { name: 'Uptown' }));

    // A refetch resolves with structurally equal data under a new identity.
    posState.memberLocations = [{ ...assigned }];
    rerender(<MemberDetail />);

    fireEvent.click(screen.getByRole('button', { name: 'common.save' }));

    await waitFor(() =>
      expect(assignLocationMock).toHaveBeenCalledWith(expect.objectContaining({ locationId: 9002 }))
    );
  });
});
