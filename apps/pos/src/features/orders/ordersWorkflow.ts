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

    return {
      id: idText(raw.id) ?? '',
      name: String(raw.name ?? '').trim(),
      price: nonNegativeNumber(raw.price ?? raw.unitPrice ?? raw.unit_price),
      quantity: positiveInteger(raw.quantity ?? raw.qty),
      notes: textOrUndefined(raw.notes),
    };
  });
}

export function mapRemoteOrders(rawOrders: unknown[]): Order[] {
  return rawOrders.flatMap((rawOrder) => {
    const raw = record(rawOrder);
    if (!raw) return [];

    const id = idText(raw.id);
    if (!id) return [];

    return [
      {
        id,
        number: textOrUndefined(raw.number ?? raw.orderNumber ?? raw.order_number),
        tableNumber: textOrUndefined(raw.tableNumber ?? raw.table_number),
        status: orderStatus(raw.status),
        items: normalizeItems(raw.items),
        subtotal: nonNegativeNumber(raw.subtotal),
        tax: nonNegativeNumber(raw.tax),
        tip: nonNegativeNumber(raw.tip),
        discount: nonNegativeNumber(raw.discount),
        total: nonNegativeNumber(raw.total),
        createdAt: textOrUndefined(raw.createdAt ?? raw.created_at) ?? new Date().toISOString(),
      },
    ];
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
