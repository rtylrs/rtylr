// Pure helpers for KDS (Kitchen Display System) order management.

import type { KDSOrderRaw } from './schemas/kdsOrders';

/** Return true if a KDS order is still in an active/open state. */
export function isKdsOrderActive(order: KDSOrderRaw): boolean {
  return order.status === 'new' || order.status === 'cooking';
}

/** Return true if a KDS order is complete (ready or served). */
export function isKdsOrderComplete(order: KDSOrderRaw): boolean {
  return order.status === 'ready' || order.status === 'served';
}

/** Return orders assigned to a specific station. */
export function ordersByStation(
  orders: KDSOrderRaw[],
  station: string,
): KDSOrderRaw[] {
  return orders.filter((o) => o.station.toLowerCase() === station.toLowerCase());
}

/** Return the next status in the KDS workflow given the current status. */
export function nextKdsStatus(
  current: KDSOrderRaw['status'],
): KDSOrderRaw['status'] {
  const flow: Record<KDSOrderRaw['status'], KDSOrderRaw['status']> = {
    new: 'cooking',
    cooking: 'ready',
    ready: 'served',
    served: 'served',
  };
  return flow[current];
}

/** Return total item count across all orders. */
export function totalKdsItemCount(orders: KDSOrderRaw[]): number {
  return orders.reduce((sum, order) => sum + order.items.reduce((s, item) => s + item.qty, 0), 0);
}

/** Sort KDS orders oldest-first (by createdAt). */
export function sortKdsOrdersOldestFirst(orders: KDSOrderRaw[]): KDSOrderRaw[] {
  return [...orders].sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
  );
}

/** Return unique station names from a list of KDS orders. */
export function uniqueStations(orders: KDSOrderRaw[]): string[] {
  const stations = new Set(orders.map((o) => o.station));
  return Array.from(stations).sort();
}
