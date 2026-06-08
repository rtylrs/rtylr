import { safe } from '../lib/api';

export interface ProductPayload {
  name: string;
  sku?: string;
  categoryId?: number;
  price?: number;
  cost?: number;
  stock?: number;
  unit?: string;
  description?: string;
  [key: string]: unknown;
}

export interface SupplierPayload {
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  [key: string]: unknown;
}

export interface InventoryAdjustPayload {
  productId: number;
  quantity: number;
  reason?: string;
  [key: string]: unknown;
}

export interface PurchaseOrderPayload {
  supplierId: number;
  items: Array<{ productId: number; quantity: number; unitPrice: number }>;
  expectedDate?: string;
  notes?: string;
  [key: string]: unknown;
}

export const erpService = {
  // Products
  getProducts: (params?: {
    search?: string;
    categoryId?: number;
    limit?: number;
    offset?: number;
  }) => safe.get('/erp/products', { params }),
  createProduct: (data: ProductPayload) => safe.post('/erp/products', data),
  updateProduct: (id: number, data: Partial<ProductPayload>) =>
    safe.put(`/erp/products/${id}`, data),
  deleteProduct: (id: number) => safe.delete(`/erp/products/${id}`),

  // Suppliers
  getSuppliers: () => safe.get('/erp/suppliers'),
  createSupplier: (data: SupplierPayload) => safe.post('/erp/suppliers', data),
  updateSupplier: (id: number, data: Partial<SupplierPayload>) =>
    safe.put(`/erp/suppliers/${id}`, data),

  // Inventory
  getInventoryReport: () => safe.get('/erp/inventory/report'),
  getLowStock: () => safe.get('/erp/inventory/low-stock'),
  adjustInventory: (data: InventoryAdjustPayload) => safe.post('/erp/inventory/adjust', data),

  // Purchase Orders
  getPurchaseOrders: () => safe.get('/erp/purchase-orders'),
  createPurchaseOrder: (data: PurchaseOrderPayload) => safe.post('/erp/purchase-orders', data),
  updatePurchaseOrderStatus: (id: number, status: string) =>
    safe.put('/erp/purchase-orders/status', { id, status }),
};
