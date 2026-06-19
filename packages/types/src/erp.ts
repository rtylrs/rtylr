export interface Product {
  id: number;
  ownerId?: number;
  name: string;
  sku?: string;
  categoryId?: number;
  categoryName?: string;
  category?: string;
  brandId?: number;
  supplierId?: number;
  costPrice: number;
  salePrice: number;
  stockQuantity: number;
  lowStockThreshold?: number;
  reorderQuantity?: number;
  unit?: string;
  isActive?: boolean;
  trackInventory?: boolean;
  createdAt?: string;
  updatedAt?: string;
  /** @deprecated Use stockQuantity from erp_products-backed contracts. */
  stock?: number;
}

export interface Supplier {
  id: string;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  products?: string[];
}

export type PurchaseOrderStatus = 'draft' | 'sent' | 'received' | 'cancelled';

export interface PurchaseOrder {
  id: string;
  supplierId: number | string;
  items: { productId: number; quantity: number; unitCost: number }[];
  total: number;
  status: PurchaseOrderStatus;
  expectedDelivery?: string;
  createdAt: string;
}
