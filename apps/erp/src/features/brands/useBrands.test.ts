import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useBrands, useCreateBrand, useDeleteBrand } from './useBrands';

const mockErpApi = {
  listBrands: vi.fn(),
  createBrand: vi.fn(),
  deleteBrand: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listBrands: (...args: unknown[]) => mockErpApi.listBrands(...args),
    createBrand: (...args: unknown[]) => mockErpApi.createBrand(...args),
    deleteBrand: (...args: unknown[]) => mockErpApi.deleteBrand(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleBrands = [
  { id: 1, name: 'Acme', productCount: 10 },
  { id: 2, name: 'BrandX', productCount: 5 },
];

describe('useBrands', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches brands list', async () => {
    mockErpApi.listBrands.mockResolvedValue(sampleBrands);
    const { result } = renderHook(() => useBrands(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when API returns null', async () => {
    mockErpApi.listBrands.mockResolvedValue(null);
    const { result } = renderHook(() => useBrands(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockErpApi.listBrands.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useBrands(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useCreateBrand', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createBrand with data', async () => {
    mockErpApi.createBrand.mockResolvedValue({ id: 3, name: 'NewBrand', productCount: 0 });
    const { result } = renderHook(() => useCreateBrand(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ name: 'NewBrand' });
    });
    expect(mockErpApi.createBrand).toHaveBeenCalledWith({ name: 'NewBrand' });
  });
});

describe('useDeleteBrand', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deleteBrand with id', async () => {
    mockErpApi.deleteBrand.mockResolvedValue({ success: true });
    const { result } = renderHook(() => useDeleteBrand(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync(1);
    });
    expect(mockErpApi.deleteBrand).toHaveBeenCalledWith(1);
  });
});
