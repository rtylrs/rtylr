/**
 * Warehouse helper functions used across inventory and stock-transfer features.
 */
import type { Warehouse } from '../services/erpApi';

/** Return the percentage of warehouse capacity that is currently used, clamped to [0, 100]. */
export function warehouseUsedPct(warehouse: Warehouse): number {
  if (warehouse.capacity <= 0) return 0;
  return Math.min(100, Math.round((warehouse.used / warehouse.capacity) * 100));
}

/** Return the number of free units remaining in a warehouse. */
export function warehouseFreeCapacity(warehouse: Warehouse): number {
  return Math.max(0, warehouse.capacity - warehouse.used);
}

/** Return true when the warehouse is at or above 90% capacity. */
export function isWarehouseNearCapacity(warehouse: Warehouse, threshold = 90): boolean {
  return warehouseUsedPct(warehouse) >= threshold;
}

/** Return true when the warehouse is at 100% capacity. */
export function isWarehouseFull(warehouse: Warehouse): boolean {
  return warehouse.used >= warehouse.capacity;
}

/** Sort warehouses with the default warehouse first, then alphabetically. */
export function sortWarehousesDefaultFirst(warehouses: Warehouse[]): Warehouse[] {
  return [...warehouses].sort((a, b) => {
    if (a.isDefault !== b.isDefault) return a.isDefault ? -1 : 1;
    return a.name.localeCompare(b.name);
  });
}

/** Return the default warehouse, or undefined if none is marked as default. */
export function findDefaultWarehouse(warehouses: Warehouse[]): Warehouse | undefined {
  return warehouses.find((w) => w.isDefault);
}

/** Return a human-readable capacity label such as "750 / 1,000". */
export function warehouseCapacityLabel(warehouse: Warehouse): string {
  return `${warehouse.used.toLocaleString()} / ${warehouse.capacity.toLocaleString()}`;
}
