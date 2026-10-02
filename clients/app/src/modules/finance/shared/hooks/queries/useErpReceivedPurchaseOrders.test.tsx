// @vitest-environment happy-dom
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockClient = vi.hoisted(() => ({
  listPurchaseOrders: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
}));
vi.mock('@voxire/voxie/erp/purchase_order/v1/client', () => ({
  PurchaseOrderServiceClient: {},
}));

import { useErpReceivedPurchaseOrders } from './useErpReceivedPurchaseOrders';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
};

describe('useErpReceivedPurchaseOrders', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('keeps status-filtered purchase orders even though quantityReceived is only an item field', async () => {
    mockClient.listPurchaseOrders.mockResolvedValue({
      purchaseOrders: [
        {
          id: 'po-1',
          poNumber: 'PO-001',
          total: 480,
          status: 'received',
          items: [{ quantityReceived: 3 }],
        },
      ],
    });

    const { result } = renderHook(() => useErpReceivedPurchaseOrders(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.listPurchaseOrders).toHaveBeenCalledWith({
      status: 'received',
      page: 1,
      pageSize: 100,
    });
    expect(result.current.data).toEqual([
      { id: 'po-1', supplierId: undefined, poNumber: 'PO-001', total: 480 },
    ]);
  });
});
