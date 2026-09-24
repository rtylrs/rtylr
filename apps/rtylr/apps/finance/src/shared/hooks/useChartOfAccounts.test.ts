// @vitest-environment happy-dom
import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const clients = vi.hoisted(() => ({
  account: { listAccounts: vi.fn() },
  journal: { listJournalEntries: vi.fn() },
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (client: { name: string }) =>
    client.name === 'AccountServiceClient' ? clients.account : clients.journal,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

import { useJournal } from './useChartOfAccounts';

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(QueryClientProvider, { client: queryClient }, children);

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useJournal', () => {
  it('uses the server hasNext flag and maps authoritative account names', async () => {
    clients.journal.listJournalEntries.mockResolvedValue({
      entries: [
        {
          id: '1',
          date: '2026-09-20',
          description: 'Balanced entry',
          lines: [
            { accountId: '10', accountName: 'Cash', debit: 100, credit: 0 },
            { accountId: '20', accountName: 'Revenue', debit: 0, credit: 100 },
          ],
        },
      ],
      pagination: { page: 1, limit: 50, hasNext: true, hasPrev: false },
    });

    const { result } = renderHook(() => useJournal(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.hasMore).toBe(true);
    expect(result.current.data?.entries[0]?.lines).toEqual([
      { accountId: '10', account: 'Cash', debit: 100, credit: 0, description: undefined },
      { accountId: '20', account: 'Revenue', debit: 0, credit: 100, description: undefined },
    ]);
  });
});
