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
  brand?: string;
  barcode?: string;
  imageUrl?: string;
  description?: string;
  unit?: string;
  tags?: string[];
  warehouseId?: number;
  reorderPoint?: number;
  leadTimeDays?: number;
}

export interface CreateProductInput {
  name: string;
  sku: string;
  category: string;
  stock: number;
  lowStock: number;
  costPrice: number;
  salePrice: number;
  brand?: string;
  barcode?: string;
  description?: string;
  unit?: string;
}

export interface UpdateProductInput extends Partial<CreateProductInput> {
  id: number;
}

export interface Category {
  id: number;
  name: string;
  description?: string;
  parentId?: number | null;
  productCount: number;
  sortOrder: number;
}

export interface Brand {
  id: number;
  name: string;
  logoUrl?: string;
  description?: string;
  productCount: number;
}

export interface Warehouse {
  id: number;
  name: string;
  address: string;
  manager: string;
  capacity: number;
  used: number;
  isDefault: boolean;
}

export interface StockTransfer {
  id: number;
  fromWarehouseId: number;
  toWarehouseId: number;
  fromName: string;
  toName: string;
  items: number;
  status: 'Draft' | 'In Transit' | 'Received' | 'Cancelled';
  createdAt: string;
  completedAt?: string;
}

export interface StockTake {
  id: number;
  name: string;
  warehouseId: number;
  warehouseName: string;
  status: 'Draft' | 'In Progress' | 'Completed';
  itemsCounted: number;
  totalItems: number;
  variance: number;
  createdAt: string;
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
  onTimeDeliveryPct?: number;
  fulfillmentAccuracy?: number;
  leadTimeDays?: number;
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
  status: 'Draft' | 'Sent' | 'Received' | 'Pending Approval' | 'Approved';
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

export interface ProductVariant {
  id: number;
  productId: number;
  name: string;
  sku: string;
  attributes: Record<string, string>;
  stock: number;
  price: number;
}

export interface PriceList {
  id: number;
  name: string;
  description: string;
  tier: 'Retail' | 'Wholesale' | 'VIP' | 'Custom';
  multiplier: number;
  activeCount: number;
}

export interface ExpiryItem {
  productId: number;
  name: string;
  sku: string;
  batchId: string;
  expiryDate: string;
  quantity: number;
  warehouseName: string;
}

export interface ReorderSuggestion {
  productId: number;
  name: string;
  sku: string;
  currentStock: number;
  avgDailySales: number;
  suggestedQuantity: number;
  leadTimeDays: number;
  confidence: 'high' | 'medium' | 'low';
}

export interface ActivityItem {
  id: number;
  type: 'product' | 'order' | 'transfer' | 'adjustment' | 'supplier';
  title: string;
  description: string;
  user: string;
  timestamp: string;
}

// ─── API ─────────────────────────────────────────────────────────────────────

export const erpApi = {
  // Products
  listProducts: (params?: { search?: string; categoryId?: number; limit?: number; offset?: number }) =>
    api.get<Product[]>('/erp/products', { params }),
  getProduct: (id: number) =>
    api.get<Product>(`/erp/products/${id}`),
  createProduct: (data: CreateProductInput) =>
    api.post<Product>('/erp/products', data),
  updateProduct: (data: UpdateProductInput) =>
    api.post<Product>('/erp/products', data),
  deleteProduct: (id: number) =>
    api.post<{ success: boolean }>('/erp/products', { id, _action: 'delete' }),
  lookupBarcode: (code: string) =>
    api.get<Product>(`/erp/products/barcode/${encodeURIComponent(code)}`),

  // Categories
  listCategories: () => api.get<Category[]>('/erp/categories'),
  createCategory: (data: any) => api.post<Category>('/erp/categories', data),

  // Brands
  listBrands: () => api.get<Brand[]>('/erp/brands'),
  createBrand: (data: any) => api.post<Brand>('/erp/brands', data),

  // Warehouses
  listWarehouses: () => api.get<Warehouse[]>('/erp/warehouses'),
  createWarehouse: (data: any) => api.post<Warehouse>('/erp/warehouses', data),

  // Stock transfers
  listStockTransfers: () => api.get<StockTransfer[]>('/erp/stock-transfers'),
  createStockTransfer: (data: any) => api.post<StockTransfer>('/erp/stock-transfers', data),

  // Stock takes
  listStockTakes: () => api.get<StockTake[]>('/erp/stock-takes'),
  createStockTake: (data: any) => api.post<StockTake>('/erp/stock-takes', data),

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
  approvePurchaseOrder: (id: number) =>
    api.post<PurchaseOrder>(`/erp/purchase-orders/${id}/approve`, {}),

  // Inventory
  getInventoryReport: () =>
    api.get<InventoryReportItem[]>('/erp/inventory/report'),
  getLowStock: () =>
    api.get<LowStockItem[]>('/erp/inventory/low-stock'),
  adjustInventory: (data: any) =>
    api.post<{ success: boolean }>('/erp/inventory/adjust', data),
  getValuation: (method: 'FIFO' | 'LIFO' | 'WAC' = 'WAC') =>
    api.get<any>('/erp/inventory/valuation', { params: { method } }),

  // Reorder / forecasting
  getReorderSuggestions: () =>
    api.get<ReorderSuggestion[]>('/erp/reorder-suggestions'),

  // Expiry
  getExpiryItems: () =>
    api.get<ExpiryItem[]>('/erp/expiry'),

  // Price lists
  listPriceLists: () => api.get<PriceList[]>('/erp/price-lists'),

  // Variants
  listVariants: (productId?: number) =>
    api.get<ProductVariant[]>('/erp/variants', { params: productId ? { productId } : undefined }),

  // Activity log
  listActivity: (limit = 20) =>
    api.get<ActivityItem[]>('/erp/activity', { params: { limit } }),

  // Dashboard (combines multiple endpoints)
  getDashboardStats: () =>
    api.get<DashboardStats>('/erp/dashboard'),
};
