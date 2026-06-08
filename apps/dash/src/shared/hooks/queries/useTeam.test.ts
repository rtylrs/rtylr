import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useTeam,
  useMember,
  useInviteMember,
  useUpdateMember,
  useDeactivateMember,
} from './useTeam';

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockPut = vi.fn();
const mockPatch = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
    patch: (...args: unknown[]) => mockPatch(...args),
  },
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useTeam', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches team members from .users key', async () => {
    mockGet.mockResolvedValue({
      users: [
        {
          id: '1',
          fullName: 'Alice',
          email: 'alice@test.com',
          role: 'admin',
          apps: [],
          status: 'active',
        },
      ],
    });

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.members).toHaveLength(1);
    expect(result.current.data?.members[0].name).toBe('Alice');
  });

  it('fetches team members from .members key', async () => {
    mockGet.mockResolvedValue({
      members: [
        { id: 2, fullName: 'Bob', email: 'bob@test.com', role: 'staff', apps: [], isActive: true },
      ],
    });

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.members).toHaveLength(1);
  });

  it('prefers fullName over name', async () => {
    mockGet.mockResolvedValue({
      users: [
        {
          id: '1',
          fullName: 'Full Name',
          name: 'Short',
          email: 'a@b.com',
          role: 'admin',
          apps: [],
          status: 'active',
        },
      ],
    });

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.members[0].name).toBe('Full Name');
  });

  it('falls back to email when no name', async () => {
    mockGet.mockResolvedValue({
      users: [{ id: '1', email: 'fallback@test.com', role: 'viewer', apps: [], status: 'active' }],
    });

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.members[0].name).toBe('fallback@test.com');
  });

  it('returns empty array for non-array response', async () => {
    mockGet.mockResolvedValue('invalid');

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.members).toEqual([]);
  });

  it('handles API error', async () => {
    mockGet.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single member by id', async () => {
    mockGet.mockResolvedValue({
      user: {
        id: '5',
        fullName: 'Charlie',
        email: 'c@c.com',
        role: 'staff',
        isActive: true,
        apps: ['pos'],
      },
    });

    const { result } = renderHook(() => useMember('5'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

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

  it('calls post with invite data', async () => {
    mockPost.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useInviteMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        email: 'new@test.com',
        fullName: 'New User',
        role: 'staff',
        apps: ['pos'],
      });
    });

    expect(mockPost).toHaveBeenCalledWith('/team/invite', {
      email: 'new@test.com',
      fullName: 'New User',
      role: 'staff',
      apps: ['pos'],
    });
  });
});

describe('useDeactivateMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deactivate endpoint', async () => {
    mockPatch.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useDeactivateMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync('42');
    });

    expect(mockPatch).toHaveBeenCalledWith('/team/members/42/deactivate');
  });
});

describe('useUpdateMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('sends profile and permission updates', async () => {
    mockPut.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useUpdateMember(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        id: '42',
        patch: { fullName: 'Edited User', role: 'admin', apps: ['pos', 'finance'] },
      });
    });

    expect(mockPut).toHaveBeenCalledWith('/team/users/permissions', {
      userId: 42,
      fullName: 'Edited User',
      role: 'admin',
      apps: ['pos', 'finance'],
    });
  });

  it('updates cached list and detail data after save', async () => {
    mockGet
      .mockResolvedValueOnce({
        users: [
          {
            id: '42',
            fullName: 'Old Name',
            email: 'member@test.com',
            role: 'staff',
            apps: ['pos'],
            isActive: true,
          },
        ],
      })
      .mockResolvedValueOnce({
        user: {
          id: '42',
          fullName: 'Old Name',
          email: 'member@test.com',
          role: 'staff',
          apps: ['pos'],
          isActive: true,
        },
      });
    mockPut.mockResolvedValue({ success: true });

    const wrapper = createWrapper();
    const team = renderHook(() => useTeam({ pageSize: 50 }), { wrapper });
    const member = renderHook(() => useMember('42'), { wrapper });

    await waitFor(() => expect(team.result.current.data?.members[0].name).toBe('Old Name'));
    await waitFor(() => expect(member.result.current.data?.name).toBe('Old Name'));

    const update = renderHook(() => useUpdateMember(), { wrapper });
    await act(async () => {
      await update.result.current.mutateAsync({
        id: '42',
        patch: { fullName: 'New Name', role: 'admin', apps: ['erp', 'flow'] },
      });
    });

    await waitFor(() =>
      expect(team.result.current.data?.members[0]).toMatchObject({
        fullName: 'New Name',
        name: 'New Name',
        role: 'admin',
        apps: ['erp', 'flow'],
      })
    );
    await waitFor(() =>
      expect(member.result.current.data).toMatchObject({
        fullName: 'New Name',
        name: 'New Name',
        role: 'admin',
        apps: ['erp', 'flow'],
      })
    );
  });
});
