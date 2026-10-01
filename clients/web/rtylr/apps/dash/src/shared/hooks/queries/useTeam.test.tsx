import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockMembersClient = vi.hoisted(() => ({
  listTeam: vi.fn(),
  getMember: vi.fn(),
  removeMember: vi.fn(),
}));
const mockPermissionsClient = vi.hoisted(() => ({
  updatePermissions: vi.fn(),
}));
const mockInvitationsClient = vi.hoisted(() => ({
  inviteUser: vi.fn(),
  bulkInvite: vi.fn(),
}));
const mockAccessClient = vi.hoisted(() => ({
  listAccessRequests: vi.fn(),
  approveAccessRequest: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (Client: { name: string }) => {
    switch (Client.name) {
      case 'MembersServiceClient':
        return mockMembersClient;
      case 'PermissionsServiceClient':
        return mockPermissionsClient;
      case 'InvitationsServiceClient':
        return mockInvitationsClient;
      case 'AccessServiceClient':
        return mockAccessClient;
      default:
        throw new Error(`unexpected client ${Client.name}`);
    }
  },
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import {
  useTeam,
  useMember,
  useInviteMember,
  useUpdateMember,
  useDeactivateMember,
} from './useTeam';

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useTeam', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches team members', async () => {
    mockMembersClient.listTeam.mockResolvedValue({
      users: [
        {
          id: '1',
          fullName: 'Alice',
          email: 'alice@test.com',
          memberType: 'admin',
          licensedAppIds: [],
          permissionProfileAssignments: [],
          isActive: true,
          createdAt: '',
        },
      ],
    });

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.members).toHaveLength(1);
    expect(result.current.data?.members[0].name).toBe('Alice');
  });

  it('exposes the authoritative total and continuation flag for a page', async () => {
    mockMembersClient.listTeam.mockResolvedValue({
      users: [
        {
          id: '51',
          fullName: 'Fifty-first member',
          email: 'member51@test.com',
          memberType: 'member',
          licensedAppIds: [],
          permissionProfileAssignments: [],
          isActive: true,
          createdAt: '',
        },
      ],
      total: 51,
    });

    const { result } = renderHook(() => useTeam({ page: 1, pageSize: 50 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toMatchObject({ total: 51, hasMore: true, page: 1, pageSize: 50 });
  });

  it('does not infer another page when the response omits total', async () => {
    mockMembersClient.listTeam.mockResolvedValue({
      users: [],
    });

    const { result } = renderHook(() => useTeam({ page: 2, pageSize: 50 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toMatchObject({ total: 0, hasMore: false, page: 2 });
  });

  it('falls back to email when no full name', async () => {
    mockMembersClient.listTeam.mockResolvedValue({
      users: [
        {
          id: '1',
          fullName: '',
          email: 'fallback@test.com',
          memberType: 'member',
          licensedAppIds: [],
          permissionProfileAssignments: [],
          isActive: true,
          createdAt: '',
        },
      ],
    });

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.members[0].name).toBe('fallback@test.com');
  });

  it('handles API error', async () => {
    mockMembersClient.listTeam.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single member by id', async () => {
    mockMembersClient.getMember.mockResolvedValue({
      user: {
        id: '5',
        fullName: 'Charlie',
        email: 'c@c.com',
        memberType: 'member',
        isActive: true,
        licensedAppIds: ['pos'],
        permissionProfileAssignments: [],
        createdAt: '',
        groups: [],
      },
    });

    const { result } = renderHook(() => useMember('5'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockMembersClient.getMember).toHaveBeenCalledWith({ userId: '5' });
    expect(result.current.data?.name).toBe('Charlie');
    expect(result.current.data?.status).toBe('active');
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useMember(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useInviteMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls inviteUser with invite data', async () => {
    mockInvitationsClient.inviteUser.mockResolvedValue({ userId: '9', message: 'ok' });

    const { result } = renderHook(() => useInviteMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        email: 'new@test.com',
        fullName: 'New User',
        memberType: 'member',
        licensedAppIds: ['pos'],
        permissionProfileAssignments: [],
      });
    });

    expect(mockInvitationsClient.inviteUser).toHaveBeenCalledWith({
      email: 'new@test.com',
      fullName: 'New User',
      memberType: 'member',
      licensedAppIds: ['pos'],
      permissionProfileAssignments: [],
    });
  });

  it('deduplicates group assignments', async () => {
    mockInvitationsClient.inviteUser.mockResolvedValue({ userId: '9', message: 'ok' });

    const { result } = renderHook(() => useInviteMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        email: 'both@test.com',
        fullName: 'Both Inputs',
        memberType: 'member',
        licensedAppIds: ['pos'],
        permissionProfileAssignments: [],
        groupIds: [11, 12, 12],
      });
    });

    const sent = mockInvitationsClient.inviteUser.mock.calls[0][0] as { groupIds: string[] };
    expect(sent.groupIds).toEqual(['11', '12']);
  });

  it('does not send duplicate group IDs', async () => {
    mockInvitationsClient.inviteUser.mockResolvedValue({ userId: '9', message: 'ok' });

    const { result } = renderHook(() => useInviteMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        email: 'dupe@test.com',
        fullName: 'Duplicate',
        memberType: 'member',
        licensedAppIds: ['pos'],
        permissionProfileAssignments: [],
        groupIds: [5, 5],
      });
    });

    const sent = mockInvitationsClient.inviteUser.mock.calls[0][0] as { groupIds: string[] };
    expect(sent.groupIds).toEqual(['5']);
  });

  it('sends no group list when neither input is given', async () => {
    mockInvitationsClient.inviteUser.mockResolvedValue({ userId: '9', message: 'ok' });

    const { result } = renderHook(() => useInviteMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        email: 'nogroup@test.com',
        fullName: 'No Group',
        memberType: 'member',
        licensedAppIds: ['pos'],
        permissionProfileAssignments: [],
      });
    });

    expect(mockInvitationsClient.inviteUser.mock.calls[0][0]).not.toHaveProperty('groupIds');
  });

  it('includes a temporary password when an administrator provisions credentials directly', async () => {
    mockInvitationsClient.inviteUser.mockResolvedValue({ userId: '9', message: 'ok' });

    const { result } = renderHook(() => useInviteMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        email: 'manual@test.com',
        fullName: 'Manual User',
        memberType: 'admin',
        licensedAppIds: [],
        permissionProfileAssignments: [],
        username: 'manual-user',
        temporaryPassword: 'strong-password',
      });
    });

    expect(mockInvitationsClient.inviteUser).toHaveBeenCalledWith({
      email: 'manual@test.com',
      fullName: 'Manual User',
      memberType: 'admin',
      licensedAppIds: [],
      permissionProfileAssignments: [],
      username: 'manual-user',
      temporaryPassword: 'strong-password',
    });
  });
});

