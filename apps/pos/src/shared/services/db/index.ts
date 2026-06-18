import Dexie, { type EntityTable } from 'dexie';

// ── Offline-first data types ───────────────────────────────────────

export interface CachedProduct {
  id: number;
  name: string;
  sku: string;
  barcode?: string;
  categoryId: number;
  categoryName: string;
  price: number;
  costPrice: number;
  taxClass: string;
  imageUrl?: string;
  available: boolean;
  variants?: CachedVariant[];
  modifierGroupIds?: number[];
  syncedAt: number;
}

export interface CachedVariant {
  id: number;
  productId: number;
  name: string;
  sku: string;
  barcode?: string;
  price: number;
  stock: number;
}

export interface CachedModifier {
  id: number;
  name: string;
  groupName: string;
  priceDelta: number;
  isRequired: boolean;
  syncedAt: number;
}

export interface CachedCombo {
  id: number;
  name: string;
  description: string;
  price: number;
  items: { productId: number; name: string; quantity: number }[];
  isActive: boolean;
  syncedAt: number;
}

export interface CachedCustomer {
  id: number;
  name: string;
  email?: string;
  phone?: string;
  loyaltyPoints: number;
  storeCredit: number;
  visitCount: number;
  totalSpent: number;
  lastVisitAt?: string;
  syncedAt: number;
}

export interface OfflineOrder {
  localId: string;
  serverId?: number;
  branchId: number;
  terminalId: number;
  employeeId: number;
  tableId?: number;
  type: 'dine_in' | 'takeaway' | 'delivery' | 'online';
  status: string;
  items: OfflineOrderItem[];
  payments: OfflinePayment[];
  customerId?: number;
  customerName?: string;
  subtotal: number;
  discountAmount: number;
  taxAmount: number;
  tipAmount: number;
  total: number;
  notes?: string;
  kitchenNotes?: string;
  synced: boolean;
  createdAt: number;
}

export interface OfflineOrderItem {
  productId: number;
  name: string;
  sku?: string;
  quantity: number;
  unitPrice: number;
  discountPct: number;
  discountAmount: number;
  taxRate: number;
  total: number;
  notes?: string;
  modifiers?: { id: number; name: string; priceDelta: number }[];
  comboId?: number;
  variantName?: string;
}

export interface OfflinePayment {
  method: 'cash' | 'card' | 'online' | 'gift_card' | 'store_credit' | 'other';
  amount: number;
  reference?: string;
  cardLast4?: string;
  giftCardCode?: string;
}

export type ApiRequestMethod = 'get' | 'post' | 'put' | 'patch' | 'delete';
export type ApiMutationMethod = Exclude<ApiRequestMethod, 'get'>;
export type ApiMutationStatus = 'pending' | 'syncing' | 'failed';
export type ApiRequestStatus = 'ok' | 'cached' | 'queued' | 'failed';

export interface ApiSnapshot {
  key: string;
  method: ApiRequestMethod;
  url: string;
  params?: unknown;
  requestBody?: unknown;
  responseBody: unknown;
  updatedAt: number;
}

export interface ApiMutation {
  id: string;
  method: ApiMutationMethod;
  url: string;
  params?: unknown;
  requestBody?: unknown;
  optimisticResponse?: unknown;
  status: ApiMutationStatus;
  retries: number;
  lastError?: string;
  createdAt: number;
  updatedAt: number;
}

export interface ApiRequestLog {
  id: string;
  method: ApiRequestMethod;
  url: string;
  params?: unknown;
  requestBody?: unknown;
  responseBody?: unknown;
  status: ApiRequestStatus;
  error?: string;
  createdAt: number;
  updatedAt: number;
}

// ── Dexie database ─────────────────────────────────────────────────

const db = new Dexie('rtylr-pos') as Dexie & {
  products: EntityTable<CachedProduct, 'id'>;
  modifiers: EntityTable<CachedModifier, 'id'>;
  combos: EntityTable<CachedCombo, 'id'>;
  customers: EntityTable<CachedCustomer, 'id'>;
  offlineOrders: EntityTable<OfflineOrder, 'localId'>;
  apiSnapshots: EntityTable<ApiSnapshot, 'key'>;
  apiMutations: EntityTable<ApiMutation, 'id'>;
  apiRequestLog: EntityTable<ApiRequestLog, 'id'>;
};

db.version(1).stores({
  products: 'id, sku, barcode, categoryId, [categoryId+available]',
  modifiers: 'id, groupName',
  combos: 'id',
  customers: 'id, phone, email',
  offlineOrders: 'localId, synced, createdAt',
});

db.version(2).stores({
  products: 'id, sku, barcode, categoryId, [categoryId+available]',
  modifiers: 'id, groupName',
  combos: 'id',
  customers: 'id, phone, email',
  offlineOrders: 'localId, synced, createdAt',
  apiSnapshots: 'key, method, url, updatedAt',
  apiMutations: 'id, status, createdAt, url',
  apiRequestLog: 'id, method, url, status, createdAt',
});

export { db };
