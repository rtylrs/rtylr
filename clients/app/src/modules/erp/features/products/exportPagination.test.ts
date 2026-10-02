import { beforeEach, describe, expect, it, vi } from 'vitest';

const productClient = vi.hoisted(() => ({ listProducts: vi.fn() }));
const purchaseOrderClient = vi.hoisted(() => ({ listPurchaseOrders: vi.fn() }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (Client: { name?: string }) =>
    Client.name === 'ProductServiceClient' ? productClient : purchaseOrderClient,
}));

import { fetchAllProducts } from './useProducts';
import { fetchAllPurchaseOrders } from '../purchasing/usePurchaseOrders';

describe('ERP export pagination', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches every product page with the active search filters', async () => {
    productClient.listProducts
      .mockResolvedValueOnce({
        products: Array.from({ length: 100 }, (_, index) => ({ id: String(index) })),
        hasMore: true,
      })
      .mockResolvedValueOnce({ products: [{ id: '100' }], hasMore: false });

    const products = await fetchAllProducts({ search: '  coffee ', categoryId: '7' });

    expect(products).toHaveLength(101);
    expect(productClient.listProducts).toHaveBeenNthCalledWith(1, {
      search: 'coffee',
      categoryId: '7',
      limit: 100,
      offset: 0,
    });
    expect(productClient.listProducts).toHaveBeenNthCalledWith(2, {
      search: 'coffee',
      categoryId: '7',
      limit: 100,
      offset: 100,
    });
  });

  it('fetches every purchase-order page', async () => {
    purchaseOrderClient.listPurchaseOrders
      .mockResolvedValueOnce({ purchaseOrders: [{ id: 'po-1' }], hasMore: true })
      .mockResolvedValueOnce({ purchaseOrders: [{ id: 'po-2' }], hasMore: false });

    const purchaseOrders = await fetchAllPurchaseOrders({ status: 'draft' });

    expect(purchaseOrders.map((order) => order.id)).toEqual(['po-1', 'po-2']);
    expect(purchaseOrderClient.listPurchaseOrders).toHaveBeenNthCalledWith(1, {
      status: 'draft',
      page: 1,
      pageSize: 100,
    });
    expect(purchaseOrderClient.listPurchaseOrders).toHaveBeenNthCalledWith(2, {
      status: 'draft',
      page: 2,
      pageSize: 100,
    });
  });
});
