import { safe } from './api';
import type { Order, PaymentEntry } from '../../types/pos';

// ── Request body types ─────────────────────────────────────────────────────

export interface BranchPayload {
  name: string;
  address?: string;
  phone?: string;
  [key: string]: unknown;
}

export interface FloorPayload {
  name: string;
  branchId?: number;
  [key: string]: unknown;
}

export interface TerminalPayload {
  name: string;
  branchId?: number;
  [key: string]: unknown;
}

export interface TablePayload {
  name: string;
  floorId?: number;
  seats?: number;
  [key: string]: unknown;
}

/** Line item matching the backend CreateOrderItemInput contract (snake_case). */
export interface CreateOrderItemPayload {
  menu_item_id: number;
  name: string;
  sku?: string;
  quantity: number;
  unit_price: number;
  discount_pct?: number;
  tax_rate?: number;
  notes?: string;
}

/** Matches the backend CreateOrderRequest contract (snake_case). */
export interface CreateOrderPayload {
  session_id?: number;
  table_id?: number;
  location_id?: number;
  type?: string;
  customer_id?: number;
  customer_name?: string;
  payment_method?: string;
  notes?: string;
  kitchen_notes?: string;
  items: CreateOrderItemPayload[];
}

export type OrderPayload = CreateOrderPayload;

/** Matches the backend ProcessPaymentRequest contract. */
export interface PaymentPayload {
  payment_method: string;
  amount_paid: number;
}

/** Matches the backend DiscountOrderRequest contract. */
export interface DiscountPayload {
  discount_type: 'flat' | 'percent';
  discount_value: number;
  reason?: string;
}

export interface SplitPayload {
  parts: number;
  splits?: number[][];
  [key: string]: unknown;
}

export interface HoldPayload {
  reason?: string;
  [key: string]: unknown;
}

export interface VoidRefundPayload {
  reason?: string;
  [key: string]: unknown;
}

export interface ModifierPayload {
  name: string;
  options?: Array<{ name: string; price: number }>;
  [key: string]: unknown;
}

export interface ComboPayload {
  name: string;
  items?: number[];
  price?: number;
  [key: string]: unknown;
}

/** Matches the backend OpenSessionRequest contract. */
export interface OpenSessionPayload {
  location_id?: number;
  staff_id?: number | null;
  opening_float?: number;
}

/** Matches the backend CloseSessionRequest contract (id goes in the URL). */
export interface CloseSessionPayload {
  closing_float: number;
}

/** Matches the backend cash movement contracts (drop/pickup/adjust). */
export interface CashMovementPayload {
  branch_id: number;
  terminal_id?: number | null;
  session_id: number;
  employee_id: number;
  amount: number;
  reason?: string | null;
}

export interface GiftCardPayload {
  amount: number;
  code?: string;
  [key: string]: unknown;
}

export interface GiftCardRedeemPayload {
  amount: number;
  orderId?: number;
  [key: string]: unknown;
}

export interface CustomerPayload {
  name: string;
  email?: string;
  phone?: string;
  [key: string]: unknown;
}

export interface LoyaltyConfigPayload {
  pointsPerDollar?: number;
  redeemRate?: number;
  [key: string]: unknown;
}

export interface LoyaltyPayload {
  customer_id: number;
  points: number;
  order_id?: number;
  [key: string]: unknown;
}

/** Matches the backend stock movement contracts (receive/adjust/waste). */
export interface StockPayload {
  branch_id: number;
  product_id: number;
  quantity: number;
  note?: string | null;
  employee_id?: number | null;
}

export interface POSUserPayload {
  name: string;
  pin: string;
  locationId?: number;
  [key: string]: unknown;
}

export interface MemberLocationPayload {
  memberId: string;
  locationId: number;
  [key: string]: unknown;
}

export interface MenuItemUpdatePayload {
  name?: string;
  price?: number;
  available?: boolean;
  [key: string]: unknown;
}

export interface OrderFilterParams {
  status?: string;
  date?: string;
  tableId?: number;
  [key: string]: unknown;
}

export interface MenuItemFilterParams {
  categoryId?: number;
  available?: boolean;
  [key: string]: unknown;
}

export interface ReceiptFilterParams {
  date?: string;
  orderId?: number;
  [key: string]: unknown;
}

export interface AuditFilterParams {
  startDate?: string;
  endDate?: string;
  type?: string;
  [key: string]: unknown;
}

// ── Service ────────────────────────────────────────────────────────────────

