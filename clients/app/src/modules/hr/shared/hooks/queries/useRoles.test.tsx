import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
}));

vi.mock('../../lib/clients', () => ({
  hrTransport: { get: mocks.get, post: mocks.post, put: mocks.put },
}));

vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => value,
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { useListRoles, useListPermissionCatalog, useCreateRole, useUpdateRole } from './useRoles';

function createWrapper(): React.FC<{ children: React.ReactNode }> {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('useListRoles', () => {
  beforeEach(() => {
    mocks.get.mockReset();
    mocks.post.mockReset();
    mocks.put.mockReset();
  });

  it('parses the permissions JSON string into an object', async () => {
    mocks.get.mockResolvedValue({
      roles: [
        {
          id: 1,
          ownerId: 1001,
          name: 'Manager',
          description: 'Team manager',
          permissions: '{"approve_time_off":true,"manage_payroll":false}',
          isSystem: true,
          createdAt: '2026-01-01T00:00:00Z',
          updatedAt: '2026-01-01T00:00:00Z',
        },
      ],
    });

    const { result } = renderHook(() => useListRoles(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        id: 1,
        ownerId: 1001,
        name: 'Manager',
        description: 'Team manager',
        permissions: { approve_time_off: true, manage_payroll: false },
        isSystem: true,
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      },
    ]);
  });
});

describe('useListPermissionCatalog', () => {
  beforeEach(() => {
    mocks.get.mockReset();
  });

  it('returns the permission catalog entries', async () => {
    mocks.get.mockResolvedValue({
      permissions: [
        {
          key: 'manage_employees',
          label: 'Manage employees',
          description: 'Create and edit employees',
          category: 'employees',
        },
      ],
    });

    const { result } = renderHook(() => useListPermissionCatalog(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        key: 'manage_employees',
        label: 'Manage employees',
        description: 'Create and edit employees',
        category: 'employees',
      },
    ]);
  });
});

describe('useUpdateRole', () => {
  beforeEach(() => {
    mocks.put.mockReset();
  });

  it('PUTs to /hr/roles/{id} with the stringified permissions', async () => {
    mocks.put.mockResolvedValue({
      role: {
        id: 1,
        ownerId: 1001,
        name: 'Manager',
        description: 'Team manager',
        permissions: '{"approve_time_off":true}',
        isSystem: true,
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      },
    });

    const { result } = renderHook(() => useUpdateRole(), { wrapper: createWrapper() });

    result.current.mutate({
      id: 1,
      name: 'Manager',
      description: 'Team manager',
      permissions: '{"approve_time_off":true}',
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.put).toHaveBeenCalledWith('/hr/roles/1', {
      name: 'Manager',
      description: 'Team manager',
      permissions: '{"approve_time_off":true}',
    });
  });
});

describe('useCreateRole', () => {
  beforeEach(() => {
    mocks.post.mockReset();
  });

  it('POSTs to /hr/roles', async () => {
    mocks.post.mockResolvedValue({
      role: {
        id: 2,
        ownerId: 1001,
        name: 'Custom Role',
        description: '',
        permissions: '{}',
        isSystem: false,
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      },
    });

    const { result } = renderHook(() => useCreateRole(), { wrapper: createWrapper() });

    result.current.mutate({ name: 'Custom Role', description: '', permissions: '{}' });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.post).toHaveBeenCalledWith('/hr/roles', {
      name: 'Custom Role',
      description: '',
      permissions: '{}',
    });
  });
});
