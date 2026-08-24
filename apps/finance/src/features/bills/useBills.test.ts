// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockClient = vi.hoisted(() => ({
  listBills: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
}));

import { useBills } from './useBills';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

const baseBill = {
  id: '1',
  ownerId: '1',
  supplierId: '1',
  supplierName: 'Acme Corp',
  amount: 500,
  currency: 'USD',
  dueDate: '2024-06-01',
  status: 'unpaid',
  description: 'ACME',
  createdAt: '',
  updatedAt: '',
};

describe('useBills', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches bills on mount', async () => {
    mockClient.listBills.mockResolvedValue({ bills: [baseBill] });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.listBills).toHaveBeenCalledWith({});
    expect(result.current.data).toHaveLength(1);
  });

  it('starts in loading state', () => {
    mockClient.listBills.mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('returns empty array when bills key is empty', async () => {
    mockClient.listBills.mockResolvedValue({ bills: [] });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('shows the resolved supplier name as the vendor, not the description', async () => {
    // Deliberately different values. When description doubles as the vendor
    // name (the create form falls back to it), both mappings agree and the
    // assertion cannot tell them apart.
    mockClient.listBills.mockResolvedValue({
      bills: [{ ...baseBill, supplierName: 'Acme Corp', description: 'March hosting' }],
    });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.vendor).toBe('Acme Corp');
  });

  it('keeps the description separate from the vendor', async () => {
    // mark-as-paid round-trips description, so conflating the two would send
    // the vendor name back as the description.
    mockClient.listBills.mockResolvedValue({
      bills: [{ ...baseBill, supplierName: 'Acme Corp', description: 'March hosting' }],
    });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.description).toBe('March hosting');
  });

  it('keeps bill id as a string', async () => {
    mockClient.listBills.mockResolvedValue({ bills: [{ ...baseBill, id: '99' }] });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.id).toBe('99');
  });
});
