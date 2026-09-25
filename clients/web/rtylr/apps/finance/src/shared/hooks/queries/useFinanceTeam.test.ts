// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockClient = vi.hoisted(() => ({ listTeam: vi.fn() }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
  createRestRouteClient: () => ({}),
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  oneKitResult: (operation: () => Promise<unknown>) => operation(),
  unwrap: (result: Promise<unknown>) => result,
}));

import { useFinanceTeam } from './useFinanceTeam';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

// hub.team.members.v1 StaffUserPayload has no `name`; the displayed name is
// full_name, decoded as fullName (see the generated types).
const member = {
  id: '7',
  email: 'jane@x.com',
  fullName: 'Jane Doe',
  role: 'staff',
  apps: ['finance'],
  isActive: true,
};

describe('useFinanceTeam', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('maps the contract full_name to fullName', async () => {
    mockClient.listTeam.mockResolvedValue({ users: [member] });
    const { result } = renderHook(() => useFinanceTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]).toEqual({
      id: '7',
      fullName: 'Jane Doe',
      email: 'jane@x.com',
      role: 'staff',
      apps: ['finance'],
      isActive: true,
    });
  });

  it('falls back to the email when full_name is empty', async () => {
    mockClient.listTeam.mockResolvedValue({ users: [{ ...member, fullName: '' }] });
    const { result } = renderHook(() => useFinanceTeam(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.fullName).toBe('jane@x.com');
  });
});
