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

/**
 * The money a queued offline sale already took from the customer.
 *
 * CreateOrderRequest carries no amount, so the tender is captured at checkout -
 * while the cash is still in the drawer - and stored with the queue entry. It
 * has to live on the entry itself: a separate store left a window in which the
 * entry existed with no tender, which is what the old payload-fingerprint
 * matching heuristic existed to paper over.
 */
export interface OfflineTender {
  paymentMethod: string;
  amountPaid: number;
  /** Local order id the tender was taken against; shown on the queue screen. */
  localOrderId: string;
}

export interface OfflineQueuedOrder {
  id: string;
  /** Generated Voxie CreateOrderRequest, replayed verbatim on reconnect. */
  order: import('@voxire/voxie/pos/orders/v1/types').CreateOrderRequest;
  queuedAt: string;
  retries: number;
  lastError?: string;
  status: 'pending' | 'syncing' | 'failed';
  /**
   * Optional only because an entry persisted by an older build has none: this
   * is a persisted shape read back across versions, not a value `enqueue` is
   * allowed to omit. The drain backfills both fields below before its first
   * send, so an entry read here is never replayed without them.
   */
  tender?: OfflineTender;
  /**
   * Idempotency-Key minted once for the user action that queued this sale, and
   * replayed verbatim by every drain and every manual Retry. The server answers
   * 428 without it and dedups on it, so a crash between sending the charge and
   * reading the response re-issues the same key instead of minting a second one
   * for what is the same sale.
   */
  idempotencyKey?: string;
  /**
   * Order id returned by createOrder, written here BEFORE the charge. A crash
   * from then on replays this charge - which the server makes a no-op on an
   * already-paid order - instead of creating a second sale.
   */
  serverOrderId?: string;
}
