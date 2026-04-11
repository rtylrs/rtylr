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

const mockErpApi = {
  listProducts: vi.fn(),
  getProduct: vi.fn(),
  createProduct: vi.fn(),
  updateProduct: vi.fn(),
  deleteProduct: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listProducts: (...args: any[]) => mockErpApi.listProducts(...args),
    getProduct: (...args: any[]) => mockErpApi.getProduct(...args),
    createProduct: (...args: any[]) => mockErpApi.createProduct(...args),
    updateProduct: (...args: any[]) => mockErpApi.updateProduct(...args),
    deleteProduct: (...args: any[]) => mockErpApi.deleteProduct(...args),
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
    name: 'Widget A',
    sku: 'WA-001',
    category: 'Electronics',
    stock: 100,
    lowStock: 10,
    costPrice: 5,
    salePrice: 15,
    status: 'In Stock' as const,
    brand: 'BrandA',
  },
  {
    id: 2,
    name: 'Widget B',
    sku: 'WB-001',
    category: 'Hardware',
    stock: 3,
    lowStock: 10,
    costPrice: 10,
    salePrice: 30,
    status: 'Low Stock' as const,
    brand: 'BrandB',
  },
  {
    id: 3,
    name: 'Gadget C',
    sku: 'GC-001',
    category: 'Electronics',
    stock: 50,
    lowStock: 5,
    costPrice: 8,
    salePrice: 25,
    status: 'In Stock' as const,
    brand: 'BrandA',
  },
];

describe('useProducts', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches products without filters', async () => {
    mockErpApi.listProducts.mockResolvedValue(sampleProducts);

    const { result } = renderHook(() => useProducts(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(3);
  });

  it('applies client-side search filter', async () => {
    mockErpApi.listProducts.mockResolvedValue(sampleProducts);

    const { result } = renderHook(() => useProducts({ search: 'widget' }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
  });

  it('applies client-side category filter', async () => {
    mockErpApi.listProducts.mockResolvedValue(sampleProducts);

    const { result } = renderHook(() => useProducts({ category: 'Electronics' }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(result.current.data!.every((p) => p.category === 'Electronics')).toBe(true);
  });

  it('applies client-side brand filter', async () => {
    mockErpApi.listProducts.mockResolvedValue(sampleProducts);

    const { result } = renderHook(() => useProducts({ brand: 'BrandB' }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
    expect(result.current.data![0].name).toBe('Widget B');
  });

  it('applies client-side status filter', async () => {
    mockErpApi.listProducts.mockResolvedValue(sampleProducts);

    const { result } = renderHook(() => useProducts({ status: 'Low Stock' }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
  });

  it('sorts by name ascending', async () => {
    mockErpApi.listProducts.mockResolvedValue(sampleProducts);

    const { result } = renderHook(() => useProducts({ sort: 'name', sortDir: 'asc' }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data![0].name).toBe('Gadget C');
    expect(result.current.data![2].name).toBe('Widget B');
  });

  it('sorts by price descending', async () => {
    mockErpApi.listProducts.mockResolvedValue(sampleProducts);

    const { result } = renderHook(() => useProducts({ sort: 'price', sortDir: 'desc' }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data![0].salePrice).toBe(30);
  });

  it('handles null response from API', async () => {
    mockErpApi.listProducts.mockResolvedValue(null);

    const { result } = renderHook(() => useProducts(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([]);
  });
});

describe('useProduct', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single product by id', async () => {
    mockErpApi.getProduct.mockResolvedValue(sampleProducts[0]);

    const { result } = renderHook(() => useProduct(1), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.name).toBe('Widget A');
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useProduct(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCreateProduct', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createProduct', async () => {
    mockErpApi.createProduct.mockResolvedValue({ id: 4, name: 'New Product' });

    const { result } = renderHook(() => useCreateProduct(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        name: 'New Product',
        sku: 'NP-001',
        category: 'Test',
        stock: 10,
        lowStock: 2,
        costPrice: 5,
        salePrice: 15,
      });
    });

    expect(mockErpApi.createProduct).toHaveBeenCalled();
  });
});

describe('useDeleteProduct', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deleteProduct with id', async () => {
    mockErpApi.deleteProduct.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useDeleteProduct(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync(1);
    });

    expect(mockErpApi.deleteProduct).toHaveBeenCalledWith(1);
  });
});
