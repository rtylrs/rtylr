import api from '../lib/api';

// ─── Types ───────────────────────────────────────────────────────────────────

export interface Product {
  id: number;
  name: string;
  sku: string;
  category: string;
  stock: number;
  lowStock: number;
  costPrice: number;
  salePrice: number;
  status: 'In Stock' | 'Low Stock' | 'Out of Stock';
}

export interface CreateProductInput {
  name: string;
  sku: string;
  category: string;
  stock: number;
  lowStock: number;
  costPrice: number;
  salePrice: number;
}

export interface UpdateProductInput extends Partial<CreateProductInput> {
  id: number;
}

export interface Supplier {
  id: number;
  name: string;
  contactName: string;
  email: string;
  phone: string;
  productCount: number;
  categories: string[];
  status: 'Active' | 'Inactive';
}

export interface CreateSupplierInput {
  name: string;
  contactName: string;
  email: string;
  phone: string;
  categories: string[];
}

export interface UpdateSupplierInput extends Partial<CreateSupplierInput> {
  id: number;
}

export interface PurchaseOrder {
  id: number;
  supplier: string;
  supplierId: number;
  date: string;
  expectedDate: string;
  items: number;
  total: string;
  status: 'Draft' | 'Sent' | 'Received';
}

export interface LowStockItem {
  name: string;
  sku: string;
  stock: number;
  minStock: number;
  unit: string;
}

export interface InventoryReportItem {
  productId: number;
  name: string;
  sku: string;
  currentStock: number;
  value: number;
}

export interface DashboardStats {
  totalProducts: number;
  lowStockCount: number;
  pendingOrders: number;
  totalSuppliers: number;
  lowStockItems: LowStockItem[];
  recentOrders: PurchaseOrder[];
}

// ─── API ─────────────────────────────────────────────────────────────────────

export const erpApi = {
  // Products
  listProducts: (params?: { search?: string; categoryId?: number; limit?: number; offset?: number }) =>
    api.get<Product[]>('/erp/products', { params }),
  createProduct: (data: CreateProductInput) =>
    api.post<Product>('/erp/products', data),
  updateProduct: (data: UpdateProductInput) =>
    api.post<Product>('/erp/products', data),
  deleteProduct: (id: number) =>
    api.post<{ success: boolean }>('/erp/products', { id, _action: 'delete' }),

  // Suppliers
  listSuppliers: () =>
    api.get<Supplier[]>('/erp/suppliers'),
  createSupplier: (data: CreateSupplierInput) =>
    api.post<Supplier>('/erp/suppliers', data),
  updateSupplier: (data: UpdateSupplierInput) =>
    api.post<Supplier>('/erp/suppliers', data),

  // Purchase Orders
  listPurchaseOrders: () =>
    api.get<PurchaseOrder[]>('/erp/purchase-orders'),
  createPurchaseOrder: (data: any) =>
    api.post<PurchaseOrder>('/erp/purchase-orders', data),
  updatePurchaseOrderStatus: (id: number, status: string) =>
    api.post<PurchaseOrder>('/erp/purchase-orders/status', { id, status }),

  // Inventory
  getInventoryReport: () =>
    api.get<InventoryReportItem[]>('/erp/inventory/report'),
  getLowStock: () =>
    api.get<LowStockItem[]>('/erp/inventory/low-stock'),
  adjustInventory: (data: any) =>
    api.post<{ success: boolean }>('/erp/inventory/adjust', data),

  // Dashboard (combines multiple endpoints)
  getDashboardStats: () =>
    api.get<DashboardStats>('/erp/dashboard'),
};
