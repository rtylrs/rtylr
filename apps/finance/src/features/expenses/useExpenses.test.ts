// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/lib/api', () => ({
  safe: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val: unknown) => Promise.resolve(val)),
}));

import {
  useExpenses,
  useExpenseCategories,
  useCreateExpense,
  useUpdateExpense,
  useDeleteExpense,
  useCreateExpenseCategory,
} from './useExpenses';
import { safe } from '../../shared/lib/api';

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
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ expenses: [], has_more: false });
    const { result } = renderHook(() => useExpenses(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/expenses', {
      params: { page: 1, page_size: 50 },
    });
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useExpenses(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useExpenseCategories', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches expense categories', async () => {
    const input = [{ id: 1, name: 'Office supplies', color: '#3b82f6' }];
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue(input);
    const { result } = renderHook(() => useExpenseCategories(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/expenses/categories');
  });
});

describe('useCreateExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls POST /finance/expenses', async () => {
    (safe.post as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 'exp-new' });
    const { result } = renderHook(() => useCreateExpense(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ amount: 75 } as any));
    expect(safe.post).toHaveBeenCalledWith('/finance/expenses', { amount: 75 });
  });
});

describe('useUpdateExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls PUT /finance/expenses/:id', async () => {
    (safe.put as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 'exp-1' });
    const { result } = renderHook(() => useUpdateExpense(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: 'exp-1', data: { amount: 150 } as any }));
    expect(safe.put).toHaveBeenCalledWith('/finance/expenses/exp-1', { amount: 150 });
  });
});

describe('useDeleteExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls DELETE /finance/expenses/:id', async () => {
    (safe.delete as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    const { result } = renderHook(() => useDeleteExpense(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync('exp-2'));
    expect(safe.delete).toHaveBeenCalledWith('/finance/expenses/exp-2');
  });
});

describe('useCreateExpenseCategory', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls POST /finance/expense-categories', async () => {
    (safe.post as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 'cat-new' });
    const { result } = renderHook(() => useCreateExpenseCategory(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ name: 'Travel' } as any));
    expect(safe.post).toHaveBeenCalledWith('/finance/expense-categories', { name: 'Travel' });
  });
});
