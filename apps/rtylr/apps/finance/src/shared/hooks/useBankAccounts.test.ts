// @vitest-environment happy-dom
import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { getBankFeedsMock } = vi.hoisted(() => ({
  getBankFeedsMock: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({}),
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  unwrap: async (value: unknown) => value,
}));

vi.mock('@voxire/voxie/finance/bank_account/v1/client', () => ({
  BankAccountServiceClient: class {},
}));

vi.mock('../services/financeApi', () => ({
  financeApi: { getBankFeeds: getBankFeedsMock },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}));

import { useBankTransactions } from './useBankAccounts';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useBankTransactions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('maps the real bank-feed wire row instead of treating it as invalid', async () => {
    getBankFeedsMock.mockResolvedValue({
      transactions: [
        {
          id: 7,
          ownerId: 12,
          bankAccountId: 3,
          amountMinorUnits: -1234,
          currency: 'USD',
          occurredAt: '2026-09-20T12:30:00Z',
          description: 'Coffee shop',
          externalId: 'ext-7',
          matchedExpenseId: null,
          matchedInvoiceId: 99,
          createdAt: '2026-09-20T12:31:00Z',
        },
      ],
    });

    const { result } = renderHook(() => useBankTransactions('account-3'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      expect.objectContaining({
        id: '7',
        matched: true,
        amount: -12.34,
        date: '2026-09-20T12:30:00Z',
        description: 'Coffee shop',
      }),
    ]);
  });

  it('exposes transport failures as query errors rather than an empty success', async () => {
    getBankFeedsMock.mockRejectedValue(new Error('feed unavailable'));

    const { result } = renderHook(() => useBankTransactions(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.data).toBeUndefined();
  });

  it('rejects a response without a transactions array', async () => {
    getBankFeedsMock.mockResolvedValue({ total: 0 });

    const { result } = renderHook(() => useBankTransactions(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.data).toBeUndefined();
  });
});
