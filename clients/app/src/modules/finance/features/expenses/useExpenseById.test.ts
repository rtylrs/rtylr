// @vitest-environment happy-dom
import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const client = vi.hoisted(() => ({ getExpense: vi.fn() }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => client,
  toFiniteNumber: (value: unknown) => Number(value ?? 0),
}));

import { useExpenseById } from './useExpenseById';

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(QueryClientProvider, { client: queryClient }, children);

beforeEach(() => vi.clearAllMocks());

describe('useExpenseById', () => {
  it('retains persisted optional expense fields', async () => {
    client.getExpense.mockResolvedValue({
      expense: {
        id: '1',
        categoryId: '2',
        description: 'Office supplies',
        amount: 42,
        currency: 'USD',
        vendor: 'Acme',
        receiptUrl: 'https://example.test/receipt',
        paymentMethod: 'Card',
        isTaxDeductible: true,
        taxAmount: 3.21,
        notes: 'Keep me',
        expenseDate: '2026-08-01',
        status: 'pending',
      },
    });

    const { result } = renderHook(() => useExpenseById('1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toMatchObject({
      paymentMethod: 'Card',
      isTaxDeductible: true,
      taxAmount: 3.21,
      notes: 'Keep me',
    });
  });

  it('rejects a successful envelope with no expense', async () => {
    client.getExpense.mockResolvedValue({ expense: undefined });

    const { result } = renderHook(() => useExpenseById('missing'), { wrapper });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.data).toBeUndefined();
  });
});
