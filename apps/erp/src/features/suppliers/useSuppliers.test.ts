import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useSuppliers, useCreateSupplier, useUpdateSupplier } from './useSuppliers';

const mockErpApi = {
  listSuppliers: vi.fn(),
  createSupplier: vi.fn(),
  updateSupplier: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listSuppliers: (...args: unknown[]) => mockErpApi.listSuppliers(...args),
    createSupplier: (...args: unknown[]) => mockErpApi.createSupplier(...args),
    updateSupplier: (...args: unknown[]) => mockErpApi.updateSupplier(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleSuppliers = [
  {
    id: 1,
    name: 'Acme',
    contactName: 'Alice',
    email: 'alice@acme.com',
    phone: '555',
    productCount: 10,
    categories: ['Tools'],
    status: 'Active' as const,
  },
  {
    id: 2,
    name: 'Beta',
    contactName: 'Bob',
    email: 'bob@beta.com',
    phone: '666',
    productCount: 5,
    categories: ['Electronics'],
    status: 'Inactive' as const,
  },
];

describe('useSuppliers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches suppliers with default pagination', async () => {
    mockErpApi.listSuppliers.mockResolvedValue(sampleSuppliers);
    const { result } = renderHook(() => useSuppliers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.suppliers).toHaveLength(2);
  });

  it('returns hasMore false when API does not signal more', async () => {
    mockErpApi.listSuppliers.mockResolvedValue(sampleSuppliers);
    const { result } = renderHook(() => useSuppliers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.hasMore).toBe(false);
  });

  it('returns hasMore true when API signals more', async () => {
    const response = { suppliers: sampleSuppliers, has_more: true };
    mockErpApi.listSuppliers.mockResolvedValue(response);
    const { result } = renderHook(() => useSuppliers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.hasMore).toBe(true);
  });

  it('returns empty suppliers when API returns null', async () => {
    mockErpApi.listSuppliers.mockResolvedValue(null);
    const { result } = renderHook(() => useSuppliers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.suppliers).toEqual([]);
  });

  it('handles API error', async () => {
    mockErpApi.listSuppliers.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useSuppliers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useCreateSupplier', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createSupplier with data', async () => {
    const newSupplier = { ...sampleSuppliers[0], id: 3 };
    mockErpApi.createSupplier.mockResolvedValue(newSupplier);
    const { result } = renderHook(() => useCreateSupplier(), { wrapper: createWrapper() });
    const input = {
      name: 'New Supplier',
      contactName: 'Carol',
      email: 'c@c.com',
      phone: '777',
      categories: [],
    };
    await act(async () => {
      await result.current.mutateAsync(input);
    });
    expect(mockErpApi.createSupplier).toHaveBeenCalledWith(input);
  });
});

describe('useUpdateSupplier', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateSupplier with data', async () => {
    mockErpApi.updateSupplier.mockResolvedValue(sampleSuppliers[0]);
    const { result } = renderHook(() => useUpdateSupplier(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ id: 1, name: 'Acme Updated' });
    });
    expect(mockErpApi.updateSupplier).toHaveBeenCalledWith({ id: 1, name: 'Acme Updated' });
  });
});
