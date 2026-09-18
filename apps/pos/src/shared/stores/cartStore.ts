import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// ── Types ────────────────────────────────────────────────────────

export interface CartItemModifier {
  id: number;
  name: string;
  priceDelta: number;
}

export interface CartItem {
  /** Unique key for this line item (product + variant + modifiers hash) */
  lineId: string;
  /** Alias for lineId - used in legacy components */
  id: string;
  productId: number;
  name: string;
  sku?: string;
  barcode?: string;
  unitPrice: number;
  /** Alias for unitPrice - used in legacy components */
  price: number;
  quantity: number;
  notes?: string;
  modifiers: CartItemModifier[];
  comboId?: number;
  variantName?: string;
  /** Per-item discount (percentage 0-100) */
  discountPct: number;
  /** Per-item discount (flat amount) */
  discountAmount: number;
  taxRate: number;
  /** Status for KDS tracking */
  status: 'pending' | 'preparing' | 'ready' | 'served' | 'cancelled';
}

export interface CartPayment {
  method: 'cash' | 'card' | 'online' | 'gift_card' | 'store_credit' | 'other';
  amount: number;
  reference?: string;
  cardLast4?: string;
  giftCardCode?: string;
}

export interface CartDiscount {
  type: 'percentage' | 'fixed';
  value: number;
  reason?: string;
  approvedBy?: number;
}

export interface CartState {
  // ── Order identity ──
  localOrderId: string | null;
  serverOrderId: number | null;
  orderNumber: string | null;
  orderType: 'dine_in' | 'takeaway' | 'delivery' | 'online';

  // ── Location context ──
  branchId: number | null;
  terminalId: number | null;
  employeeId: number | null;
  tableId: number | null;
  floorId: number | null;

  // ── Customer ──
  customerId: number | null;
  customerName: string | null;

  // ── Items ──
  items: CartItem[];

  // ── Order-level discount ──
  orderDiscount: CartDiscount | null;

  // ── Tip ──
  tipAmount: number;

  // ── Payments ──
  payments: CartPayment[];

  // ── Notes ──
  notes: string;
  kitchenNotes: string;

  // ── Held/parked state ──
  isHeld: boolean;
  holdLabel: string;

  // ── Legacy aliases ──
  orderId: string | null;
  tableNumber: string | null;
}

// ── Computed helpers ─────────────────────────────────────────────

function generateLineId(
  productId: number,
  variantName?: string,
  modifiers?: CartItemModifier[]
): string {
  const modKey = (modifiers ?? [])
    .map((m) => m.id)
    .sort()
    .join(',');
  return `${productId}:${variantName ?? ''}:${modKey}`;
}

function computeItemTotal(item: CartItem): number {
  const modifierTotal = item.modifiers.reduce((s, m) => s + nonNegativeNumber(m.priceDelta), 0);
  const basePrice = (item.unitPrice + modifierTotal) * item.quantity;
  const discountedPrice =
    item.discountPct > 0
      ? basePrice * (1 - item.discountPct / 100)
      : basePrice - item.discountAmount;
  return Math.max(0, discountedPrice);
}

function nonNegativeNumber(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  return value;
}

function clampNumber(value: number, min: number, max: number): number {
  if (!Number.isFinite(value)) return min;
  return Math.min(max, Math.max(min, value));
}

function positiveInteger(value: number): number | null {
  if (!Number.isFinite(value) || value <= 0) return null;
  return Math.floor(value);
}

function normalizeContextId(id: number): number | null {
  if (!Number.isFinite(id) || id <= 0) return null;
  return Math.floor(id);
}

function trimmedOrNull(value: string | null): string | null {
  const trimmed = value?.trim() ?? '';
  return trimmed || null;
}

// ── Store ────────────────────────────────────────────────────────

interface CartActions {
  // Items
  addItem: (product: {
    id?: number | string;
    productId: number;
    name: string;
    sku?: string;
    barcode?: string;
    unitPrice: number;
    price?: number;
    taxRate?: number;
    modifiers?: CartItemModifier[];
    comboId?: number;
    variantName?: string;
    [key: string]: unknown;
  }) => void;
  removeItem: (lineId: string) => void;
  updateQuantity: (lineId: string, quantity: number) => void;
  updateItemNotes: (lineId: string, notes: string) => void;
  setItemDiscount: (lineId: string, discountPct: number, discountAmount: number) => void;
  setItemStatus: (lineId: string, status: CartItem['status']) => void;

  // Order-level
  setOrderDiscount: (discount: CartDiscount | null) => void;
  setTip: (amount: number) => void;
  setOrderType: (type: CartState['orderType']) => void;
  setNotes: (notes: string) => void;
  setKitchenNotes: (notes: string) => void;

  // Customer
  setCustomer: (id: number | null, name: string | null) => void;

  // Table
  setTable: (tableId: number | null, floorId?: number | null) => void;

  // Context
  setContext: (branchId: number, terminalId: number, employeeId: number) => void;

