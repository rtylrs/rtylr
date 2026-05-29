import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useWarehouses, useCreateWarehouse, useDeleteWarehouse } from './useWarehouses';

const mockErpApi = {
  listWarehouses: vi.fn(),
  createWarehouse: vi.fn(),
  deleteWarehouse: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listWarehouses: (...args: unknown[]) => mockErpApi.listWarehouses(...args),
    createWarehouse: (...args: unknown[]) => mockErpApi.createWarehouse(...args),
    deleteWarehouse: (...args: unknown[]) => mockErpApi.deleteWarehouse(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleWarehouses = [
  { id: 1, name: 'Main', address: '123 Main St', manager: 'Alice', capacity: 1000, used: 750, isDefault: true },
  { id: 2, name: 'Secondary', address: '456 Side Ave', manager: 'Bob', capacity: 500, used: 200, isDefault: false },
];

describe('useWarehouses', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches warehouses list', async () => {
    mockErpApi.listWarehouses.mockResolvedValue(sampleWarehouses);
    const { result } = renderHook(() => useWarehouses(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when API returns null', async () => {
    mockErpApi.listWarehouses.mockResolvedValue(null);
    const { result } = renderHook(() => useWarehouses(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockErpApi.listWarehouses.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useWarehouses(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useCreateWarehouse', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createWarehouse with data', async () => {
    mockErpApi.createWarehouse.mockResolvedValue({ id: 3, name: 'New WH', address: '', manager: '', capacity: 100, used: 0, isDefault: false });
    const { result } = renderHook(() => useCreateWarehouse(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ name: 'New WH' });
    });
    expect(mockErpApi.createWarehouse).toHaveBeenCalledWith({ name: 'New WH' });
  });
});

describe('useDeleteWarehouse', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deleteWarehouse with id', async () => {
    mockErpApi.deleteWarehouse.mockResolvedValue({ success: true });
    const { result } = renderHook(() => useDeleteWarehouse(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync(1);
    });
    expect(mockErpApi.deleteWarehouse).toHaveBeenCalledWith(1);
  });
});
