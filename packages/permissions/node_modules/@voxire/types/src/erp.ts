export interface Product {
  id: string;
  name: string;
  sku?: string;
  category?: string;
  costPrice: number;
  salePrice: number;
  stock: number;
  lowStockThreshold?: number;
  unit?: string;
  supplierId?: string;
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
  supplierId: string;
  items: { productId: string; quantity: number; unitCost: number }[];
  total: number;
  status: PurchaseOrderStatus;
  expectedDelivery?: string;
  createdAt: string;
}
