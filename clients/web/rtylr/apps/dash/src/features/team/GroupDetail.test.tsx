import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
  },
}));

const navigateMock = vi.hoisted(() => vi.fn());
const routeParams = vi.hoisted(() => ({ id: '5' as string | undefined }));
const createGroupMock = vi.hoisted(() => vi.fn());
const updateGroupMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
  useParams: () => routeParams,
}));

type GroupMember = { userId: number; fullName: string; email: string };
type Group = {
  id: number;
  name: string;
  description: string;
  color: string;
  memberCount: number;
  members: GroupMember[];
  licensedAppIds: string[];
  permissionProfileAssignments: { appId: string; permissionProfileIds: string[] }[];
};
type TeamMember = { id: string; name: string; email: string; role: string };

const groupState = vi.hoisted(() => ({
  data: undefined as Group | undefined,
  isLoading: false,
}));

const teamRequest = vi.hoisted(() => ({ page: 1, pageSize: 50 }));
const teamState = vi.hoisted(() => ({
  data: {
    members: [] as TeamMember[],
    hasMore: false,
    total: 0,
    page: 1,
    pageSize: 50,
  } as
    | {
        members: TeamMember[];
        hasMore: boolean;
        total: number;
        page: number;
        pageSize: number;
      }
    | undefined,
  isLoading: false,
  isError: false,
  refetch: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useGroups', () => ({
  useGroup: () => groupState,
  useCreateGroup: () => ({ mutateAsync: createGroupMock, isPending: false }),
  useUpdateGroup: () => ({ mutateAsync: updateGroupMock, isPending: false }),
  useAddGroupMembers: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemoveGroupMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/usePermissions', () => ({
  usePermissionProfiles: () => ({ data: [] }),
}));

vi.mock('../../shared/hooks/queries/useTeam', () => ({
  useTeam: (params: { page?: number; pageSize?: number }) => {
    teamRequest.page = params.page ?? 1;
    teamRequest.pageSize = params.pageSize ?? 50;
    return teamState;
  },
}));

import GroupDetail from './GroupDetail';

const makeGroup = (overrides: Partial<Group>): Group => ({
  id: 5,
  name: 'Team A',
  description: '',
  color: '#4583C4',
  memberCount: 0,
  members: [],
  licensedAppIds: [],
  permissionProfileAssignments: [],
  ...overrides,
});

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(async () => {
  vi.clearAllMocks();
  await i18next.changeLanguage('en');
  routeParams.id = '5';
  createGroupMock.mockResolvedValue({ id: 6 });
  groupState.data = makeGroup({});
  groupState.isLoading = false;
  teamState.data = {
    members: [],
    hasMore: false,
    total: 0,
    page: 1,
    pageSize: 50,
  };
  teamState.isLoading = false;
  teamState.isError = false;
  teamRequest.page = 1;
  teamRequest.pageSize = 50;
});

describe('GroupDetail: no-members empty state', () => {
  it('shows the illustration, title and an add-members call-to-action when the group has no members', () => {
    teamState.data = {
      members: [{ id: '1', name: 'Jane Doe', email: 'jane@rtylr.com', role: 'staff' }],
      hasMore: false,
      total: 1,
      page: 1,
      pageSize: 50,
    };
    groupState.data = makeGroup({ members: [] });
    render(<GroupDetail />);

    fireEvent.click(screen.getByRole('button', { name: /Members \(/ }));

    expect(screen.getByText('No members yet')).toBeInTheDocument();
    expect(screen.getByText('Add team members to this group.')).toBeInTheDocument();

    // The panel header also has an "add members" button; click the one inside the empty-state card.
    const addMembersButtons = screen.getAllByRole('button', {
      name: 'Add members',
    });
    fireEvent.click(addMembersButtons[addMembersButtons.length - 1]);

    expect(screen.getByText('Add members', { selector: 'p' })).toBeInTheDocument();
  });
});

describe('GroupDetail: candidate pagination', () => {
  it('keeps the next candidate page reachable when the first page is full', () => {
    teamState.data = {
      members: Array.from({ length: 50 }, (_, index) => ({
        id: String(index + 1),
        name: `Member ${index + 1}`,
        email: `member${index + 1}@rtylr.com`,
        role: 'staff',
      })),
      hasMore: true,
      total: 51,
      page: 1,
      pageSize: 50,
    };
    groupState.data = makeGroup({
      members: Array.from({ length: 50 }, (_, index) => ({
        userId: index + 1,
        fullName: `Member ${index + 1}`,
        email: `member${index + 1}@rtylr.com`,
      })),
    });

    render(<GroupDetail />);
    fireEvent.click(screen.getByRole('button', { name: /Members \(/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Add members' }));

    const next = screen.getByRole('button', { name: 'Next' });
    expect(next).toBeEnabled();
    fireEvent.click(next);

    expect(teamRequest.page).toBe(2);
  });
});

describe('GroupDetail: all-members-in-group empty state', () => {
  it('shows the "everyone already added" state and closes the panel when its call-to-action is clicked', () => {
    teamState.data = {
      members: [{ id: '1', name: 'Jane Doe', email: 'jane@rtylr.com', role: 'staff' }],
      hasMore: false,
      total: 1,
      page: 1,
      pageSize: 50,
    };
    groupState.data = makeGroup({
      members: [{ userId: 1, fullName: 'Jane Doe', email: 'jane@rtylr.com' }],
    });
    render(<GroupDetail />);

    fireEvent.click(screen.getByRole('button', { name: /Members \(/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Add members' }));

    expect(screen.getByText('All team members are already in this group.')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Close' }));

    expect(screen.queryByText('Add members', { selector: 'p' })).not.toBeInTheDocument();
  });
});

describe('GroupDetail: static creation route', () => {
  it('creates a group when the new route has no id parameter', async () => {
    routeParams.id = undefined;
    groupState.data = undefined;
    render(<GroupDetail />);

    expect(screen.getByRole('heading', { name: 'New group' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Members \(/ })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
    fireEvent.change(screen.getByPlaceholderText('e.g. Store Managers'), {
      target: { value: 'Operations' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() =>
      expect(createGroupMock).toHaveBeenCalledWith({
        name: 'Operations',
        description: '',
        color: '#6366f1',
        licensedAppIds: [],
        permissionProfileAssignments: [],
      })
    );
    expect(updateGroupMock).not.toHaveBeenCalled();
    expect(navigateMock).toHaveBeenCalledWith({ to: '/team/groups/6', replace: true });
  });
});

describe('GroupDetail: translated member counts', () => {
  it.each([
    ['en', 1, '1 member'],
    ['en', 2, '2 members'],
    ['fr', 1, '1 membre'],
    ['fr', 2, '2 membres'],
    ['ar', 1, '1 عضو'],
    ['ar', 2, '2 أعضاء'],
  ])('renders the %s member count for %i members', async (language, count, label) => {
    await initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales'), String(language));
    const members = Array.from({ length: Number(count) }, (_, index) => ({
      userId: index + 1,
      fullName: `Member ${index + 1}`,
      email: `member${index + 1}@example.test`,
    }));
    groupState.data = makeGroup({ memberCount: Number(count), members });
    render(<GroupDetail />);

    expect(screen.getByText(String(label))).toBeInTheDocument();
    fireEvent.click(
      screen.getByRole('button', {
        name: i18next.t('team.groupDetail.tabMembers', { count }),
      })
    );
    expect(screen.getAllByText(String(label))).toHaveLength(2);
  });
});
