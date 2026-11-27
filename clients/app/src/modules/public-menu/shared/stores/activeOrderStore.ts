import { useEffect, useState } from 'react';
import { menuOrderStatusPath } from '../../menuPaths';

export interface ActivePublicOrder {
  orderId: string;
  orderNumber: string;
  businessId: string;
  slug?: string;
  publicToken: string;
  status: string;
  createdAt: number;
}

const STORAGE_KEY = 'rtylr-active-orders';
const ACTIVE_ORDER_TTL_MS = 24 * 60 * 60 * 1000;
const MAX_STORED_ORDERS = 5;

const TERMINAL_STATUSES = new Set([
  'delivered',
  'completed',
  'done',
  'served',
  'paid',
  'cancelled',
  'refunded',
  'voided',
]);

const listeners = new Set<() => void>();

function notifyListeners(): void {
  for (const listener of listeners) {
    listener();
  }
}

export function isTerminalOrderStatus(status?: string | null): boolean {
  if (!status) return false;
  return TERMINAL_STATUSES.has(status.toLowerCase());
}

function getStorage(): Storage | null {
  try {
    return typeof window !== 'undefined' && window.localStorage ? window.localStorage : null;
  } catch {
    return null;
  }
}

export function readActiveOrders(now = Date.now()): ActivePublicOrder[] {
  const storage = getStorage();
  if (!storage) return [];
  try {
    const raw = storage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    const valid = parsed.filter(
      (item): item is ActivePublicOrder =>
        Boolean(item) &&
        typeof item.orderId === 'string' &&
        item.orderId.length > 0 &&
        typeof item.orderNumber === 'string' &&
        typeof item.businessId === 'string' &&
        item.businessId.length > 0 &&
        typeof item.publicToken === 'string' &&
        item.publicToken.length > 0 &&
        typeof item.createdAt === 'number' &&
        now - item.createdAt < ACTIVE_ORDER_TTL_MS &&
        !isTerminalOrderStatus(item.status)
    );
    if (valid.length !== parsed.length) {
      writeActiveOrders(valid, false);
    }
    return valid;
  } catch {
    return [];
  }
}

function writeActiveOrders(orders: ActivePublicOrder[], shouldNotify = true): void {
  const storage = getStorage();
  if (!storage) return;
  try {
    if (orders.length === 0) {
      storage.removeItem(STORAGE_KEY);
    } else {
      storage.setItem(STORAGE_KEY, JSON.stringify(orders.slice(0, MAX_STORED_ORDERS)));
    }
  } catch {
    // Ignore storage quota / private browsing failures.
  }
  if (shouldNotify) notifyListeners();
}

export function saveActiveOrder(order: ActivePublicOrder): void {
  if (!order.orderId || !order.businessId || !order.publicToken) return;
  if (isTerminalOrderStatus(order.status)) {
    dismissActiveOrder(order.orderId);
    return;
  }
  const current = readActiveOrders();
  const prev = current.find((item) => item.orderId === order.orderId);
  const merged: ActivePublicOrder = {
    ...order,
    orderNumber: order.orderNumber || prev?.orderNumber || order.orderId,
    slug: order.slug ?? prev?.slug,
    status:
      order.status && order.status !== 'pending'
        ? order.status
        : (prev?.status ?? order.status ?? 'pending'),
    createdAt: prev?.createdAt ?? order.createdAt,
  };
  const rest = current.filter((item) => item.orderId !== order.orderId);
  writeActiveOrders([merged, ...rest]);
}

export function updateActiveOrderStatus(orderId: string, status: string): void {
  if (!orderId || !status) return;
  if (isTerminalOrderStatus(status)) {
    dismissActiveOrder(orderId);
    return;
  }
  const current = readActiveOrders();
  let changed = false;
  const next = current.map((item) => {
    if (item.orderId !== orderId) return item;
    if (item.status === status) return item;
    changed = true;
    return { ...item, status };
  });
  if (changed) writeActiveOrders(next);
}

export function dismissActiveOrder(orderId: string): void {
  if (!orderId) return;
  const current = readActiveOrders();
  const next = current.filter((item) => item.orderId !== orderId);
  if (next.length !== current.length) {
    writeActiveOrders(next);
  }
}

export function clearActiveOrders(): void {
  writeActiveOrders([]);
}

export function getActiveOrdersForBusiness(
  businessId?: string | number | null
): ActivePublicOrder[] {
  if (businessId == null || businessId === '' || businessId === 0) return [];
  const key = String(businessId);
  return readActiveOrders().filter((item) => item.businessId === key);
}

export function getMostRecentActiveOrder(): ActivePublicOrder | null {
  return readActiveOrders()[0] ?? null;
}

export function buildOrderStatusUrl(
  order: Pick<ActivePublicOrder, 'orderId' | 'orderNumber' | 'businessId' | 'slug' | 'publicToken'>
): string {
  const params = new URLSearchParams();
  if (order.orderNumber) params.set('number', order.orderNumber);
  if (order.businessId) params.set('business_id', order.businessId);
  if (order.slug) params.set('slug', order.slug);
  if (order.publicToken) params.set('public_token', order.publicToken);
  const query = params.toString();
  return `${menuOrderStatusPath(order.orderId)}${query ? `?${query}` : ''}`;
}

export function useActiveOrders(businessId?: string | number | null): ActivePublicOrder[] {
  const [orders, setOrders] = useState<ActivePublicOrder[]>(() =>
    businessId != null ? getActiveOrdersForBusiness(businessId) : readActiveOrders()
  );

  useEffect(() => {
    const sync = () => {
      setOrders(businessId != null ? getActiveOrdersForBusiness(businessId) : readActiveOrders());
    };
    sync();
    listeners.add(sync);
    const onStorage = (event: StorageEvent) => {
      if (event.key === null || event.key === STORAGE_KEY) sync();
    };
    window.addEventListener('storage', onStorage);
    return () => {
      listeners.delete(sync);
      window.removeEventListener('storage', onStorage);
    };
  }, [businessId]);

  return orders;
}
