import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  locationFlagsForRole,
  toAssignLocationPayload,
  useMemberLocations,
  usePOSLocations,
} from './usePOSAccess';

const mockGet = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
    post: vi.fn(),
  },
}));

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
    mockGet.mockResolvedValueOnce({
      locations: [{ location_id: 3, name: 'Main', address: 'A' }],
    });

    const { result } = renderHook(() => usePOSLocations(true), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGet).toHaveBeenCalledWith('/pos/access/my-locations');
    expect(result.current.data).toEqual([{ id: 3, name: 'Main', address: 'A' }]);
  });
});

describe('useMemberLocations', () => {
  it('preserves all POS access flags from the backend', async () => {
    mockGet.mockResolvedValueOnce({
      locations: [
        {
          id: 9,
          location_id: 3,
          can_manage_pos_users: true,
          can_manage_terminals: true,
          can_view_reports: true,
        },
      ],
    });

    const { result } = renderHook(() => useMemberLocations(42, true), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

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

describe('toAssignLocationPayload', () => {
  it('maps dashboard location assignments to the backend snake_case contract', () => {
    expect(
      toAssignLocationPayload({
        userId: 42,
        locationId: 3,
        canManagePosUsers: true,
        canManageTerminals: false,
        canViewReports: true,
      })
    ).toEqual({
      user_id: 42,
      location_id: 3,
      can_manage_pos_users: true,
      can_manage_terminals: false,
      can_view_reports: true,
    });
  });
});

describe('locationFlagsForRole', () => {
  it('grants store managers management flags for assigned locations', () => {
    expect(locationFlagsForRole('store_manager')).toEqual({
      canManagePosUsers: true,
      canManageTerminals: true,
      canViewReports: true,
    });
  });

  it('keeps non-manager assignments as plain location access', () => {
    expect(locationFlagsForRole('staff')).toEqual({
      canManagePosUsers: false,
      canManageTerminals: false,
      canViewReports: false,
    });
  });
});
