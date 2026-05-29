import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useCategories,
  useCreateCategory,
  useUpdateCategory,
  useDeleteCategory,
} from './useCategories';

const mockErpApi = {
  listCategories: vi.fn(),
  createCategory: vi.fn(),
  updateCategory: vi.fn(),
  deleteCategory: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listCategories: (...args: unknown[]) => mockErpApi.listCategories(...args),
    createCategory: (...args: unknown[]) => mockErpApi.createCategory(...args),
    updateCategory: (...args: unknown[]) => mockErpApi.updateCategory(...args),
    deleteCategory: (...args: unknown[]) => mockErpApi.deleteCategory(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleCategories = [
  { id: 1, name: 'Electronics', parentId: null, productCount: 50, sortOrder: 1 },
  { id: 2, name: 'Phones', parentId: 1, productCount: 30, sortOrder: 2 },
];

describe('useCategories', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches categories list', async () => {
    mockErpApi.listCategories.mockResolvedValue(sampleCategories);
    const { result } = renderHook(() => useCategories(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when API returns null', async () => {
    mockErpApi.listCategories.mockResolvedValue(null);
    const { result } = renderHook(() => useCategories(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockErpApi.listCategories.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useCategories(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useCreateCategory', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createCategory with data', async () => {
    mockErpApi.createCategory.mockResolvedValue({
      id: 3,
      name: 'Tablets',
      parentId: null,
      productCount: 0,
      sortOrder: 3,
    });
    const { result } = renderHook(() => useCreateCategory(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ name: 'Tablets' });
    });
    expect(mockErpApi.createCategory).toHaveBeenCalledWith({ name: 'Tablets' });
  });
});

describe('useUpdateCategory', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateCategory with data', async () => {
    mockErpApi.updateCategory.mockResolvedValue({ id: 1, name: 'Consumer Electronics' });
    const { result } = renderHook(() => useUpdateCategory(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ id: 1, name: 'Consumer Electronics' });
    });
    expect(mockErpApi.updateCategory).toHaveBeenCalledWith({ id: 1, name: 'Consumer Electronics' });
  });
});

describe('useDeleteCategory', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deleteCategory with id', async () => {
    mockErpApi.deleteCategory.mockResolvedValue({ success: true });
    const { result } = renderHook(() => useDeleteCategory(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync(2);
    });
    expect(mockErpApi.deleteCategory).toHaveBeenCalledWith(2);
  });
});
