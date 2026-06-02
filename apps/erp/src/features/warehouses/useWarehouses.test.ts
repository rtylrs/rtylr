import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useWarehouses,
  useCreateWarehouse,
  useUpdateWarehouse,
  useDeleteWarehouse,
} from './useWarehouses';

const mockSafe = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockSafe.get(...args),
    post: (...args: unknown[]) => mockSafe.post(...args),
    put: (...args: unknown[]) => mockSafe.put(...args),
    delete: (...args: unknown[]) => mockSafe.delete(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

// Response keys are camelCase: @voxire/api converts snake_case (manager_name etc.) on the way in.
const sampleWarehouses = [
  {
    id: 1,
    ownerId: 1,
    name: 'Main',
    address: '123 Main St',
    city: 'Beirut',
    country: 'LB',
    managerName: 'Alice',
    capacity: 1000,
    isDefault: true,
    isActive: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    ownerId: 1,
    name: 'Secondary',
    address: '456 Side Ave',
    city: '',
    country: '',
    managerName: 'Bob',
    capacity: 500,
    isDefault: false,
    isActive: false,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

describe('useWarehouses', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches warehouses list', async () => {
    mockSafe.get.mockResolvedValue({ warehouses: sampleWarehouses });
    const { result } = renderHook(() => useWarehouses(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
    expect(result.current.data?.[0].managerName).toBe('Alice');
  });

  it('returns empty array when API returns null', async () => {
    mockSafe.get.mockResolvedValue(null);
    const { result } = renderHook(() => useWarehouses(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockSafe.get.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useWarehouses(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useCreateWarehouse', () => {
  beforeEach(() => vi.clearAllMocks());

  it('posts the backend create payload shape', async () => {
    mockSafe.post.mockResolvedValue({ warehouseId: 3, message: 'ok' });
    const { result } = renderHook(() => useCreateWarehouse(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        name: 'New WH',
        address: '',
        managerName: 'Carol',
        capacity: 100,
      });
    });
    expect(mockSafe.post).toHaveBeenCalledWith('/erp/warehouses', {
      name: 'New WH',
      address: '',
      managerName: 'Carol',
      capacity: 100,
    });
  });
});

describe('useUpdateWarehouse', () => {
  beforeEach(() => vi.clearAllMocks());

  it('puts the update payload with id', async () => {
    mockSafe.put.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useUpdateWarehouse(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ id: 1, name: 'Renamed', isActive: false });
    });
    expect(mockSafe.put).toHaveBeenCalledWith('/erp/warehouses', {
      id: 1,
      name: 'Renamed',
      isActive: false,
    });
  });
});

describe('useDeleteWarehouse', () => {
  beforeEach(() => vi.clearAllMocks());

  it('deletes with id in the request body', async () => {
    mockSafe.delete.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useDeleteWarehouse(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync(1);
    });
    expect(mockSafe.delete).toHaveBeenCalledWith('/erp/warehouses', { data: { id: 1 } });
  });
});
