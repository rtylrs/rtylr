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

import { erpApi } from './erpApi';
import { safe } from '../lib/api';

describe('erpApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Products
  it('listProducts calls GET /erp/products with params', () => {
    const params = { search: 'widget', page_size: 10 };
    erpApi.listProducts(params);
    expect(safe.get).toHaveBeenCalledWith('/erp/products', { params });
  });

  it('getProduct calls GET /erp/products/:id', () => {
    erpApi.getProduct(5);
    expect(safe.get).toHaveBeenCalledWith('/erp/products/5');
  });

  it('createProduct calls POST /erp/products', () => {
    const data = {
      name: 'Widget',
      sku: 'W1',
      category: 'Parts',
      stock: 100,
      lowStock: 10,
      costPrice: 5,
      salePrice: 10,
    };
    erpApi.createProduct(data);
    expect(safe.post).toHaveBeenCalledWith('/erp/products', data);
  });

  it('deleteProduct calls DELETE /erp/products with id', () => {
    erpApi.deleteProduct(3);
    expect(safe.delete).toHaveBeenCalledWith('/erp/products', { data: { id: 3 } });
  });

  it('lookupBarcode calls GET /erp/products/barcode/:code', () => {
    erpApi.lookupBarcode('ABC123');
    expect(safe.get).toHaveBeenCalledWith('/erp/products/barcode/ABC123');
  });

  // Categories & Brands
  it('listCategories calls GET /erp/categories', () => {
    erpApi.listCategories();
    expect(safe.get).toHaveBeenCalledWith('/erp/categories');
  });

  it('listBrands calls GET /erp/brands', () => {
    erpApi.listBrands();
    expect(safe.get).toHaveBeenCalledWith('/erp/brands');
  });

  // Warehouses
  it('listWarehouses calls GET /erp/warehouses', () => {
    erpApi.listWarehouses();
    expect(safe.get).toHaveBeenCalledWith('/erp/warehouses');
  });

  it('createWarehouse calls POST /erp/warehouses', () => {
    const data = { name: 'Main' };
    erpApi.createWarehouse(data);
    expect(safe.post).toHaveBeenCalledWith('/erp/warehouses', data);
  });

  // Stock transfers & takes
  it('listStockTransfers calls GET /erp/stock-transfers', () => {
    erpApi.listStockTransfers();
    expect(safe.get).toHaveBeenCalledWith('/erp/stock-transfers', { params: undefined });
  });

  it('listStockTakes calls GET /erp/stock-takes', () => {
    erpApi.listStockTakes();
    expect(safe.get).toHaveBeenCalledWith('/erp/stock-takes', { params: undefined });
  });

  // Suppliers
  it('listSuppliers calls GET /erp/suppliers', () => {
    erpApi.listSuppliers();
    expect(safe.get).toHaveBeenCalledWith('/erp/suppliers', { params: undefined });
  });

  it('createSupplier calls POST /erp/suppliers', () => {
    const data = {
      name: 'Acme',
      contactName: 'Bob',
      email: 'bob@acme.com',
      phone: '555',
      categories: ['Parts'],
    };
    erpApi.createSupplier(data);
    expect(safe.post).toHaveBeenCalledWith('/erp/suppliers', {
      name: 'Acme',
      email: 'bob@acme.com',
      phone: '555',
      isActive: undefined,
    });
  });

  // Purchase Orders
  it('listPurchaseOrders calls GET /erp/purchase-orders', () => {
    erpApi.listPurchaseOrders();
    expect(safe.get).toHaveBeenCalledWith('/erp/purchase-orders', { params: undefined });
  });

  it('updatePurchaseOrderStatus calls PUT /erp/purchase-orders/status', () => {
    erpApi.updatePurchaseOrderStatus(1, 'Approved');
    expect(safe.put).toHaveBeenCalledWith('/erp/purchase-orders/status', {
      id: 1,
      status: 'Approved',
    });
  });

  it('approvePurchaseOrder calls POST /erp/purchase-orders/:id/approve', () => {
    erpApi.approvePurchaseOrder(2);
    expect(safe.post).toHaveBeenCalledWith('/erp/purchase-orders/2/approve', {});
  });

  // Inventory
  it('getInventoryReport calls GET /erp/inventory/report', () => {
    erpApi.getInventoryReport();
    expect(safe.get).toHaveBeenCalledWith('/erp/inventory/report');
  });

  it('getLowStock calls GET /erp/inventory/low-stock', () => {
    erpApi.getLowStock();
    expect(safe.get).toHaveBeenCalledWith('/erp/inventory/low-stock');
  });

  it('adjustInventory calls POST /erp/inventory/adjust', () => {
    const data = { productId: 1, quantity: 5 };
    erpApi.adjustInventory(data);
    expect(safe.post).toHaveBeenCalledWith('/erp/inventory/adjust', {
      productId: 1,
      change: 5,
      type: 'adjustment',
    });
  });

  it('getValuation calls GET /erp/inventory/valuation with method param', () => {
    erpApi.getValuation('FIFO');
    expect(safe.get).toHaveBeenCalledWith('/erp/inventory/valuation', {
      params: { method: 'FIFO' },
    });
  });

  // Misc
  it('getReorderSuggestions calls GET /erp/reorder-suggestions', () => {
    erpApi.getReorderSuggestions();
    expect(safe.get).toHaveBeenCalledWith('/erp/reorder-suggestions');
  });

  it('getExpiryItems calls GET /erp/expiry', () => {
    erpApi.getExpiryItems();
    expect(safe.get).toHaveBeenCalledWith('/erp/expiry');
  });

  it('listPriceLists calls GET /erp/price-lists', () => {
    erpApi.listPriceLists();
    expect(safe.get).toHaveBeenCalledWith('/erp/price-lists');
  });

  it('listVariants calls GET /erp/variants with productId', () => {
    erpApi.listVariants(10);
    expect(safe.get).toHaveBeenCalledWith('/erp/variants', { params: { productId: 10 } });
  });

  it('listActivity calls GET /erp/activity with limit', () => {
    erpApi.listActivity(50);
    expect(safe.get).toHaveBeenCalledWith('/erp/activity', { params: { limit: 50 } });
  });

  it('getDashboardStats calls GET /erp/dashboard', () => {
    erpApi.getDashboardStats();
    expect(safe.get).toHaveBeenCalledWith('/erp/dashboard');
  });
});
