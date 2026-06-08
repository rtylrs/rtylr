import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
  },
}));

import { erpService } from './erpService';
import { safe } from '../lib/api';

describe('erpService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Products
  it('getProducts calls GET /erp/products with params', () => {
    const params = { search: 'bolt' };
    erpService.getProducts(params);
    expect(safe.get).toHaveBeenCalledWith('/erp/products', { params });
  });

  it('createProduct calls POST /erp/products', () => {
    const data = { name: 'Bolt' };
    erpService.createProduct(data);
    expect(safe.post).toHaveBeenCalledWith('/erp/products', data);
  });

  it('updateProduct calls PUT /erp/products/:id', () => {
    erpService.updateProduct(1, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/erp/products/1', { name: 'Updated' });
  });

  it('deleteProduct calls DELETE /erp/products/:id', () => {
    erpService.deleteProduct(1);
    expect(safe.delete).toHaveBeenCalledWith('/erp/products/1');
  });

  // Suppliers
  it('getSuppliers calls GET /erp/suppliers', () => {
    erpService.getSuppliers();
    expect(safe.get).toHaveBeenCalledWith('/erp/suppliers');
  });

  it('createSupplier calls POST /erp/suppliers', () => {
    const data = { name: 'Acme' };
    erpService.createSupplier(data);
    expect(safe.post).toHaveBeenCalledWith('/erp/suppliers', data);
  });

  it('updateSupplier calls PUT /erp/suppliers/:id', () => {
    erpService.updateSupplier(2, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/erp/suppliers/2', { name: 'Updated' });
  });

  // Inventory
  it('getInventoryReport calls GET /erp/inventory/report', () => {
    erpService.getInventoryReport();
    expect(safe.get).toHaveBeenCalledWith('/erp/inventory/report');
  });

  it('getLowStock calls GET /erp/inventory/low-stock', () => {
    erpService.getLowStock();
    expect(safe.get).toHaveBeenCalledWith('/erp/inventory/low-stock');
  });

  it('adjustInventory calls POST /erp/inventory/adjust', () => {
    const data = { productId: 1, qty: 10 };
    erpService.adjustInventory(data);
    expect(safe.post).toHaveBeenCalledWith('/erp/inventory/adjust', data);
  });

  // Purchase Orders
  it('getPurchaseOrders calls GET /erp/purchase-orders', () => {
    erpService.getPurchaseOrders();
    expect(safe.get).toHaveBeenCalledWith('/erp/purchase-orders');
  });

  it('createPurchaseOrder calls POST /erp/purchase-orders', () => {
    const data = { supplierId: 1 };
    erpService.createPurchaseOrder(data);
    expect(safe.post).toHaveBeenCalledWith('/erp/purchase-orders', data);
  });

  it('updatePurchaseOrderStatus calls PUT /erp/purchase-orders/status with id', () => {
    erpService.updatePurchaseOrderStatus(1, 'Approved');
    expect(safe.put).toHaveBeenCalledWith('/erp/purchase-orders/status', {
      id: 1,
      status: 'Approved',
    });
  });
});
