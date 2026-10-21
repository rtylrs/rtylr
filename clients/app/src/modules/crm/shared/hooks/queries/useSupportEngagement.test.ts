import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useSupportEngagement } from './useSupportEngagement';

const listCustomers = vi.hoisted(() => vi.fn());

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ listCustomers }),
}));

vi.mock('@voxire/voxie/support/customer/v1/client', () => ({
  CustomerServiceClient: class {},
}));

function createWrapper(client: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
}

function supportCustomers(count: number, offset = 0) {
  return Array.from({ length: count }, (_, index) => ({
    id: String(index + offset + 1),
    email: `customer${index + offset + 1}@example.com`,
    totalTickets: 1,
  }));
}

describe('useSupportEngagement', () => {
  beforeEach(() => {
    listCustomers.mockReset();
  });

  it('continues past a capped first page to find an older matching customer', async () => {
    listCustomers
      .mockResolvedValueOnce({
        customers: supportCustomers(20),
        total: 21,
        page: 1,
        pageSize: 100,
      })
      .mockResolvedValueOnce({
        customers: [
          {
            id: '21',
            email: 'TARGET@example.com',
            totalTickets: 4,
            lastTicketAt: '2026-09-23T10:00:00Z',
          },
        ],
        total: 21,
        page: 2,
        pageSize: 100,
      });

    const { result } = renderHook(() => useSupportEngagement('target@example.com'), {
      wrapper: createWrapper(createClient()),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({
      totalTickets: 4,
      lastTicketAt: '2026-09-23T10:00:00Z',
    });
    expect(listCustomers).toHaveBeenNthCalledWith(1, { page: 1, pageSize: 100 });
    expect(listCustomers).toHaveBeenNthCalledWith(2, { page: 2, pageSize: 100 });
  });

  it('reads all advertised pages before caching a no-match result', async () => {
    listCustomers
      .mockResolvedValueOnce({
        customers: supportCustomers(20),
        total: 40,
        page: 1,
        pageSize: 100,
      })
      .mockResolvedValueOnce({
        customers: supportCustomers(20, 20),
        total: 40,
        page: 2,
        pageSize: 100,
      });

    const { result } = renderHook(() => useSupportEngagement('missing@example.com'), {
      wrapper: createWrapper(createClient()),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeNull();
    expect(listCustomers).toHaveBeenCalledTimes(2);
  });
});