export const posService = {
  // ── Branches (locations) ──────────────────────────────────────
  getBranches: () => safe.get('/pos/locations'),
  createBranch: (data: BranchPayload) => safe.post('/pos/locations', data),
  updateBranch: (id: number, data: Partial<BranchPayload>) =>
    safe.put(`/pos/locations/${id}`, data),

  // ── Floors ────────────────────────────────────────────────────
  getFloors: (branchId: number) => safe.get('/pos/floors', { params: { branch_id: branchId } }),
  createFloor: (data: FloorPayload) => safe.post('/pos/floors', data),
  updateFloor: (id: number, data: Partial<FloorPayload>) => safe.put(`/pos/floors/${id}`, data),
  deleteFloor: (id: number) => safe.delete(`/pos/floors/${id}`),

  // ── Terminals ─────────────────────────────────────────────────
  getTerminals: () => safe.get('/pos/terminals'),
  createTerminal: (data: TerminalPayload) => safe.post('/pos/terminals', data),
  updateTerminal: (id: number, data: Partial<TerminalPayload>) =>
    safe.put(`/pos/terminals/${id}`, data),
  heartbeat: () => safe.post('/pos/terminals/heartbeat'),

  // ── Tables ────────────────────────────────────────────────────
  getTables: () => safe.get('/pos/tables'),
  createTable: (data: TablePayload) => safe.post('/pos/tables', data),
  updateTable: (id: number, data: Partial<TablePayload>) => safe.put(`/pos/tables/${id}`, data),
  updateTableStatus: (id: number, status: string) =>
    safe.put(`/pos/tables/${id}/status`, { status }),
  deleteTable: (id: number) => safe.delete(`/pos/tables/${id}`),

  // ── Orders ────────────────────────────────────────────────────
  getOrders: (params?: OrderFilterParams) => safe.get('/pos/orders', { params }),
  getOrder: (id: number) => safe.get(`/pos/orders/${id}`),
  createOrder: (data: OrderPayload) => safe.post('/pos/orders', data),
  updateOrderStatus: (id: number, status: string) =>
    safe.put(`/pos/orders/${id}/status`, { status }),
  chargeOrder: (id: number, data: PaymentPayload) => safe.post(`/pos/orders/${id}/payment`, data),
  voidOrder: (id: number, data: VoidRefundPayload) => safe.post(`/pos/orders/${id}/void`, data),
  refundOrder: (id: number, data: VoidRefundPayload) => safe.post(`/pos/orders/${id}/refund`, data),
  reopenOrder: (id: number) => safe.put(`/pos/orders/${id}/status`, { status: 'open' }),
  applyDiscount: (id: number, data: DiscountPayload) =>
    safe.post(`/pos/orders/${id}/discount`, data),
  splitOrder: (id: number, data: SplitPayload) => safe.post(`/pos/orders/${id}/split`, data),
  holdOrder: (id: number, data: HoldPayload) => safe.post(`/pos/orders/${id}/hold`, data),
  getHeldOrders: () => safe.get('/pos/orders/held'),
  sendToKitchen: (id: number) => safe.post(`/pos/orders/${id}/send-to-kitchen`, {}),
  transferTable: (id: number, newTableId: number) =>
    safe.post(`/pos/orders/${id}/transfer-table`, { new_table_id: newTableId }),
  voidOrderItem: (orderId: number, itemId: number) =>
    safe.delete(`/pos/orders/${orderId}/items/${itemId}`),

  // ── Menu / Products ───────────────────────────────────────────
  getMenuItems: (params?: MenuItemFilterParams) => safe.get('/pos/menu-items', { params }),
  updateMenuItem: (id: number, data: MenuItemUpdatePayload) =>
    safe.put(`/pos/menu-items/${id}`, data),
  getModifiers: () => safe.get('/pos/modifiers'),
  createModifier: (data: ModifierPayload) => safe.post('/pos/modifiers', data),
  updateModifier: (id: number, data: Partial<ModifierPayload>) =>
    safe.put(`/pos/modifiers/${id}`, data),
  deleteModifier: (id: number) => safe.delete(`/pos/modifiers/${id}`),
  getCombos: () => safe.get('/pos/combos'),
  createCombo: (data: ComboPayload) => safe.post('/pos/combos', data),
  updateCombo: (id: number, data: Partial<ComboPayload>) => safe.put(`/pos/combos/${id}`, data),
  deleteCombo: (id: number) => safe.delete(`/pos/combos/${id}`),

  // ── Sessions ──────────────────────────────────────────────────
  getSessions: () => safe.get('/pos/sessions'),
  getCurrentSession: () => safe.get('/pos/sessions/current'),
  openSession: (data: OpenSessionPayload) => safe.post('/pos/sessions/open', data),
  closeSession: (id: number, data: CloseSessionPayload) =>
    safe.post(`/pos/sessions/${id}/close`, data),

  // ── Cash movements ────────────────────────────────────────────
  getCashMovements: (sessionId: number) =>
    safe.get('/pos/cash/movements', { params: { session_id: sessionId } }),
  cashDrop: (data: CashMovementPayload) => safe.post('/pos/cash/drop', data),
  cashPickup: (data: CashMovementPayload) => safe.post('/pos/cash/pickup', data),
  cashAdjust: (data: CashMovementPayload) => safe.post('/pos/cash/adjust', data),

  // ── Gift cards ────────────────────────────────────────────────
  getGiftCards: () => safe.get('/pos/gift-cards'),
  createGiftCard: (data: GiftCardPayload) => safe.post('/pos/gift-cards', data),
  redeemGiftCard: (code: string, data: GiftCardRedeemPayload) =>
    safe.post(`/pos/gift-cards/${code}/redeem`, data),
  getGiftCardBalance: (code: string) => safe.get(`/pos/gift-cards/${code}`),

  // ── KDS ────────────────────────────────────────────────────────
  getKDSOrders: (station?: string) => safe.get('/pos/kds/orders', { params: { station } }),
  updateKDSItemStatus: (itemId: number, status: string) =>
    safe.put(`/pos/kds/items/${itemId}`, { status }),

  // ── Customers ─────────────────────────────────────────────────
  searchCustomers: (q: string) => safe.get('/pos/customers', { params: { q } }),
  createCustomer: (data: CustomerPayload) => safe.post('/pos/customers', data),
  getCustomer: (id: number) => safe.get(`/pos/customers/${id}`),

  // ── Receipts ──────────────────────────────────────────────────
  getReceipts: (params?: ReceiptFilterParams) => safe.get('/pos/receipts', { params }),
  emailReceipt: (id: number, email: string) => safe.post(`/pos/receipts/${id}/email`, { email }),

  // ── Day summary / Reports ─────────────────────────────────────
  getDaySummary: (date?: string) => safe.get('/pos/day-summary', { params: { date } }),

  // ── Manager override ──────────────────────────────────────────
  verifyManagerPin: (pin: string) => safe.post('/pos/manager/verify', { pin }),

  // ── Loyalty ───────────────────────────────────────────────────
  getLoyaltyConfig: () => safe.get('/pos/loyalty/config'),
  updateLoyaltyConfig: (data: LoyaltyConfigPayload) => safe.put('/pos/loyalty/config', data),
  getLoyaltyBalance: (customerId: number) =>
    safe.get('/pos/loyalty/balance', { params: { customer_id: customerId } }),
  earnLoyaltyPoints: (data: LoyaltyPayload) => safe.post('/pos/loyalty/earn', data),
  redeemLoyaltyPoints: (data: LoyaltyPayload) => safe.post('/pos/loyalty/redeem', data),

  // ── Stock / Inventory ─────────────────────────────────────────
  getBranchInventory: (branchId: number) =>
    safe.get('/pos/stock/inventory', { params: { branch_id: branchId } }),
  getLowStock: (branchId: number) =>
    safe.get('/pos/stock/low-stock', { params: { branch_id: branchId } }),
  receiveStock: (data: StockPayload) => safe.post('/pos/stock/receive', data),
  adjustStock: (data: StockPayload) => safe.post('/pos/stock/adjust', data),
  wasteStock: (data: StockPayload) => safe.post('/pos/stock/waste', data),
  transferStock: (data: StockPayload) => safe.post('/pos/stock/transfer', data),

  // ── Audit ─────────────────────────────────────────────────────
  getAuditLog: (params?: AuditFilterParams) => safe.get('/pos/audit', { params }),

  // ── POS Users (cashiers/staff - PIN login) ────────────────────
  getPOSUsers: (locationId?: number) =>
    safe.get('/pos/users', { params: locationId ? { location_id: locationId } : {} }),
  createPOSUser: (data: POSUserPayload) => safe.post('/pos/users', data),
  updatePOSUser: (data: Partial<POSUserPayload> & { id: number }) =>
    safe.patch('/pos/users/update', data),
  deactivatePOSUser: (id: number) => safe.patch('/pos/users/deactivate', { id }),
  verifyPIN: (pin: string) => safe.post('/pos/users/verify-pin', { pin }),

  // ── Member Location Access ────────────────────────────────────
  getMyLocations: () => safe.get('/pos/access/my-locations'),
  assignMemberToLocation: (data: MemberLocationPayload) => safe.post('/pos/access/assign', data),
  removeMemberFromLocation: (id: number) => safe.post('/pos/access/remove', { id }),
};
