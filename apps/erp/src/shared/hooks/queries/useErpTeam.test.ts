import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useErpTeam, useErpInviteMember, useErpDeactivateMember } from './useErpTeam';

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockPatch = vi.fn();

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    patch: (...args: unknown[]) => mockPatch(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleMembers = [
  {
    id: '1',
    name: 'Alice',
    email: 'alice@example.com',
    role: 'admin',
    apps: ['erp'],
    isActive: true,
  },
  { id: '2', name: 'Bob', email: 'bob@example.com', role: 'staff', apps: ['erp'], isActive: true },
];

describe('useErpTeam', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches team members', async () => {
    mockGet.mockResolvedValue({ members: sampleMembers });
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('falls back to users key if members not present', async () => {
    mockGet.mockResolvedValue({ users: sampleMembers });
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when response is empty', async () => {
    mockGet.mockResolvedValue({});
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockGet.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useErpInviteMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls invite endpoint with correct payload', async () => {
    mockPost.mockResolvedValue({ success: true });
    mockGet.mockResolvedValue({ members: [] });
    const { result } = renderHook(() => useErpInviteMember(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        email: 'new@example.com',
        fullName: 'New User',
        role: 'staff',
      });
    });
    expect(mockPost).toHaveBeenCalledWith('/team/invite', {
      email: 'new@example.com',
      fullName: 'New User',
      role: 'staff',
      apps: ['erp'],
    });
  });
});

describe('useErpDeactivateMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deactivate endpoint with member id', async () => {
    mockPatch.mockResolvedValue({ success: true });
    mockGet.mockResolvedValue({ members: [] });
    const { result } = renderHook(() => useErpDeactivateMember(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync('user-123');
    });
    expect(mockPatch).toHaveBeenCalledWith('/team/members/user-123/deactivate', {});
  });
});
