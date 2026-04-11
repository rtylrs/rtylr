import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/supportService', () => ({
  supportService: {
    getCustomers: vi.fn(),
    getCustomer: vi.fn(),
    getCustomerTickets: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useCustomers, useCustomer, useCustomerTickets } from './useCustomers';
import { supportService } from '../../shared/services/supportService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useCustomers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches customers on mount', async () => {
    const data = [{ id: 1, name: 'Alice' }];
    (supportService.getCustomers as any).mockResolvedValue(data);
    const { result } = renderHook(() => useCustomers(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes params to the service', async () => {
    (supportService.getCustomers as any).mockResolvedValue([]);
    const params = { search: 'bob' };
    const { result } = renderHook(() => useCustomers(params), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(supportService.getCustomers).toHaveBeenCalledWith(params);
  });

  it('starts in loading state', () => {
    (supportService.getCustomers as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useCustomers(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useCustomer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches a single customer by id', async () => {
    const cust = { id: 3, name: 'Charlie' };
    (supportService.getCustomer as any).mockResolvedValue(cust);
    const { result } = renderHook(() => useCustomer(3), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(cust);
  });
});

describe('useCustomerTickets', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches tickets for a customer', async () => {
    const tickets = [{ id: 10, subject: 'Issue' }];
    (supportService.getCustomerTickets as any).mockResolvedValue(tickets);
    const { result } = renderHook(() => useCustomerTickets(3), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(tickets);
    expect(supportService.getCustomerTickets).toHaveBeenCalledWith(3);
  });

  it('does not fetch when customerId is falsy', () => {
    const { result } = renderHook(() => useCustomerTickets(0), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
  });
});
