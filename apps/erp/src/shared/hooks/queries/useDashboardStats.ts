import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';

type RawRecord = Record<string, unknown>;

export interface LowStockDashboardItem {
  productId: number;
  name: string;
  sku: string;
  stock: number;
  minStock: number;
  unit: string;
  warehouseName: string;
}

export interface DashboardRecentOrder {
  id: number;
  poNumber: string;
  supplier: string;
  status: string;
  total: number;
  date: string;
}

export interface DashboardTopCategory {
  name: string;
  productCount: number;
}

export interface DashboardStats {
  totalProducts: number;
  lowStockCount: number;
  pendingOrders: number;
  totalSuppliers: number;
  inventoryValue: number;
  lowStockItems: LowStockDashboardItem[];
  recentOrders: DashboardRecentOrder[];
  topCategories: DashboardTopCategory[];
}

function asRecord(value: unknown): RawRecord {
  return value && typeof value === 'object' ? (value as RawRecord) : {};
}

function pick(source: RawRecord, ...keys: string[]): unknown {
  for (const key of keys) {
    if (source[key] !== undefined && source[key] !== null) return source[key];
  }
  return undefined;
}

function asNumber(value: unknown): number {
  if (typeof value === 'number' && Number.isFinite(value)) return value;
  if (typeof value === 'string') {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return 0;
}

function asString(value: unknown): string {
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return String(value);
  return '';
}

function asArray(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
}

function formatStatus(status: string): string {
  const normalized = status.trim().toLowerCase().replace(/_/g, ' ');
  if (!normalized) return '';
  return normalized
    .split(' ')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

function formatDate(value: string): string {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString('en-US');
}

function normalizeDashboardStats(value: unknown): DashboardStats {
  const raw = asRecord(value);
  const lowStockItems = asArray(pick(raw, 'lowStockItems', 'low_stock_items')).map((item) => {
    const row = asRecord(item);
    return {
      productId: asNumber(pick(row, 'productId', 'product_id')),
      name: asString(pick(row, 'name')),
      sku: asString(pick(row, 'sku', 'SKU')),
      stock: asNumber(pick(row, 'stock', 'currentStock', 'current_stock')),
      minStock: asNumber(pick(row, 'minStock', 'reorderPoint', 'reorder_point')),
      unit: asString(pick(row, 'unit')) || 'units',
      warehouseName: asString(pick(row, 'warehouseName', 'warehouse_name')),
    };
  });

  const recentOrders = asArray(pick(raw, 'recentOrders', 'recent_orders')).map((item) => {
    const row = asRecord(item);
    const id = asNumber(pick(row, 'id'));
    return {
      id,
      poNumber: asString(pick(row, 'poNumber', 'po_number')) || `PO-${id}`,
      supplier: asString(pick(row, 'supplier', 'supplierName', 'supplier_name')) || 'Unknown',
      status: formatStatus(asString(pick(row, 'status'))),
      total: asNumber(pick(row, 'total')),
      date: formatDate(asString(pick(row, 'date', 'createdAt', 'created_at'))),
    };
  });

  const topCategories = asArray(pick(raw, 'topCategories', 'top_categories')).map((item) => {
    const row = asRecord(item);
    return {
      name: asString(pick(row, 'name')),
      productCount: asNumber(pick(row, 'productCount', 'product_count')),
    };
  });

  return {
    totalProducts: asNumber(pick(raw, 'totalProducts', 'total_products')),
    lowStockCount: asNumber(pick(raw, 'lowStockCount', 'low_stock_count')),
    pendingOrders: asNumber(pick(raw, 'pendingOrders', 'pending_orders')),
    totalSuppliers: asNumber(pick(raw, 'totalSuppliers', 'total_suppliers')),
    inventoryValue: asNumber(pick(raw, 'inventoryValue', 'inventory_value')),
    lowStockItems,
    recentOrders,
    topCategories,
  };
}

export const useDashboardStats = (): UseQueryResult<DashboardStats> => {
  return useQuery<DashboardStats>({
    queryKey: ['erp-dashboard-stats'],
    queryFn: async (): Promise<DashboardStats> => {
      const res = await unwrap(safe.get<unknown>('/erp/dashboard'));
      return normalizeDashboardStats(res);
    },
  });
};
