import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/bookService', () => ({
  bookService: {
    getCustomers: vi.fn(),
    getCustomer: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useCustomers, useCustomer } from './useCustomers';
import { bookService } from '../../shared/services/bookService';

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
    (bookService.getCustomers as any).mockResolvedValue(data);
    const { result } = renderHook(() => useCustomers(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes filters to the service', async () => {
    (bookService.getCustomers as any).mockResolvedValue([]);
    const filters = { search: 'alice' };
    const { result } = renderHook(() => useCustomers(filters), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(bookService.getCustomers).toHaveBeenCalledWith(filters);
  });

  it('starts in loading state', () => {
    (bookService.getCustomers as any).mockReturnValue(new Promise(() => {}));
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
    const cust = { id: 3, name: 'Bob' };
    (bookService.getCustomer as any).mockResolvedValue(cust);
    const { result } = renderHook(() => useCustomer(3), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(cust);
  });

  it('does not fetch when id is undefined', () => {
    const { result } = renderHook(() => useCustomer(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
    expect(bookService.getCustomer).not.toHaveBeenCalled();
  });
});