describe('useDeactivateMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls removeMember with the member id', async () => {
    mockMembersClient.removeMember.mockResolvedValue({ message: 'ok' });

    const { result } = renderHook(() => useDeactivateMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync('42');
    });

    expect(mockMembersClient.removeMember).toHaveBeenCalledWith({ userId: '42' });
  });
});

describe('useUpdateMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('sends permission updates', async () => {
    mockPermissionsClient.updatePermissions.mockResolvedValue({});

    const { result } = renderHook(() => useUpdateMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        id: '42',
        patch: {
          fullName: 'Edited User',
          memberType: 'admin',
          licensedAppIds: ['pos', 'finance'],
          permissionProfileAssignments: [],
        },
      });
    });

    expect(mockPermissionsClient.updatePermissions).toHaveBeenCalledWith({
      userId: '42',
      fullName: 'Edited User',
      memberType: 'admin',
      replaceLicensedAppIds: true,
      licensedAppIds: ['pos', 'finance'],
      replacePermissionProfileAssignments: true,
      permissionProfileAssignments: [],
      replaceGroupIds: false,
    });
  });

  it('does not turn an unrelated profile update into an app and role reset', async () => {
    mockPermissionsClient.updatePermissions.mockResolvedValue({});

    const { result } = renderHook(() => useUpdateMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ id: '42', patch: { fullName: 'Edited User' } });
    });

    expect(mockPermissionsClient.updatePermissions).toHaveBeenCalledWith({
      userId: '42',
      fullName: 'Edited User',
      replaceLicensedAppIds: false,
      replacePermissionProfileAssignments: false,
      replaceGroupIds: false,
    });
  });

  it('updates cached list and detail data after save', async () => {
    mockMembersClient.listTeam.mockResolvedValueOnce({
      users: [
        {
          id: '42',
          fullName: 'Old Name',
          email: 'member@test.com',
          memberType: 'member',
          licensedAppIds: ['pos'],
          permissionProfileAssignments: [],
          isActive: true,
          createdAt: '',
        },
      ],
    });
    mockMembersClient.getMember.mockResolvedValueOnce({
      user: {
        id: '42',
        fullName: 'Old Name',
        email: 'member@test.com',
        memberType: 'member',
        licensedAppIds: ['pos'],
        permissionProfileAssignments: [],
        isActive: true,
        createdAt: '',
        groups: [],
      },
    });
    mockPermissionsClient.updatePermissions.mockResolvedValue({});

    const wrapper = createWrapper();
    const team = renderHook(() => useTeam({ pageSize: 50 }), { wrapper });
    const member = renderHook(() => useMember('42'), { wrapper });

    await waitFor(() => expect(team.result.current.data?.members[0].name).toBe('Old Name'));
    await waitFor(() => expect(member.result.current.data?.name).toBe('Old Name'));

    const update = renderHook(() => useUpdateMember(), { wrapper });
    await act(async () => {
      await update.result.current.mutateAsync({
        id: '42',
        patch: {
          fullName: 'New Name',
          memberType: 'admin',
          licensedAppIds: ['erp', 'flow'],
          permissionProfileAssignments: [],
        },
      });
    });

    expect(mockPermissionsClient.updatePermissions).toHaveBeenCalledWith({
      userId: '42',
      fullName: 'New Name',
      memberType: 'admin',
      replaceLicensedAppIds: true,
      licensedAppIds: ['erp', 'flow'],
      replacePermissionProfileAssignments: true,
      permissionProfileAssignments: [],
      replaceGroupIds: false,
    });

    await waitFor(() =>
      expect(team.result.current.data?.members[0]).toMatchObject({
        fullName: 'New Name',
        name: 'New Name',
        memberType: 'admin',
        licensedAppIds: ['erp', 'flow'],
      })
    );
    await waitFor(() =>
      expect(member.result.current.data).toMatchObject({
        fullName: 'New Name',
        name: 'New Name',
        memberType: 'admin',
        licensedAppIds: ['erp', 'flow'],
      })
    );
  });
});
