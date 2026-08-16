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

const mockGet = vi.fn();
const mockPut = vi.fn();
const mockPost = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
  getOneKitErrorMessage: vi.fn((_err, fallback) => fallback),
}));

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
    put: (...args: unknown[]) => mockPut(...args),
    post: (...args: unknown[]) => mockPost(...args),
  },
}));

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
    mockGet.mockResolvedValue({
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
    mockGet.mockResolvedValue({
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
    mockPut.mockResolvedValue({ success: true });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useUpdatePermissions(), { wrapper });

    act(() => {
      result.current.mutate({ role: 'staff', key: 'orders.view', enabled: true });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockPut).toHaveBeenCalledWith('/hub/permission/v1/permissions', {
      role: 'staff',
      key: 'orders.view',
      enabled: true,
    });
  });

  it('creates custom role successfully', async () => {
    mockPost.mockResolvedValue({ id: 1, name: 'Manager' });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreateCustomRole(), { wrapper });

    act(() => {
      result.current.mutate({ name: 'Manager', description: 'Store manager' });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockPost).toHaveBeenCalledWith('/hub/permission/v1/roles', {
      name: 'Manager',
      description: 'Store manager',
    });
  });
});
