import api from '../lib/api';

export const erpService = {
  // Products
  getProducts: (params?: { search?: string; categoryId?: number; limit?: number; offset?: number }) =>
    api.get('/erp/products', { params }),
  createProduct: (data: any) => api.post('/erp/products', data),
  updateProduct: (id: number, data: any) => api.put(`/erp/products/${id}`, data),
  deleteProduct: (id: number) => api.delete(`/erp/products/${id}`),

  // Suppliers
  getSuppliers: () => api.get('/erp/suppliers'),
  createSupplier: (data: any) => api.post('/erp/suppliers', data),
  updateSupplier: (id: number, data: any) => api.put(`/erp/suppliers/${id}`, data),

  // Inventory
  getInventoryReport: () => api.get('/erp/inventory/report'),
  getLowStock: () => api.get('/erp/inventory/low-stock'),
  adjustInventory: (data: any) => api.post('/erp/inventory/adjust', data),

  // Purchase Orders
  getPurchaseOrders: () => api.get('/erp/purchase-orders'),
  createPurchaseOrder: (data: any) => api.post('/erp/purchase-orders', data),
  updatePurchaseOrderStatus: (id: number, status: string) => api.put(`/erp/purchase-orders/${id}/status`, { status }),
};
