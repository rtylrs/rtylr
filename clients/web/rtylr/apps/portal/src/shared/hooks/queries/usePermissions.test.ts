import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  usePermissionCatalog,
  usePermissionProfiles,
  useCreatePermissionProfile,
} from './usePermissions';

const mockPermissionClient = vi.hoisted(() => ({
  listPermissionCatalog: vi.fn(),
  listRoles: vi.fn(),
  createRole: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
  getOneKitErrorMessage: vi.fn((_err, fallback) => fallback),
}));

vi.mock('../../lib/clients', () => ({ permissionClient: mockPermissionClient }));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper(client: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('permission-profile queries', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches the normalized permission catalog', async () => {
    mockPermissionClient.listPermissionCatalog.mockResolvedValue({
      permissions: [
        {
          key: 'flow:projects:read',
          appId: 'flow',
          resource: 'projects',
          action: 'read',
          isSensitive: false,
          isHighRisk: false,
        },
      ],
    });

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => usePermissionCatalog(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.permissions).toEqual([
      {
        key: 'flow:projects:read',
        appId: 'flow',
        resource: 'projects',
        action: 'read',
        isSensitive: false,
        isHighRisk: false,
      },
    ]);
  });

  it('fetches permission profiles successfully', async () => {
    mockPermissionClient.listRoles.mockResolvedValue({
      roles: [{ id: 1, name: 'Shift Lead', appId: 'flow', permissions: ['flow:projects:read'] }],
    });

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => usePermissionProfiles(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0].name).toBe('Shift Lead');
  });

  it('creates a permission profile successfully', async () => {
    mockPermissionClient.createRole.mockResolvedValue({ id: 1, name: 'Manager' });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreatePermissionProfile(), { wrapper });

    act(() => {
      result.current.mutate({ name: 'Manager', description: 'Store manager' });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockPermissionClient.createRole).toHaveBeenCalledWith({
      name: 'Manager',
      description: 'Store manager',
    });
  });
});
