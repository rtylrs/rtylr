import type { CartProductInput } from '../../shared/utils/cart/cartProduct';
import type { HeldOrder } from '../../shared/stores/heldOrdersStore';

export const HELD_ORDER_ITEM_ERROR = 'This held order contains an item that cannot be restored.';
export const SERVER_HELD_ORDER_ITEM_ERROR =
  'This server held order needs synced menu item data before it can be resumed.';

export interface ServerHeldOrder {
  id: number | string;
  orderNumber: string;
  customerName?: string;
  total: number;
  createdAt: string;
  items?: ServerHeldOrderItem[];
}

interface ServerHeldOrderItem {
  id?: number | string;
  itemId?: number | string;
  menuItemId?: number | string;
  productId?: number | string;
  name: string;
  quantity?: number;
  qty?: number;
  unitPrice?: number;
  price?: number;
}

function record(value: unknown): Record<string, unknown> | null {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

function text(value: unknown): string | undefined {
  if (typeof value !== 'string' && typeof value !== 'number') return undefined;
  const normalized = String(value).trim();
  return normalized || undefined;
}

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

function nonNegativeNumber(value: unknown): number {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric < 0) return 0;
  return numeric;
}

function quantity(value: unknown): number {
  return positiveInteger(value) ?? 1;
}

function productIdFromLineId(lineId: string): number | undefined {
  return positiveInteger(lineId) ?? positiveInteger(lineId.split(':')[0]);
}

function repeatProduct(product: CartProductInput, count: number): CartProductInput[] {
  return Array.from({ length: count }, () => product);
}

function serverOrdersArray(response: unknown): unknown[] {
  if (Array.isArray(response)) return response;

  const root = record(response);
  if (!root) return [];

  if (Array.isArray(root.orders)) return root.orders;

  const data = record(root.data);
  if (data && Array.isArray(data.orders)) return data.orders;
  if (data && Array.isArray(data.data)) return data.data;

  return [];
}

function normalizeServerHeldOrderItem(raw: unknown): ServerHeldOrderItem | null {
  const item = record(raw);
  if (!item) return null;

  const itemId = text(item.itemId ?? item.item_id);
  const menuItemId = text(item.menuItemId ?? item.menu_item_id);
  const productId = text(item.productId ?? item.product_id);
  const id = text(item.id);
  const price = nonNegativeNumber(item.price);
  const unitPrice = nonNegativeNumber(item.unitPrice ?? item.unit_price);

  return {
    ...(id ? { id } : {}),
    ...(itemId ? { itemId } : {}),
    ...(menuItemId ? { menuItemId } : {}),
    ...(productId ? { productId } : {}),
    name: text(item.name) ?? 'Item',
    ...(item.quantity != null ? { quantity: quantity(item.quantity) } : {}),
    ...(item.qty != null ? { qty: quantity(item.qty) } : {}),
    ...(item.unitPrice != null || item.unit_price != null ? { unitPrice } : {}),
    ...(item.price != null ? { price } : {}),
  };
}

export function normalizeServerHeldOrders(response: unknown): ServerHeldOrder[] {
  return serverOrdersArray(response).flatMap((raw) => {
    const order = record(raw);
    if (!order) return [];

    const id = text(order.id);
    const orderNumber = text(order.orderNumber ?? order.order_number) ?? id;
    const total = Number(order.total);
    const createdAt = text(order.createdAt ?? order.created_at);
    if (!id || !orderNumber || !Number.isFinite(total) || total < 0 || !createdAt) return [];

    const customerName = text(order.customerName ?? order.customer_name);
    const items = Array.isArray(order.items)
      ? order.items.flatMap((item) => normalizeServerHeldOrderItem(item) ?? [])
      : undefined;

    return [
      {
        id,
        orderNumber,
        ...(customerName ? { customerName } : {}),
        total,
        createdAt,
        ...(items ? { items } : {}),
      },
    ];
  });
}

export function buildHeldOrderCartProducts(order: HeldOrder): CartProductInput[] {
  return order.items.flatMap((item) => {
    const productId = productIdFromLineId(item.id);
    if (productId == null) {
      throw new Error(HELD_ORDER_ITEM_ERROR);
    }

    return repeatProduct(
      {
        id: item.id,
        productId,
        name: item.name,
        unitPrice: nonNegativeNumber(item.price),
        price: nonNegativeNumber(item.price),
      },
      quantity(item.quantity),
    );
  });
}

export function buildServerHeldOrderResume(order: ServerHeldOrder): {
  orderId: number;
  products: CartProductInput[];
} {
  const orderId = positiveInteger(order.id);
  if (orderId == null) {
    throw new Error(SERVER_HELD_ORDER_ITEM_ERROR);
  }

  const products = (order.items ?? []).flatMap((item) => {
    const productId =
      positiveInteger(item.menuItemId) ??
      positiveInteger(item.productId) ??
      positiveInteger(item.itemId);
    if (productId == null) {
      throw new Error(SERVER_HELD_ORDER_ITEM_ERROR);
    }

    const price = nonNegativeNumber(item.unitPrice ?? item.price);
    return repeatProduct(
      {
        id: String(productId),
        productId,
        name: item.name.trim() || 'Item',
        unitPrice: price,
        price,
      },
      quantity(item.quantity ?? item.qty),
    );
  });

  return { orderId, products };
}
