import { safe } from '../lib/api';

export const erpService = {
  // Products
  getProducts: (params?: { search?: string; categoryId?: number; limit?: number; offset?: number }) =>
    safe.get('/erp/products', { params }),
  createProduct: (data: any) => safe.post('/erp/products', data),
  updateProduct: (id: number, data: any) => safe.put(`/erp/products/${id}`, data),
  deleteProduct: (id: number) => safe.delete(`/erp/products/${id}`),

  // Suppliers
  getSuppliers: () => safe.get('/erp/suppliers'),
  createSupplier: (data: any) => safe.post('/erp/suppliers', data),
  updateSupplier: (id: number, data: any) => safe.put(`/erp/suppliers/${id}`, data),

  // Inventory
  getInventoryReport: () => safe.get('/erp/inventory/report'),
  getLowStock: () => safe.get('/erp/inventory/low-stock'),
  adjustInventory: (data: any) => safe.post('/erp/inventory/adjust', data),

  // Purchase Orders
  getPurchaseOrders: () => safe.get('/erp/purchase-orders'),
  createPurchaseOrder: (data: any) => safe.post('/erp/purchase-orders', data),
  updatePurchaseOrderStatus: (id: number, status: string) => safe.put(`/erp/purchase-orders/${id}/status`, { status }),
};
