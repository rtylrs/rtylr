import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useProducts,
  useProduct,
  useCreateProduct,
  useUpdateProduct,
  useDeleteProduct,
} from './useProducts';

const mockStoreApi = {
  listProducts: vi.fn(),
  getProduct: vi.fn(),
  createProduct: vi.fn(),
  updateProduct: vi.fn(),
  deleteProduct: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/storeService', () => ({
  storeApi: {
    listProducts: (...args: unknown[]) => mockStoreApi.listProducts(...args),
    getProduct: (...args: unknown[]) => mockStoreApi.getProduct(...args),
    createProduct: (...args: unknown[]) => mockStoreApi.createProduct(...args),
    updateProduct: (...args: unknown[]) => mockStoreApi.updateProduct(...args),
    deleteProduct: (...args: unknown[]) => mockStoreApi.deleteProduct(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleProducts = [
  {
    id: 1,
    name: 'T-Shirt',
    sku: 'TSH-001',
    category: 'Apparel',
    price: 29.99,
    stock: 100,
    status: 'active',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    name: 'Mug',
    sku: 'MUG-001',
    category: 'Drinkware',
    price: 12.99,
    stock: 50,
    status: 'draft',
    createdAt: '2026-01-02T00:00:00Z',
    updatedAt: '2026-01-02T00:00:00Z',
  },
];

describe('useProducts', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches products without filters', async () => {
    mockStoreApi.listProducts.mockResolvedValue(sampleProducts);

    const { result } = renderHook(() => useProducts(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(result.current.data![0].name).toBe('T-Shirt');
  });

  it('passes filters to listProducts', async () => {
    mockStoreApi.listProducts.mockResolvedValue([]);

    const filters = { search: 'shirt', status: 'active' as const, limit: 10 };
    renderHook(() => useProducts(filters), { wrapper: createWrapper() });

    await waitFor(() =>
      expect(mockStoreApi.listProducts).toHaveBeenCalledWith({
        search: 'shirt',
        categoryId: undefined,
        status: 'active',
        limit: 10,
        offset: undefined,
      })
    );
  });

  it('handles API error', async () => {
    mockStoreApi.listProducts.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useProducts(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('uses query key with filters for cache separation', async () => {
    mockStoreApi.listProducts.mockResolvedValue([]);

    const w = createWrapper();
    renderHook(() => useProducts({ search: 'a' }), { wrapper: w });
    renderHook(() => useProducts({ search: 'b' }), { wrapper: w });

    await waitFor(() => expect(mockStoreApi.listProducts).toHaveBeenCalledTimes(2));
  });
});

describe('useProduct', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single product by id', async () => {
    mockStoreApi.getProduct.mockResolvedValue(sampleProducts[0]);

    const { result } = renderHook(() => useProduct(1), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.name).toBe('T-Shirt');
    expect(mockStoreApi.getProduct).toHaveBeenCalledWith(1);
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useProduct(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCreateProduct', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createProduct and invalidates queries', async () => {
    mockStoreApi.createProduct.mockResolvedValue({ id: 3, name: 'Hat' });

    const { result } = renderHook(() => useCreateProduct(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ name: 'Hat', sku: 'HAT-001', price: 19.99, stock: 25 });
    });

    expect(mockStoreApi.createProduct).toHaveBeenCalledWith({
      name: 'Hat',
      sku: 'HAT-001',
      price: 19.99,
      stock: 25,
    });
  });
});

describe('useUpdateProduct', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateProduct', async () => {
    mockStoreApi.updateProduct.mockResolvedValue({ id: 1, name: 'Updated T-Shirt' });

    const { result } = renderHook(() => useUpdateProduct(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ id: 1, name: 'Updated T-Shirt' });
    });

    expect(mockStoreApi.updateProduct).toHaveBeenCalledWith({ id: 1, name: 'Updated T-Shirt' });
  });
});

describe('useDeleteProduct', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deleteProduct with id', async () => {
    mockStoreApi.deleteProduct.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useDeleteProduct(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync(1);
    });

    expect(mockStoreApi.deleteProduct).toHaveBeenCalledWith(1);
  });
});
