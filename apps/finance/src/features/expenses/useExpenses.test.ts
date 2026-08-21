// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockClient = vi.hoisted(() => ({
  listExpenses: vi.fn(),
  listExpenseCategories: vi.fn(),
  createExpense: vi.fn(),
  updateExpense: vi.fn(),
  deleteExpense: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  toFiniteNumber: (value: unknown, fallback = 0) => {
    if (value === null || value === undefined || value === '') return fallback;
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  },
}));

import {
  useExpenses,
  useExpenseCategories,
  useCreateExpense,
  useUpdateExpense,
  useUpdateExpenseReceipt,
  useDeleteExpense,
} from './useExpenses';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useExpenses', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches expenses on mount', async () => {
    mockClient.listExpenses.mockResolvedValue({ expenses: [] });
    const { result } = renderHook(() => useExpenses(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.listExpenses).toHaveBeenCalledWith({});
    expect(result.current.data?.expenses).toEqual([]);
  });

  it('paginates client-side over the full result set', async () => {
    mockClient.listExpenses.mockResolvedValue({
      expenses: Array.from({ length: 55 }, (_, i) => ({
        id: `${i}`,
        categoryId: null,
        description: '',
        amount: 0,
        currency: 'USD',
        vendor: '',
        receiptUrl: '',
        expenseDate: '',
        status: 'pending',
        createdAt: '',
        updatedAt: '',
      })),
    });
    const { result } = renderHook(() => useExpenses({ page: 1, pageSize: 50 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.expenses).toHaveLength(50);
    expect(result.current.data?.hasMore).toBe(true);
  });

  it('starts in loading state', () => {
    mockClient.listExpenses.mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useExpenses(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useExpenseCategories', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches expense categories', async () => {
    mockClient.listExpenseCategories.mockResolvedValue({
      categories: [
        { id: '1', ownerId: '1', name: 'Office supplies', color: '#3b82f6', createdAt: '' },
      ],
    });
    const { result } = renderHook(() => useExpenseCategories(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([{ id: 1, name: 'Office supplies', color: '#3b82f6' }]);
  });
});

describe('useCreateExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('creates an expense', async () => {
    mockClient.createExpense.mockResolvedValue({ expenseId: 'exp-new', message: 'ok' });
    const { result } = renderHook(() => useCreateExpense(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ amount: 75, notes: 'Client dinner' }));
    expect(mockClient.createExpense).toHaveBeenCalledWith(
      expect.objectContaining({ amount: 75, notes: 'Client dinner' })
    );
  });
});

describe('useUpdateExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('updates an expense', async () => {
    mockClient.updateExpense.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useUpdateExpense(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: 'exp-1', data: { amount: 150 } }));
    expect(mockClient.updateExpense).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'exp-1', body: expect.objectContaining({ amount: 150 }) })
    );
  });
});

describe('useUpdateExpenseReceipt', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('persists the receipt url alongside the existing expense fields', async () => {
    mockClient.updateExpense.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useUpdateExpenseReceipt(), { wrapper: createWrapper() });
    await act(() =>
      result.current.mutateAsync({
        id: 'exp-1',
        receiptUrl: 'https://cdn.example.com/receipt.png',
        expense: {
          id: 'exp-1',
          description: 'Office supplies',
          amount: 42,
          currency: 'USD',
          vendor: 'Acme',
          expenseDate: '2026-08-01',
          status: 'pending',
        },
      })
    );
    expect(mockClient.updateExpense).toHaveBeenCalledWith(
      expect.objectContaining({
        id: 'exp-1',
        body: expect.objectContaining({
          receiptUrl: 'https://cdn.example.com/receipt.png',
          description: 'Office supplies',
          vendor: 'Acme',
          expenseDate: '2026-08-01',
        }),
      })
    );
  });
});

describe('useDeleteExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('deletes an expense', async () => {
    mockClient.deleteExpense.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useDeleteExpense(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync('exp-2'));
    expect(mockClient.deleteExpense).toHaveBeenCalledWith({ id: 'exp-2' });
  });
});
