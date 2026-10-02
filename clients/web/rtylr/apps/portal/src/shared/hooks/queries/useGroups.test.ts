import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { decodeGroup } from '@voxire/voxie/hub/team/groups/v1/types';

const mocks = vi.hoisted(() => ({
  groupsClient: {
    listGroups: vi.fn(),
    removeGroupMember: vi.fn(),
  },
  invitationsClient: {
    bulkInvite: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: vi.fn(() => mocks.invitationsClient),
  getOneKitErrorMessage: vi.fn((_error: unknown, fallback: string) => fallback),
}));

vi.mock('../../lib/clients', () => ({
  groupsClient: mocks.groupsClient,
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { useGroups, useRemoveGroupMember } from './useGroups';

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const wrapper = ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
  return { queryClient, wrapper };
}

describe('useGroups', () => {
  it('loads group assignments when unscoped grants omit the empty scopes list', async () => {
    mocks.groupsClient.listGroups.mockResolvedValue({
      groups: [
        decodeGroup({
          id: '7',
          name: 'Readers',
          permission_profile_assignments: [
            {
              app_id: 'flow',
              permission_profile_ids: ['23'],
              assignment_source: 'group',
              status: 'active',
            },
          ],
        }),
      ],
    });
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useGroups(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0].permissionProfileAssignments).toEqual([
      {
        appId: 'flow',
        permissionProfileIds: ['23'],
        assignmentSource: 'group',
        status: 'active',
        scopes: [],
      },
    ]);
  });
});

describe('useRemoveGroupMember', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.groupsClient.removeGroupMember.mockResolvedValue({});
  });

  it('invalidates the group list as well as the group detail and team', async () => {
    const { queryClient, wrapper } = createWrapper();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useRemoveGroupMember('7'), { wrapper });

    await act(async () => {
      await result.current.mutateAsync(42);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mocks.groupsClient.removeGroupMember).toHaveBeenCalledWith({
      id: '7',
      userId: '42',
    });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['groups'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['team'] });
  });
});
