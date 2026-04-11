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

import { storeApi } from './storeService';
import { safe } from '../lib/api';

describe('storeApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Dashboard
  it('getDashboardStats calls GET /store/dashboard', () => {
    storeApi.getDashboardStats();
    expect(safe.get).toHaveBeenCalledWith('/store/dashboard');
  });

  // Products
  it('listProducts calls GET /store/products with params', () => {
    const params = { search: 'shirt', status: 'active' };
    storeApi.listProducts(params);
    expect(safe.get).toHaveBeenCalledWith('/store/products', { params });
  });

  it('getProduct calls GET /store/products/:id', () => {
    storeApi.getProduct(5);
    expect(safe.get).toHaveBeenCalledWith('/store/products/5');
  });

  it('createProduct calls POST /store/products', () => {
    const data = { name: 'Shirt', sku: 'SH1', price: 29.99, stock: 10 };
    storeApi.createProduct(data as any);
    expect(safe.post).toHaveBeenCalledWith('/store/products', data);
  });

  it('updateProduct calls POST /store/products', () => {
    const data = { id: 1, name: 'Updated Shirt' };
    storeApi.updateProduct(data as any);
    expect(safe.post).toHaveBeenCalledWith('/store/products', data);
  });

  it('deleteProduct calls POST /store/products with delete action', () => {
    storeApi.deleteProduct(3);
    expect(safe.post).toHaveBeenCalledWith('/store/products', { id: 3, _action: 'delete' });
  });

  // Orders
  it('listOrders calls GET /store/orders with params', () => {
    const params = { status: 'pending', limit: 10 };
    storeApi.listOrders(params);
    expect(safe.get).toHaveBeenCalledWith('/store/orders', { params });
  });

  it('getOrder calls GET /store/orders/:id', () => {
    storeApi.getOrder(7);
    expect(safe.get).toHaveBeenCalledWith('/store/orders/7');
  });

  it('updateOrderStatus calls POST /store/orders/status', () => {
    storeApi.updateOrderStatus(1, 'shipped');
    expect(safe.post).toHaveBeenCalledWith('/store/orders/status', { id: 1, status: 'shipped' });
  });

  // Categories
  it('listCategories calls GET /store/categories', () => {
    storeApi.listCategories();
    expect(safe.get).toHaveBeenCalledWith('/store/categories');
  });

  it('createCategory calls POST /store/categories', () => {
    const data = { name: 'Tops' };
    storeApi.createCategory(data);
    expect(safe.post).toHaveBeenCalledWith('/store/categories', data);
  });

  it('deleteCategory calls POST /store/categories with delete action', () => {
    storeApi.deleteCategory(2);
    expect(safe.post).toHaveBeenCalledWith('/store/categories', { id: 2, _action: 'delete' });
  });

  // Shipping
  it('listShippingZones calls GET /store/shipping/zones', () => {
    storeApi.listShippingZones();
    expect(safe.get).toHaveBeenCalledWith('/store/shipping/zones');
  });

  it('createShippingZone calls POST /store/shipping/zones', () => {
    const data = { name: 'US', regions: ['NY', 'CA'] };
    storeApi.createShippingZone(data);
    expect(safe.post).toHaveBeenCalledWith('/store/shipping/zones', data);
  });

  it('listShippingMethods calls GET /store/shipping/methods', () => {
    storeApi.listShippingMethods();
    expect(safe.get).toHaveBeenCalledWith('/store/shipping/methods');
  });

  // Payments
  it('listPaymentMethods calls GET /store/payments/methods', () => {
    storeApi.listPaymentMethods();
    expect(safe.get).toHaveBeenCalledWith('/store/payments/methods');
  });

  it('updatePaymentMethod calls POST /store/payments/methods', () => {
    const data = { id: 1, enabled: true };
    storeApi.updatePaymentMethod(data as any);
    expect(safe.post).toHaveBeenCalledWith('/store/payments/methods', data);
  });

  // Storefront
  it('getStorefrontConfig calls GET /store/storefront', () => {
    storeApi.getStorefrontConfig();
    expect(safe.get).toHaveBeenCalledWith('/store/storefront');
  });

  it('updateStorefrontConfig calls POST /store/storefront', () => {
    const data = { primaryColor: '#ff0000' };
    storeApi.updateStorefrontConfig(data);
    expect(safe.post).toHaveBeenCalledWith('/store/storefront', data);
  });

  // Settings
  it('getSettings calls GET /store/settings', () => {
    storeApi.getSettings();
    expect(safe.get).toHaveBeenCalledWith('/store/settings');
  });

  it('updateSettings calls POST /store/settings', () => {
    const data = { storeName: 'My Store' };
    storeApi.updateSettings(data);
    expect(safe.post).toHaveBeenCalledWith('/store/settings', data);
  });
});
