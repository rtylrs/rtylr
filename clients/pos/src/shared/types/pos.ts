// Shared POS types used across the new screens, components, hooks and stores.

export interface MenuItem {
  id: string;
  name: string;
  price: number;
  category: string;
  image?: string;
  available?: boolean;
  sku?: string;
  stock?: number;
  description?: string;
  taxRate?: number;
  barcode?: string;
}

export interface Modifier {
  id: string;
  name: string;
  priceDelta: number;
  group?: string;
  required?: boolean;
  defaultSelected?: boolean;
}

export interface ModifierGroup {
  id: string;
  name: string;
  min: number;
  max: number;
  modifiers: Modifier[];
}

export interface Combo {
  id: string;
  name: string;
  price: number;
  items: { itemId: string; qty: number; name?: string }[];
  description?: string;
}

export interface GiftCard {
  id: string;
  code: string;
  balance: number;
  issuedAt: string;
  active: boolean;
  customerName?: string;
}

export interface Customer {
  id: string;
  name: string;
  phone?: string;
  email?: string;
  loyaltyPoints?: number;
  lastVisit?: string;
}

export type OrderStatus =
  | 'pending'
  | 'confirmed'
  | 'preparing'
  | 'ready'
  | 'served'
  | 'paid'
  | 'cancelled'
  | 'voided'
  | 'refunded'
  | 'held'
  | 'open'
  | 'sent';

export interface OrderLineItem {
  id: string;
  name: string;
  price: number;
  quantity: number;
  sku?: string;
  total?: number;
  status?: string;
  modifiers?: Modifier[];
  notes?: string;
  course?: 'appetizer' | 'main' | 'dessert' | 'drink' | 'other';
  seat?: number;
}

export interface Order {
  id: string;
  number?: string;
  tableNumber?: string | null;
  customerId?: string;
  status: OrderStatus;
  items: OrderLineItem[];
  subtotal: number;
  tax: number;
  tip: number;
  discount: number;
  total: number;
  createdAt: string;
  updatedAt?: string;
  paidAt?: string;
  notes?: string;
}

export type PaymentMethod = 'cash' | 'card' | 'gift_card' | 'other';

export interface PaymentEntry {
  id: string;
  method: PaymentMethod;
  amount: number;
  reference?: string;
  tipAmount?: number;
  createdAt: string;
}

export interface Receipt {
  id: string;
  orderId: string;
  total: number;
  customerName?: string;
  emailed?: boolean;
  printedAt: string;
  items: { name: string; qty: number; price: number }[];
  payments: PaymentEntry[];
}

export interface DaySummary {
  date: string;
  grossSales: number;
  netSales: number;
  taxCollected: number;
  tips: number;
  refunds: number;
  voids: number;
  orderCount: number;
  averageTicket: number;
  cashDrops: number;
  openingFloat: number;
  closingCash: number;
}

export interface OfflineQueuedOrder {
  id: string;
  /** Generated Voxie CreateOrderRequest, replayed verbatim on reconnect. */
  order: import('@voxire/voxie/pos/orders/v1/types').CreateOrderRequest;
  queuedAt: string;
  retries: number;
  lastError?: string;
  status: 'pending' | 'syncing' | 'failed';
}
