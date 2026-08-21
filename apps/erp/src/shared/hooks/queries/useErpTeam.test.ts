import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useErpTeam, useErpInviteMember, useErpDeactivateMember } from './useErpTeam';

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
  oneKitResult: (operation: () => Promise<unknown>) => operation(),
}));

const mockTeamClient = vi.hoisted(() => ({
  listMembers: vi.fn(),
  inviteMember: vi.fn(),
  deactivateMember: vi.fn(),
}));

vi.mock('../../lib/clients', () => ({ teamCompatibilityClient: mockTeamClient }));

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
  { id: '2', name: 'Bob', email: 'bob@example.com', role: 'member', apps: ['erp'], isActive: true },
];

describe('useErpTeam', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches team members', async () => {
    mockTeamClient.listMembers.mockResolvedValue({ members: sampleMembers });
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('falls back to users key if members not present', async () => {
    mockTeamClient.listMembers.mockResolvedValue({ users: sampleMembers });
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when response is empty', async () => {
    mockTeamClient.listMembers.mockResolvedValue({});
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockTeamClient.listMembers.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useErpTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useErpInviteMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls invite endpoint with correct payload', async () => {
    mockTeamClient.inviteMember.mockResolvedValue({ success: true });
    const { result } = renderHook(() => useErpInviteMember(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        email: 'new@example.com',
        fullName: 'New User',
        role: 'member',
      });
    });
    expect(mockTeamClient.inviteMember).toHaveBeenCalledWith({
      body: {
        email: 'new@example.com',
        fullName: 'New User',
        role: 'member',
        apps: ['erp'],
      },
    });
  });
});

describe('useErpDeactivateMember', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deactivate endpoint with member id', async () => {
    mockTeamClient.deactivateMember.mockResolvedValue({ success: true });
    const { result } = renderHook(() => useErpDeactivateMember(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync('user-123');
    });
    expect(mockTeamClient.deactivateMember).toHaveBeenCalledWith({ id: 'user-123', body: {} });
  });
});
