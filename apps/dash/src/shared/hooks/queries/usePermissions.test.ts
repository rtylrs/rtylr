import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  usePermissions,
  useCustomRoles,
  useUpdatePermissions,
  useCreateCustomRole,
} from './usePermissions';

const mockPermissionClient = vi.hoisted(() => ({
  getPermissions: vi.fn(),
  listRoles: vi.fn(),
  updatePermissionMatrix: vi.fn(),
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

describe('usePermissions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches permissions matrix successfully', async () => {
    mockPermissionClient.getPermissions.mockResolvedValue({
      permissions: [{ key: 'orders.view', label: 'View Orders', group: 'Orders', description: '' }],
      roles: ['admin', 'staff'],
      values: { admin: { 'orders.view': true } },
    });

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => usePermissions(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.permissions).toHaveLength(1);
    expect(result.current.data?.values.admin['orders.view']).toBe(true);
  });

  it('fetches custom roles successfully', async () => {
    mockPermissionClient.listRoles.mockResolvedValue({
      roles: [
        { id: 1, name: 'Shift Lead', color: '#123456', permissions: ['orders.view'], members: 1 },
      ],
    });

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCustomRoles(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0].name).toBe('Shift Lead');
  });

  it('updates permissions successfully', async () => {
    mockPermissionClient.updatePermissionMatrix.mockResolvedValue({ success: true });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useUpdatePermissions(), { wrapper });

    act(() => {
      result.current.mutate({ role: 'staff', key: 'orders.view', enabled: true });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockPermissionClient.updatePermissionMatrix).toHaveBeenCalledWith({
      role: 'staff',
      permissionKey: 'orders.view',
      enabled: true,
    });
  });

  it('creates custom role successfully', async () => {
    mockPermissionClient.createRole.mockResolvedValue({ id: 1, name: 'Manager' });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreateCustomRole(), { wrapper });

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
