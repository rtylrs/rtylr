import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/financeApi', () => ({
  financeApi: {
    getExpenses: vi.fn(),
    getExpenseCategories: vi.fn(),
    createExpense: vi.fn(),
    updateExpense: vi.fn(),
    deleteExpense: vi.fn(),
    createExpenseCategory: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import {
  useExpenses,
  useExpenseCategories,
  useCreateExpense,
  useUpdateExpense,
  useDeleteExpense,
  useCreateExpenseCategory,
} from './useExpenses';
import { financeApi } from '../../shared/services/financeApi';

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
    const data = [{ id: 'exp-1', amount: 100 }];
    (financeApi.getExpenses as any).mockResolvedValue(data);
    const { result } = renderHook(() => useExpenses(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (financeApi.getExpenses as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useExpenses(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useExpenseCategories', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches expense categories', async () => {
    const cats = [{ id: 'cat-1', name: 'Office supplies' }];
    (financeApi.getExpenseCategories as any).mockResolvedValue(cats);
    const { result } = renderHook(() => useExpenseCategories(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(cats);
  });
});

describe('useCreateExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createExpense service method', async () => {
    (financeApi.createExpense as any).mockResolvedValue({ id: 'exp-new' });
    const { result } = renderHook(() => useCreateExpense(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ amount: 75 } as any));
    expect(financeApi.createExpense).toHaveBeenCalledWith({ amount: 75 });
  });
});

describe('useUpdateExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateExpense with id and data', async () => {
    (financeApi.updateExpense as any).mockResolvedValue({ id: 'exp-1' });
    const { result } = renderHook(() => useUpdateExpense(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 'exp-1', data: { amount: 150 } as any }));
    expect(financeApi.updateExpense).toHaveBeenCalledWith('exp-1', { amount: 150 });
  });
});

describe('useDeleteExpense', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteExpense with id', async () => {
    (financeApi.deleteExpense as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useDeleteExpense(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync('exp-2'));
    expect(financeApi.deleteExpense).toHaveBeenCalledWith('exp-2');
  });
});

describe('useCreateExpenseCategory', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createExpenseCategory service method', async () => {
    (financeApi.createExpenseCategory as any).mockResolvedValue({ id: 'cat-new' });
    const { result } = renderHook(() => useCreateExpenseCategory(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ name: 'Travel' } as any));
    expect(financeApi.createExpenseCategory).toHaveBeenCalledWith({ name: 'Travel' });
  });
});
