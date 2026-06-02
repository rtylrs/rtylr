import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  usePurchaseOrders,
  useCreatePurchaseOrder,
  useUpdatePurchaseOrderStatus,
  useApprovePurchaseOrder,
  useSendPurchaseOrder,
} from './usePurchaseOrders';

const mockErpApi = {
  listPurchaseOrders: vi.fn(),
  createPurchaseOrder: vi.fn(),
  updatePurchaseOrderStatus: vi.fn(),
  approvePurchaseOrder: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listPurchaseOrders: (...args: unknown[]) => mockErpApi.listPurchaseOrders(...args),
    createPurchaseOrder: (...args: unknown[]) => mockErpApi.createPurchaseOrder(...args),
    updatePurchaseOrderStatus: (...args: unknown[]) =>
      mockErpApi.updatePurchaseOrderStatus(...args),
    approvePurchaseOrder: (...args: unknown[]) => mockErpApi.approvePurchaseOrder(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

// Backend (erp.v1.PurchaseOrder) shape, camelCased by @voxire/api.
const sampleOrders = [
  {
    id: 1,
    supplierId: 10,
    poNumber: 'PO-001',
    status: 'draft',
    subtotal: 450,
    taxAmount: 50,
    total: 500,
    expectedDelivery: '2024-01-10',
    notes: '',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
    items: [
      {
        id: 1,
        purchaseOrderId: 1,
        productId: 5,
        quantityOrdered: 2,
        quantityReceived: 0,
        unitCost: 100,
        totalCost: 200,
      },
      {
        id: 2,
        purchaseOrderId: 1,
        productId: 6,
        quantityOrdered: 1,
        quantityReceived: 0,
        unitCost: 250,
        totalCost: 250,
      },
    ],
  },
];

describe('usePurchaseOrders', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches purchase orders list', async () => {
    mockErpApi.listPurchaseOrders.mockResolvedValue({
      purchaseOrders: sampleOrders,
      hasMore: false,
    });
    const { result } = renderHook(() => usePurchaseOrders(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.orders).toHaveLength(1);
  });

  it('adapts the backend shape onto the UI shape', async () => {
    mockErpApi.listPurchaseOrders.mockResolvedValue({
      purchaseOrders: sampleOrders,
      hasMore: false,
    });
    const { result } = renderHook(() => usePurchaseOrders(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const order = result.current.data?.orders[0];
    expect(order?.items).toBe(2);
    expect(order?.supplier).toBe('Supplier #10');
    expect(order?.expectedDate).toBe('2024-01-10');
    expect(order?.total).toBe('500');
    expect(order?.status).toBe('Draft');
  });

  it('uses default pagination params', async () => {
    mockErpApi.listPurchaseOrders.mockResolvedValue({ purchaseOrders: [], hasMore: false });
    const { result } = renderHook(() => usePurchaseOrders(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockErpApi.listPurchaseOrders).toHaveBeenCalledWith({ page: 1, page_size: 50 });
  });

  it('passes custom pagination params', async () => {
    mockErpApi.listPurchaseOrders.mockResolvedValue({ purchaseOrders: [], hasMore: false });
    const { result } = renderHook(() => usePurchaseOrders({ page: 2, pageSize: 25 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockErpApi.listPurchaseOrders).toHaveBeenCalledWith({ page: 2, page_size: 25 });
  });

  it('handles API error', async () => {
    mockErpApi.listPurchaseOrders.mockRejectedValue(new Error('network error'));
    const { result } = renderHook(() => usePurchaseOrders(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useCreatePurchaseOrder', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createPurchaseOrder with data', async () => {
    mockErpApi.createPurchaseOrder.mockResolvedValue({ id: 2, supplierId: 10 });
    const { result } = renderHook(() => useCreatePurchaseOrder(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ supplierId: 10, items: [] } as unknown as Parameters<
        typeof result.current.mutateAsync
      >[0]);
    });
    expect(mockErpApi.createPurchaseOrder).toHaveBeenCalledWith({ supplierId: 10, items: [] });
  });
});

describe('useUpdatePurchaseOrderStatus', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updatePurchaseOrderStatus with id and lowercase status', async () => {
    mockErpApi.updatePurchaseOrderStatus.mockResolvedValue({ id: 1, status: 'sent' });
    const { result } = renderHook(() => useUpdatePurchaseOrderStatus(), {
      wrapper: createWrapper(),
    });
    await act(async () => {
      await result.current.mutateAsync({ id: 1, status: 'sent' });
    });
    expect(mockErpApi.updatePurchaseOrderStatus).toHaveBeenCalledWith(1, 'sent');
  });
});

describe('useApprovePurchaseOrder', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls approvePurchaseOrder with id', async () => {
    mockErpApi.approvePurchaseOrder.mockResolvedValue({ id: 1, status: 'Approved' });
    const { result } = renderHook(() => useApprovePurchaseOrder(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync(1);
    });
    expect(mockErpApi.approvePurchaseOrder).toHaveBeenCalledWith(1);
  });
});

describe('useSendPurchaseOrder', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updatePurchaseOrderStatus with sent status', async () => {
    mockErpApi.updatePurchaseOrderStatus.mockResolvedValue({ id: 1, status: 'Sent' });
    const { result } = renderHook(() => useSendPurchaseOrder(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync(1);
    });
    expect(mockErpApi.updatePurchaseOrderStatus).toHaveBeenCalledWith(1, 'sent');
  });
});