  // Payments
  addPayment: (payment: CartPayment) => void;
  removePayment: (index: number) => void;
  clearPayments: () => void;

  // Hold/park
  holdOrder: (label: string) => void;
  resumeOrder: () => void;

  // Order tracking
  setServerOrderId: (id: number) => void;
  setOrderNumber: (num: string) => void;

  // Table number
  setTableNumber: (num: string | null) => void;

  // Notes update
  updateNotes: (notes: string) => void;

  // Reset
  clearCart: () => void;
  newOrder: () => void;

  // Computed
  subtotal: () => number;
  discountTotal: () => number;
  taxTotal: () => number;
  total: () => number;
  amountPaid: () => number;
  amountDue: () => number;
  itemCount: () => number;
}

const initialState: CartState = {
  localOrderId: null,
  serverOrderId: null,
  orderNumber: null,
  orderType: 'takeaway',
  branchId: null,
  terminalId: null,
  employeeId: null,
  tableId: null,
  floorId: null,
  customerId: null,
  customerName: null,
  items: [],
  orderDiscount: null,
  tipAmount: 0,
  payments: [],
  notes: '',
  kitchenNotes: '',
  isHeld: false,
  holdLabel: '',
  orderId: null,
  tableNumber: null,
};

export const useCartStore = create<CartState & CartActions>()(
  persist(
    (set, get) => ({
      ...initialState,

  // ── Items ──────────────────────────────────────────────────────

  addItem: (product) => {
    const modifiers = (product.modifiers ?? []).map((modifier) => ({
      ...modifier,
      priceDelta: nonNegativeNumber(modifier.priceDelta),
    }));
    const lineId = generateLineId(product.productId, product.variantName, modifiers);
    set((state) => {
      const existing = state.items.find((i) => i.lineId === lineId);
      if (existing) {
        return {
          items: state.items.map((i) =>
            i.lineId === lineId ? { ...i, quantity: i.quantity + 1 } : i
          ),
        };
      }
      const newItem: CartItem = {
        lineId,
        id: lineId,
        productId: product.productId,
        name: product.name,
        sku: product.sku,
        barcode: product.barcode,
        unitPrice: nonNegativeNumber(product.unitPrice),
        price: nonNegativeNumber(product.unitPrice),
        quantity: 1,
        modifiers,
        comboId: product.comboId,
        variantName: product.variantName,
        discountPct: 0,
        discountAmount: 0,
        taxRate: nonNegativeNumber(product.taxRate ?? 0),
        status: 'pending',
      };
      return { items: [...state.items, newItem] };
    });
  },

  removeItem: (lineId) =>
    set((state) => ({ items: state.items.filter((i) => i.lineId !== lineId) })),

  updateQuantity: (lineId, quantity) => {
    const normalizedQuantity = positiveInteger(quantity);
    if (!normalizedQuantity) {
      get().removeItem(lineId);
    } else {
      set((state) => ({
        items: state.items.map((i) =>
          i.lineId === lineId ? { ...i, quantity: normalizedQuantity } : i
        ),
      }));
    }
  },

  updateItemNotes: (lineId, notes) =>
    set((state) => ({
      items: state.items.map((i) => (i.lineId === lineId ? { ...i, notes } : i)),
    })),

  setItemDiscount: (lineId, discountPct, discountAmount) =>
    set((state) => ({
      items: state.items.map((i) =>
        i.lineId === lineId
          ? {
              ...i,
              discountPct: clampNumber(discountPct, 0, 100),
              discountAmount: nonNegativeNumber(discountAmount),
            }
          : i
      ),
    })),

  setItemStatus: (lineId, status) =>
    set((state) => ({
      items: state.items.map((i) => (i.lineId === lineId ? { ...i, status } : i)),
    })),

  // ── Order-level ───────────────────────────────────────────────

  setOrderDiscount: (discount) =>
    set({
      orderDiscount: discount
        ? {
            ...discount,
            value:
              discount.type === 'percentage'
                ? clampNumber(discount.value, 0, 100)
                : nonNegativeNumber(discount.value),
          }
        : null,
    }),
  setTip: (tipAmount) => set({ tipAmount: nonNegativeNumber(tipAmount) }),
  setOrderType: (orderType) => set({ orderType }),
  setNotes: (notes) => set({ notes }),
  setKitchenNotes: (kitchenNotes) => set({ kitchenNotes }),

  // ── Customer ──────────────────────────────────────────────────

  setCustomer: (customerId, customerName) => {
    const normalizedCustomerId = customerId === null ? null : positiveInteger(customerId);
    set({
      customerId: normalizedCustomerId,
      customerName: normalizedCustomerId ? trimmedOrNull(customerName) : null,
    });
  },

  // ── Table ─────────────────────────────────────────────────────

  setTable: (tableId, floorId) =>
    set({
      tableId: tableId === null ? null : positiveInteger(tableId),
      floorId: floorId === undefined || floorId === null ? null : positiveInteger(floorId),
    }),

  // ── Context ───────────────────────────────────────────────────

  setContext: (branchId, terminalId, employeeId) => {
    const normalizedBranchId = normalizeContextId(branchId);
    const normalizedTerminalId = normalizeContextId(terminalId);
    const normalizedEmployeeId = normalizeContextId(employeeId);
    if (!normalizedBranchId || !normalizedTerminalId || !normalizedEmployeeId) {
      set({ branchId: null, terminalId: null, employeeId: null });
      return;
    }
    set({
      branchId: normalizedBranchId,
      terminalId: normalizedTerminalId,
      employeeId: normalizedEmployeeId,
    });
  },

  // ── Payments ──────────────────────────────────────────────────

  addPayment: (payment) =>
    set((state) => ({
      payments: [...state.payments, { ...payment, amount: nonNegativeNumber(payment.amount) }],
    })),

  removePayment: (index) =>
    set((state) => ({ payments: state.payments.filter((_, i) => i !== index) })),

  clearPayments: () => set({ payments: [] }),

  // ── Hold/park ─────────────────────────────────────────────────

  holdOrder: (label) => set({ isHeld: true, holdLabel: label }),
  resumeOrder: () => set({ isHeld: false, holdLabel: '' }),

  // ── Order tracking ────────────────────────────────────────────

  setServerOrderId: (serverOrderId) => set({ serverOrderId: positiveInteger(serverOrderId) }),
  setOrderNumber: (orderNumber) => set({ orderNumber: trimmedOrNull(orderNumber) }),

  // ── Table number ──────────────────────────────────────────────
  setTableNumber: (num) => set({ tableNumber: trimmedOrNull(num) }),

  // ── Notes update ─────────────────────────────────────────────
  updateNotes: (notes) => set({ notes }),

  // ── Reset ─────────────────────────────────────────────────────

  clearCart: () => set({ ...initialState }),

  newOrder: () => {
    const { branchId, terminalId, employeeId } = get();
    set({
      ...initialState,
      branchId,
      terminalId,
      employeeId,
      localOrderId: `local-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    });
  },

  // ── Computed ──────────────────────────────────────────────────

  subtotal: () => {
    return get().items.reduce((sum, item) => sum + computeItemTotal(item), 0);
  },

  discountTotal: () => {
    const { items, orderDiscount } = get();
    const itemDiscounts = items.reduce((sum, item) => {
      const modTotal = item.modifiers.reduce((s, m) => s + m.priceDelta, 0);
      const gross = (item.unitPrice + modTotal) * item.quantity;
      const net = computeItemTotal(item);
      return sum + (gross - net);
    }, 0);

    let orderDiscountAmount = 0;
    if (orderDiscount) {
      const sub = get().subtotal();
      orderDiscountAmount =
        orderDiscount.type === 'percentage'
          ? sub * (orderDiscount.value / 100)
          : orderDiscount.value;
    }

    return itemDiscounts + orderDiscountAmount;
  },

  taxTotal: () => {
    const { items, orderDiscount } = get();
    const sub = get().subtotal();
    const orderDiscountRatio = orderDiscount
      ? orderDiscount.type === 'percentage'
        ? orderDiscount.value / 100
        : orderDiscount.value / (sub || 1)
      : 0;

    return items.reduce((sum, item) => {
      const itemTotal = computeItemTotal(item);
      const afterOrderDiscount = itemTotal * (1 - orderDiscountRatio);
      return sum + afterOrderDiscount * (item.taxRate / 100);
    }, 0);
  },

  total: () => {
    const sub = get().subtotal();
    const disc = get().orderDiscount;
    const tax = get().taxTotal();
    const tip = get().tipAmount;

    let orderDiscountAmount = 0;
    if (disc) {
      orderDiscountAmount = disc.type === 'percentage' ? sub * (disc.value / 100) : disc.value;
    }

    return Math.max(0, sub - orderDiscountAmount + tax + tip);
  },

  amountPaid: () => get().payments.reduce((sum, p) => sum + p.amount, 0),

  amountDue: () => Math.max(0, get().total() - get().amountPaid()),

      itemCount: () => get().items.reduce((sum, i) => sum + i.quantity, 0),
    }),
    {
      name: 'pos-active-cart',
      version: 1,
      partialize: (state): CartState => ({
        localOrderId: state.localOrderId,
        serverOrderId: state.serverOrderId,
        orderNumber: state.orderNumber,
        orderType: state.orderType,
        branchId: state.branchId,
        terminalId: state.terminalId,
        employeeId: state.employeeId,
        tableId: state.tableId,
        floorId: state.floorId,
        customerId: state.customerId,
        customerName: state.customerName,
        items: state.items,
        orderDiscount: state.orderDiscount,
        tipAmount: state.tipAmount,
        payments: state.payments,
        notes: state.notes,
        kitchenNotes: state.kitchenNotes,
        isHeld: state.isHeld,
        holdLabel: state.holdLabel,
        orderId: state.orderId,
        tableNumber: state.tableNumber,
      }),
    }
  )
);
