// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const { mockClient, clientServices, toastMock } = vi.hoisted(() => ({
  mockClient: {
    listExpenseCategories: vi.fn(),
    createExpenseCategory: vi.fn(),
    updateExpenseCategory: vi.fn(),
    deleteExpenseCategory: vi.fn(),
  },
  clientServices: [] as unknown[],
  toastMock: { success: vi.fn(), info: vi.fn(), error: vi.fn() },
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (service: unknown) => {
    clientServices.push(service);
    return mockClient;
  },
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  toFiniteNumber: (value: unknown, fallback = 0) => {
    if (value === null || value === undefined || value === '') return fallback;
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  },
}));

vi.mock('sonner', () => ({ toast: toastMock }));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (_key: string, fallback?: string, vars?: Record<string, string>) =>
      (fallback ?? _key).replace(/{{(\w+)}}/g, (_, name: string) => vars?.[name] ?? ''),
  }),
}));

import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import {
  useCreateExpenseCategory,
  useDeleteExpenseCategory,
  useExpenseCategories,
  useUpdateExpenseCategory,
} from './useExpenseCategories';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('expense category hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Regression: creates went to finance.report.v1 while the list came from
  // finance.expenses.v1, so a new category never showed up. Both sides must
  // talk to the same service.
  it('reads and writes through finance.expenses.v1 only', () => {
    expect(clientServices).toEqual([ExpensesServiceClient]);
  });

  it('shows a created category on the next list read', async () => {
    const stored: Array<{ id: string; name: string; color: string }> = [];
    mockClient.listExpenseCategories.mockImplementation(async () => ({ categories: [...stored] }));
    mockClient.createExpenseCategory.mockImplementation(
      async (req: { name: string; color: string }) => {
        const category = { id: '7', name: req.name, color: req.color };
        stored.push(category);
        return { category, created: true };
      }
    );

    const wrapper = createWrapper();
    const { result } = renderHook(
      () => ({ list: useExpenseCategories(), create: useCreateExpenseCategory() }),
      { wrapper }
    );
    await waitFor(() => expect(result.current.list.data).toEqual([]));

    await act(async () => {
      await result.current.create.mutateAsync({ name: 'Communications', color: '#14b8a6' });
    });

    expect(mockClient.createExpenseCategory).toHaveBeenCalledWith({
      name: 'Communications',
      color: '#14b8a6',
    });
    await waitFor(() =>
      expect(result.current.list.data).toEqual([
        { id: 7, name: 'Communications', color: '#14b8a6' },
      ])
    );
    expect(toastMock.success).toHaveBeenCalledWith('Category created');
  });

  it('says so when the name already exists instead of claiming a new category', async () => {
    mockClient.createExpenseCategory.mockResolvedValue({
      category: { id: '7', name: 'Communications', color: '#14b8a6' },
      created: false,
    });
    const { result } = renderHook(() => useCreateExpenseCategory(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ name: 'communications', color: '#000000' });
    });

    expect(toastMock.success).not.toHaveBeenCalled();
    expect(toastMock.info).toHaveBeenCalledWith('“Communications” already exists');
  });

  it('updates and deletes by id through the expenses service', async () => {
    mockClient.updateExpenseCategory.mockResolvedValue({
      category: { id: '3', name: 'Rent', color: '#111111' },
    });
    mockClient.deleteExpenseCategory.mockResolvedValue({ message: 'ok' });
    const wrapper = createWrapper();
    const { result } = renderHook(
      () => ({ update: useUpdateExpenseCategory(), remove: useDeleteExpenseCategory() }),
      { wrapper }
    );

    await act(async () => {
      await result.current.update.mutateAsync({ id: 3, data: { name: 'Rent', color: '#111111' } });
      await result.current.remove.mutateAsync(3);
    });

    expect(mockClient.updateExpenseCategory).toHaveBeenCalledWith({
      id: '3',
      body: { name: 'Rent', color: '#111111' },
    });
    expect(mockClient.deleteExpenseCategory).toHaveBeenCalledWith({ id: '3' });
  });

  it('surfaces a refused delete instead of closing silently', async () => {
    mockClient.deleteExpenseCategory.mockRejectedValue(new Error('409'));
    const { result } = renderHook(() => useDeleteExpenseCategory(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync(3).catch(() => undefined);
    });

    expect(toastMock.error).toHaveBeenCalled();
    expect(toastMock.success).not.toHaveBeenCalled();
  });
});
