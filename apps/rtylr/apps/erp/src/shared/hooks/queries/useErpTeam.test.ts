import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useErpTeam, useErpInviteMember, useErpDeactivateMember } from './useErpTeam';

const { mockClient } = vi.hoisted(() => ({
  mockClient: {
    listTeam: vi.fn(),
    inviteUser: vi.fn(),
    removeMember: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: vi.fn(() => mockClient),
  getOneKitErrorMessage: vi.fn((_error, fallback) => fallback),
  oneKitResult: vi.fn((operation: () => Promise<unknown>) => operation()),
  unwrap: vi.fn(async (value: unknown) => value),
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleMembers = [
  {
    id: '1',
    fullName: 'Alice',
    email: 'alice@example.com',
    role: 'admin',
    apps: ['erp'],
    isActive: true,
  },
  {
    id: '2',
    fullName: 'Bob',
    email: 'bob@example.com',
    role: 'staff',
    apps: ['erp'],
    isActive: true,
  },
];

describe('useErpTeam', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches and maps canonical team members', async () => {
    mockClient.listTeam.mockResolvedValue({ users: sampleMembers });
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
    expect(result.current.data?.[0]?.name).toBe('Alice');
  });

  it('returns empty array when response is empty', async () => {
    mockClient.listTeam.mockResolvedValue({});
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockClient.listTeam.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useErpInviteMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls invite endpoint with correct payload', async () => {
    mockClient.inviteUser.mockResolvedValue({ success: true });
    mockClient.listTeam.mockResolvedValue({ users: [] });
    const { result } = renderHook(() => useErpInviteMember(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        email: 'new@example.com',
        fullName: 'New User',
        role: 'staff',
      });
    });
    expect(mockClient.inviteUser).toHaveBeenCalledWith({
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
    mockClient.removeMember.mockResolvedValue({ success: true });
    mockClient.listTeam.mockResolvedValue({ users: [] });
    const { result } = renderHook(() => useErpDeactivateMember(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync('user-123');
    });
    expect(mockClient.removeMember).toHaveBeenCalledWith({ userId: 'user-123' });
  });
});
