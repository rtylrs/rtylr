import type { Order, OrderLineItem, OrderStatus } from '../../shared/types/pos';

const statuses: OrderStatus[] = [
  'pending',
  'confirmed',
  'preparing',
  'ready',
  'served',
  'paid',
  'cancelled',
  'voided',
  'refunded',
  'held',
  'open',
  'sent',
];

function textOrUndefined(value: unknown): string | undefined {
  const trimmed = String(value ?? '').trim();
  return trimmed || undefined;
}

function record(value: unknown): Record<string, unknown> | null {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

function idText(value: unknown): string | null {
  if (typeof value === 'number') {
    if (!Number.isFinite(value) || value <= 0) return null;
    return String(Math.floor(value));
  }

  const trimmed = String(value ?? '').trim();
  return trimmed || null;
}

function nonNegativeNumber(value: unknown): number {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric < 0) return 0;
  return numeric;
}

function optionalNonNegativeNumber(value: unknown): number | undefined {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric < 0) return undefined;
  return numeric;
}

function positiveInteger(value: unknown): number {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return 0;
  return Math.floor(numeric);
}

function orderStatus(value: unknown): OrderStatus {
  return statuses.includes(value as OrderStatus) ? (value as OrderStatus) : 'pending';
}

function normalizeItems(items: unknown): OrderLineItem[] {
  if (!Array.isArray(items)) return [];

  return items.flatMap((item) => {
    const raw = record(item);
    if (!raw) return [];

    const id = idText(raw.id ?? raw.item_id ?? raw.order_item_id ?? raw.menu_item_id) ?? '';
    const name =
      textOrUndefined(raw.name ?? raw.itemName ?? raw.item_name ?? raw.menuItemName) ??
      (id ? `Item #${id}` : 'Item');
    const price = nonNegativeNumber(raw.price ?? raw.unitPrice ?? raw.unit_price ?? raw.unit_amount);
    const quantity = positiveInteger(raw.quantity ?? raw.qty);

    const normalized: OrderLineItem = {
      id,
      name,
      price,
      quantity,
    };
    const sku = textOrUndefined(raw.sku);
    const total = optionalNonNegativeNumber(raw.total ?? raw.lineTotal ?? raw.line_total);
    const status = textOrUndefined(raw.status);
    const notes = textOrUndefined(raw.notes);
    if (sku) normalized.sku = sku;
    if (total !== undefined) normalized.total = total;
    if (status) normalized.status = status;
    if (notes) normalized.notes = notes;
    return normalized;
  });
}

export function mapRemoteOrders(rawOrders: unknown[]): Order[] {
  return rawOrders.flatMap((rawOrder) => {
    const raw = record(rawOrder);
    if (!raw) return [];

    const id = idText(raw.id);
    if (!id) return [];

    const order: Order = {
      id,
      status: orderStatus(raw.status),
      items: normalizeItems(raw.items),
      subtotal: nonNegativeNumber(raw.subtotal),
      tax: nonNegativeNumber(raw.tax ?? raw.taxAmount ?? raw.tax_amount),
      tip: nonNegativeNumber(raw.tip ?? raw.tipAmount ?? raw.tip_amount),
      discount: nonNegativeNumber(raw.discount ?? raw.discountAmount ?? raw.discount_amount),
      total: nonNegativeNumber(raw.total),
      createdAt: textOrUndefined(raw.createdAt ?? raw.created_at) ?? new Date().toISOString(),
    };

    const number = textOrUndefined(raw.number ?? raw.orderNumber ?? raw.order_number);
    const tableNumber = textOrUndefined(raw.tableNumber ?? raw.table_number);
    const updatedAt = textOrUndefined(raw.updatedAt ?? raw.updated_at);
    const paidAt = textOrUndefined(raw.paidAt ?? raw.paid_at);
    const notes = textOrUndefined(raw.notes);
    if (number) order.number = number;
    if (tableNumber) order.tableNumber = tableNumber;
    if (updatedAt) order.updatedAt = updatedAt;
    if (paidAt) order.paidAt = paidAt;
    if (notes) order.notes = notes;

    return [order];
  });
}

function remoteOrdersArray(response: unknown): unknown[] {
  if (Array.isArray(response)) return response;

  const root = record(response);
  if (!root) return [];
  if (Array.isArray(root.orders)) return root.orders;

  const data = record(root.data);
  if (data && Array.isArray(data.orders)) return data.orders;
  if (data && Array.isArray(data.data)) return data.data;
  if (Array.isArray(root.data)) return root.data;

  return [];
}

export function normalizeOrdersResponse(response: unknown): Order[] {
  return mapRemoteOrders(remoteOrdersArray(response));
}

function remoteOrderObject(response: unknown): unknown | null {
  const root = record(response);
  if (!root) return null;

  if (record(root.order)) return root.order;

  const data = record(root.data);
  if (data && record(data.order)) return data.order;
  if (data && record(data.data)) return data.data;

  if (idText(root.id)) return root;

  return null;
}

export function normalizeOrderResponse(response: unknown): Order | null {
  const rawOrder = remoteOrderObject(response);
  if (!rawOrder) return null;
  return mapRemoteOrders([rawOrder])[0] ?? null;
}
