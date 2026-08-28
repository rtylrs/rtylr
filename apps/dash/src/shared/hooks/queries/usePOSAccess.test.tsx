import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockClient = vi.hoisted(() => ({
  getMyLocations: vi.fn(),
  getMemberLocations: vi.fn(),
  assignLocation: vi.fn(),
  removeLocation: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { useMemberLocations, usePOSLocations } from './usePOSAccess';

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('usePOSLocations', () => {
  it('loads locations from the access endpoint', async () => {
    mockClient.getMyLocations.mockResolvedValueOnce({
      locations: [{ locationId: '3', name: 'Main', address: 'A', phone: '' }],
    });

    const { result } = renderHook(() => usePOSLocations(true), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockClient.getMyLocations).toHaveBeenCalledWith({});
    expect(result.current.data).toEqual([{ id: 3, name: 'Main', address: 'A' }]);
  });
});

describe('useMemberLocations', () => {
  it('preserves all POS access flags from the backend', async () => {
    mockClient.getMemberLocations.mockResolvedValueOnce({
      locations: [
        {
          id: '9',
          locationId: '3',
          canManagePosUsers: true,
          canManageTerminals: true,
          canViewReports: true,
        },
      ],
    });

    const { result } = renderHook(() => useMemberLocations(42, true), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockClient.getMemberLocations).toHaveBeenCalledWith({ userId: '42' });
    expect(result.current.data).toEqual([
      {
        id: 9,
        locationId: 3,
        canManagePosUsers: true,
        canManageTerminals: true,
        canViewReports: true,
      },
    ]);
  });
});
