import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { CartDiscount, CartItemModifier } from "./cartStore";

export interface HeldOrderItem {
  /** Stable cart line key captured at hold time. */
  id: string;
  /** Product/combo id used to rebuild the cart product. */
  productId?: number;
  name: string;
  price: number;
  quantity: number;
  sku?: string;
  barcode?: string;
  notes?: string;
  taxRate?: number;
  modifiers?: CartItemModifier[];
  comboId?: number;
  variantName?: string;
  discountPct?: number;
  discountAmount?: number;
}

export interface HeldOrder {
  id: string;
  label: string;
  items: HeldOrderItem[];
  tableNumber?: string | null;
  note?: string;
  orderDiscount?: CartDiscount | null;
  tipAmount?: number;
  createdAt: string;
}

interface HeldOrdersState {
  orders: HeldOrder[];
}

interface HeldOrdersActions {
  hold: (partial: Omit<HeldOrder, "id" | "createdAt">) => HeldOrder | undefined;
  release: (id: string) => HeldOrder | undefined;
  remove: (id: string) => void;
  updateNote: (id: string, note: string) => void;
}

function textOrUndefined(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim() ?? "";
  return trimmed || undefined;
}

function requiredText(value: string): string | null {
  const trimmed = value.trim();
  return trimmed || null;
}

function nonNegativeNumber(value: unknown): number {
  const numeric = Number(value);
  return Number.isFinite(numeric) && numeric >= 0 ? numeric : 0;
}

function positiveInteger(value: unknown): number {
  const numeric = Number(value);
  return Number.isFinite(numeric) && numeric > 0 ? Math.floor(numeric) : 0;
}

function normalizeModifiers(value: unknown): CartItemModifier[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap((modifier) => {
    if (!modifier || typeof modifier !== "object") return [];
    const candidate = modifier as Partial<CartItemModifier>;
    const id = positiveInteger(candidate.id);
    if (!id) return [];
    return [
      {
        id,
        name: typeof candidate.name === "string" ? candidate.name : "",
        priceDelta: nonNegativeNumber(candidate.priceDelta),
      },
    ];
  });
}

function normalizeItems(items: HeldOrderItem[]): HeldOrderItem[] {
  return items.map((item) => ({
    ...item,
    id: String(item.id ?? "").trim(),
    productId: positiveInteger(item.productId) || undefined,
    name: String(item.name ?? "").trim(),
    price: nonNegativeNumber(item.price),
    quantity: positiveInteger(item.quantity) || 1,
    sku: typeof item.sku === "string" ? item.sku : undefined,
    barcode: typeof item.barcode === "string" ? item.barcode : undefined,
    notes: textOrUndefined(item.notes),
    taxRate: nonNegativeNumber(item.taxRate),
    modifiers: normalizeModifiers(item.modifiers),
    comboId: positiveInteger(item.comboId) || undefined,
    variantName: textOrUndefined(item.variantName),
    discountPct: Math.min(100, nonNegativeNumber(item.discountPct)),
    discountAmount: nonNegativeNumber(item.discountAmount),
  }));
}

function normalizeOrder(order: HeldOrder): HeldOrder {
  return {
    ...order,
    label: String(order.label ?? "").trim(),
    items: normalizeItems(order.items ?? []),
    tableNumber: textOrUndefined(order.tableNumber),
    note: textOrUndefined(order.note),
    orderDiscount: order.orderDiscount
      ? {
          ...order.orderDiscount,
          value:
            order.orderDiscount.type === "percentage"
              ? Math.min(100, nonNegativeNumber(order.orderDiscount.value))
              : nonNegativeNumber(order.orderDiscount.value),
        }
      : null,
    tipAmount: nonNegativeNumber(order.tipAmount),
  };
}

export const useHeldOrdersStore = create<HeldOrdersState & HeldOrdersActions>()(
  persist(
    (set, get) => ({
      orders: [],
      hold: (partial) => {
        const label = requiredText(partial.label);
        if (!label) return undefined;

        const order: HeldOrder = {
          ...partial,
          label,
          items: normalizeItems(partial.items),
          tableNumber: textOrUndefined(partial.tableNumber),
          note: textOrUndefined(partial.note),
          id: "held_" + Math.random().toString(36).slice(2, 10),
          createdAt: new Date().toISOString(),
        };
        set((s) => ({ orders: [order, ...s.orders] }));
        return order;
      },
      release: (id) => {
        const order = get().orders.find((o) => o.id === id);
        if (order) {
          set((s) => ({ orders: s.orders.filter((o) => o.id !== id) }));
        }
        return order;
      },
      remove: (id) =>
        set((s) => ({ orders: s.orders.filter((o) => o.id !== id) })),
      updateNote: (id, note) =>
        set((s) => ({
          orders: s.orders.map((o) =>
            o.id === id ? { ...o, note: textOrUndefined(note) } : o,
          ),
        })),
    }),
    {
      name: "pos-held-orders",
      version: 2,
      merge: (persistedState, currentState) => {
        const persisted = persistedState as Partial<HeldOrdersState> | null;
        const orders = Array.isArray(persisted?.orders) ? persisted.orders : [];
        return {
          ...currentState,
          orders: orders.map(normalizeOrder),
        };
      },
    },
  ),
);
