import { safe } from '../lib/api';
import type { CreatePurchaseOrderRequest } from '@voxire/contracts/erp/purchase_order/v1/create_purchase_order';

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
  categoryId?: number;
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

// Keys are camelCase and converted to the backend's snake_case by @voxire/api
// (categoryId -> category_id, etc.), so they map 1:1 to the onion
// CreateProductRequest. category is sent as categoryId (not a name string) so it
// actually persists.
export interface CreateProductInput {
  name: string;
  sku: string;
  categoryId: number;
  supplierId: number;
  barcode: string;
  description: string;
  unit: string;
  costPrice: number;
  salePrice: number;
  compareAtPrice?: number;
  taxRate?: number;
  currency?: string;
  stockQuantity?: number;
  lowStockThreshold?: number;
  reorderQuantity?: number;
  allowBackorder?: boolean;
  leadTimeDays?: number;
  brandId?: number;
  weight?: number;
  weightUnit?: string;
  length?: number;
  width?: number;
  height?: number;
  dimensionUnit?: string;
  countryOfOrigin?: string;
  hsCode?: string;
  isActive?: boolean;
  trackInventory?: boolean;
  attributes?: { name: string; value: string }[];
}

export interface ProductAttribute {
  id: number;
  name: string;
  value: string;
  position: number;
}

