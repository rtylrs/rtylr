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

import { BILLS_PAGE_SIZE, useBills } from './useBills';

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

  it('fetches bills on mount with an explicit page', async () => {
    mockClient.listBills.mockResolvedValue({ bills: [baseBill], hasMore: false });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.listBills).toHaveBeenCalledWith({ page: 1, pageSize: BILLS_PAGE_SIZE });
    expect(result.current.data).toHaveLength(1);
  });

  it('follows has_more so the complete set is available for metrics and actions', async () => {
    const firstPage = Array.from({ length: BILLS_PAGE_SIZE }, (_, index) => ({
      ...baseBill,
      id: String(index + 1),
    }));
    const secondPage = { ...baseBill, id: 'last', amount: 750 };

    mockClient.listBills.mockImplementation(async ({ page }: { page: number }) =>
      page === 1
        ? { bills: firstPage, hasMore: true, page: 1, pageSize: BILLS_PAGE_SIZE }
        : { bills: [secondPage], hasMore: false, page: 2, pageSize: BILLS_PAGE_SIZE }
    );

    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(BILLS_PAGE_SIZE + 1);
    expect(result.current.data?.[BILLS_PAGE_SIZE]?.id).toBe('last');
    expect(mockClient.listBills).toHaveBeenNthCalledWith(1, {
      page: 1,
      pageSize: BILLS_PAGE_SIZE,
    });
    expect(mockClient.listBills).toHaveBeenNthCalledWith(2, {
      page: 2,
      pageSize: BILLS_PAGE_SIZE,
    });
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
      hasMore: false,
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
      hasMore: false,
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