// Complete product as returned by GET /erp/products/{id}/full (every PIM column,
// camelCased by @voxire/api). Used to fully prefill the edit form.
export interface FullProduct {
  id: number;
  categoryId?: number;
  supplierId?: number;
  brandId?: number;
  name: string;
  sku: string;
  barcode?: string;
  description?: string;
  unit?: string;
  costPrice: number;
  salePrice: number;
  stockQuantity: number;
  lowStockThreshold: number;
  reorderQuantity: number;
  isActive: boolean;
  trackInventory: boolean;
  compareAtPrice: number;
  taxRate: number;
  currency: string;
  weight: number;
  length: number;
  width: number;
  height: number;
  countryOfOrigin: string;
  hsCode: string;
  allowBackorder: boolean;
  leadTimeDays: number;
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

interface CategoryMutationPayload {
  name?: string;
  description?: string;
}

function toCategoryMutationPayload(data: Partial<Category>): CategoryMutationPayload {
  return {
    name: data.name,
    description: data.description,
  };
}

interface BrandMutationPayload {
  name?: string;
  logoUrl?: string;
  description?: string;
}

function toBrandMutationPayload(data: Partial<Brand>): BrandMutationPayload {
  return {
    name: data.name,
    logoUrl: data.logoUrl,
    description: data.description,
  };
}

interface WarehouseMutationPayload {
  name?: string;
  address?: string;
  managerName?: string;
  isDefault?: boolean;
}

interface UpdateWarehousePayload extends WarehouseMutationPayload {
  id: number;
}

function toWarehouseMutationPayload(
  data: Partial<Warehouse> & { managerName?: string }
): WarehouseMutationPayload {
  return {
    name: data.name,
    address: data.address,
    managerName: data.managerName ?? data.manager,
    isDefault: data.isDefault,
  };
}

function toUpdateWarehousePayload(data: {
  id: number;
  name: string;
  address?: string;
  manager?: string;
}): UpdateWarehousePayload {
  return {
    id: data.id,
    ...toWarehouseMutationPayload(data),
  };
}

interface StockTransferMutationPayload {
  fromWarehouseId?: number;
  toWarehouseId?: number;
  notes?: string;
  items?: Array<{ productId: number; quantity: number }>;
}

function toStockTransferMutationPayload(
  data: Partial<StockTransfer> & {
    notes?: string;
    transferItems?: Array<{ productId: number; quantity: number }>;
  }
): StockTransferMutationPayload {
  return {
    fromWarehouseId: data.fromWarehouseId,
    toWarehouseId: data.toWarehouseId,
    notes: data.notes,
    items: data.transferItems,
  };
}

interface SupplierMutationPayload {
  name?: string;
  email?: string;
  phone?: string;
  isActive?: boolean;
}

function toSupplierMutationPayload(
  data: CreateSupplierInput | UpdateSupplierInput
): SupplierMutationPayload {
  return {
    name: data.name,
    email: data.email,
    phone: data.phone,
    isActive: 'status' in data ? data.status !== 'Inactive' : undefined,
  };
}

interface UpdateSupplierPayload extends SupplierMutationPayload {
  id: number;
}

function toUpdateSupplierPayload(data: UpdateSupplierInput): UpdateSupplierPayload {
  return {
    id: data.id,
    ...toSupplierMutationPayload(data),
  };
}

interface InventoryAdjustmentPayload {
  productId: number;
  change: number;
  type: 'adjustment';
}

interface PriceListMutationPayload {
  name?: string;
  currency?: string;
  isDefault?: boolean;
}

function toPriceListMutationPayload(data: Partial<PriceList>): PriceListMutationPayload {
  return {
    name: data.name,
    currency: 'USD',
    isDefault: false,
  };
}

interface VariantMutationPayload {
  productId?: number;
  name?: string;
  sku?: string;
  salePrice?: number;
  stockQuantity?: number;
  attributes?: string;
}

function toVariantMutationPayload(data: Partial<ProductVariant>): VariantMutationPayload {
  return {
    productId: data.productId,
    name: data.name,
    sku: data.sku,
    salePrice: data.price,
    stockQuantity: data.stock,
    attributes: data.attributes ? JSON.stringify(data.attributes) : undefined,
  };
}

// ─── API ─────────────────────────────────────────────────────────────────────

export const erpApi = {
  // Products
  listProducts: (params?: {
    search?: string;
    category_id?: number;
    page?: number;
    page_size?: number;
  }) => safe.get<Product[]>('/erp/products', { params }),
  getProduct: (id: number) => safe.get<Product>(`/erp/products/${id}`),

  getProductAttributes: (id: number) =>
    safe.get<{ attributes: ProductAttribute[] }>(`/erp/products/${id}/attributes`),

  getProductFull: (id: number) => safe.get<{ product: FullProduct }>(`/erp/products/${id}/full`),
  createProduct: (data: CreateProductInput) => safe.post<Product>('/erp/products', data),
  updateProduct: (data: UpdateProductInput) => safe.put<Product>('/erp/products', data),
  deleteProduct: (id: number) =>
    safe.delete<{ success: boolean }>('/erp/products', { data: { id } }),
  lookupBarcode: (code: string) =>
    safe.get<Product>(`/erp/products/barcode/${encodeURIComponent(code)}`),

  // Categories
  listCategories: () => safe.get<Category[]>('/erp/categories'),
  createCategory: (data: Partial<Category>) =>
    safe.post<Category>('/erp/categories', toCategoryMutationPayload(data)),
  deleteCategory: (id: number) =>
    safe.delete<{ success: boolean }>('/erp/categories', { data: { id } }),
  updateCategory: (data: { id: number; name: string; description?: string }) =>
    safe.put<Category>('/erp/categories', data),

  // Brands
  listBrands: () => safe.get<Brand[]>('/erp/brands'),
  createBrand: (data: Partial<Brand>) =>
    safe.post<Brand>('/erp/brands', toBrandMutationPayload(data)),
  deleteBrand: (id: number) => safe.delete<{ success: boolean }>('/erp/brands', { data: { id } }),
  updateBrand: (data: { id: number; name: string; description?: string }) =>
    safe.put<Brand>('/erp/brands', data),

  // Warehouses
  listWarehouses: () => safe.get<Warehouse[]>('/erp/warehouses'),
  createWarehouse: (data: Partial<Warehouse>) =>
    safe.post<Warehouse>('/erp/warehouses', toWarehouseMutationPayload(data)),
  deleteWarehouse: (id: number) =>
    safe.delete<{ success: boolean }>('/erp/warehouses', { data: { id } }),
  updateWarehouse: (data: { id: number; name: string; address?: string; manager?: string }) =>
    safe.put<Warehouse>('/erp/warehouses', toUpdateWarehousePayload(data)),

  // Stock transfers
  listStockTransfers: (params?: { page?: number; page_size?: number }) =>
    safe.get<StockTransfer[]>('/erp/stock-transfers', { params }),
  createStockTransfer: (data: Partial<StockTransfer>) =>
    safe.post<StockTransfer>('/erp/stock-transfers', toStockTransferMutationPayload(data)),

  // Stock takes
  listStockTakes: (params?: { page?: number; page_size?: number }) =>
    safe.get<StockTake[]>('/erp/stock-takes', { params }),
  createStockTake: (data: {
    warehouseId: number;
    notes?: string;
    items: Array<{ productId: number; expectedQty: number; countedQty: number }>;
  }) => safe.post<StockTake>('/erp/stock-takes', data),

  // Suppliers
  listSuppliers: (params?: { page?: number; page_size?: number }) =>
    safe.get<Supplier[]>('/erp/suppliers', { params }),
  createSupplier: (data: CreateSupplierInput) =>
    safe.post<Supplier>('/erp/suppliers', toSupplierMutationPayload(data)),
  updateSupplier: (data: UpdateSupplierInput) =>
    safe.put<Supplier>('/erp/suppliers', toUpdateSupplierPayload(data)),
  deleteSupplier: (id: number) => safe.delete<{ success: boolean }>(`/erp/suppliers/${id}`),

  // Purchase Orders
  listPurchaseOrders: (params?: { page?: number; page_size?: number }) =>
    safe.get<PurchaseOrder[]>('/erp/purchase-orders', { params }),
  createPurchaseOrder: (data: CreatePurchaseOrderRequest) =>
    safe.post<PurchaseOrder>('/erp/purchase-orders', data),
  updatePurchaseOrderStatus: (id: number, status: string) =>
    safe.put<PurchaseOrder>('/erp/purchase-orders/status', { id, status }),
  approvePurchaseOrder: (id: number) =>
    safe.post<PurchaseOrder>(`/erp/purchase-orders/${id}/approve`, {}),

  // Inventory
  getInventoryReport: () => safe.get<InventoryReportItem[]>('/erp/inventory/report'),
  getLowStock: () => safe.get<LowStockItem[]>('/erp/inventory/low-stock'),
  adjustInventory: (data: { productId: number; delta?: number; quantity?: number }) =>
    safe.post<{ success: boolean }>('/erp/inventory/adjust', {
      productId: data.productId,
      change: data.delta ?? data.quantity ?? 0,
      type: 'adjustment',
    } satisfies InventoryAdjustmentPayload),
  getValuation: (method: 'FIFO' | 'LIFO' | 'WAC' = 'WAC') =>
    safe.get<any>('/erp/inventory/valuation', { params: { method } }),

  // Reorder / forecasting
  getReorderSuggestions: () => safe.get<ReorderSuggestion[]>('/erp/reorder-suggestions'),

  // Expiry
  getExpiryItems: () => safe.get<ExpiryItem[]>('/erp/expiry'),

  // Price lists
  listPriceLists: () => safe.get<PriceList[]>('/erp/price-lists'),
  createPriceList: (data: Partial<PriceList>) =>
    safe.post<PriceList>('/erp/price-lists', toPriceListMutationPayload(data)),

  // Variants
  listVariants: (productId?: number) =>
    safe.get<ProductVariant[]>('/erp/variants', { params: productId ? { productId } : undefined }),
  createVariant: (data: Partial<ProductVariant>) =>
    safe.post<ProductVariant>('/erp/variants', toVariantMutationPayload(data)),
  // Backend stores attributes as a JSON string and uses salePrice/stockQuantity,
  // so write inputs intentionally diverge from the ProductVariant display type.
  updateVariant: (data: {
    id: number;
    productId?: number;
    name: string;
    sku?: string;
    salePrice?: number;
    stockQuantity?: number;
    attributes?: string;
  }) => safe.put<{ message: string }>('/erp/variants', data),
  deleteVariant: (id: number) =>
    safe.delete<{ message: string }>('/erp/variants', { data: { id } }),

  // Activity log
  listActivity: (limit = 20) => safe.get<ActivityItem[]>('/erp/activity', { params: { limit } }),

  // Dashboard (combines multiple endpoints)
  getDashboardStats: () => safe.get<DashboardStats>('/erp/dashboard'),
};
