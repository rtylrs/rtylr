import type { StockPayload, StockTransferPayload } from '../../shared/services/api/posService';

export type StockAction = 'receive' | 'adjust' | 'waste' | 'transfer';

export interface InventoryItem {
  itemId: number;
  name: string;
  sku?: string;
  currentQty: number;
  unit: string;
  lowStockThreshold: number;
}

export interface InventoryLocation {
  id: number;
  name: string;
}

export interface StockActionParams {
  itemId: number;
  action: StockAction;
  quantityInput: string;
  branchId?: number | null;
  toBranchId?: number | null;
  employeeId?: number | null;
  reason?: string;
  note?: string;
}

export type BuiltStockAction =
  | { payload: StockPayload; action: Exclude<StockAction, 'transfer'> }
  | { payload: StockTransferPayload; action: 'transfer' };

function parsePositiveQuantity(input: string): number {
  const qty = Number(input);
  if (!Number.isFinite(qty) || qty <= 0) {
    throw new Error('Quantity must be greater than zero.');
  }
  return qty;
}

export function buildStockAction({
  itemId,
  action,
  quantityInput,
  branchId,
  toBranchId,
  employeeId,
  reason,
  note,
}: StockActionParams): BuiltStockAction {
  if (!itemId || itemId <= 0) {
    throw new Error('Select a valid inventory item.');
  }
  if (!branchId || branchId <= 0) {
    throw new Error('Select a terminal location before recording stock movements.');
  }

  const quantity = parsePositiveQuantity(quantityInput);
  const trimmedNote = reason?.trim() || note?.trim() || null;

  if (action === 'transfer') {
    if (!toBranchId || toBranchId <= 0) {
      throw new Error('Select a destination location before transferring stock.');
    }
    if (toBranchId === branchId) {
      throw new Error('Destination location must be different from the current location.');
    }

    return {
      action,
      payload: {
        from_branch_id: branchId,
        to_branch_id: toBranchId,
        product_id: itemId,
        quantity,
        note: trimmedNote,
        employee_id: employeeId ?? null,
      },
    };
  }

  return {
    action,
    payload: {
      branch_id: branchId,
      product_id: itemId,
      quantity,
      note: trimmedNote,
      employee_id: employeeId ?? null,
    },
  };
}

function inventoryRows(raw: unknown): unknown[] {
  if (Array.isArray(raw)) return raw;
  if (!raw || typeof raw !== 'object') return [];
  const record = raw as Record<string, unknown>;
  if (Array.isArray(record.inventory)) return record.inventory;
  if (Array.isArray(record.items)) return record.items;
  return [];
}

export function mapInventoryItems(raw: unknown): InventoryItem[] {
  return inventoryRows(raw)
    .filter(
      (item): item is Record<string, unknown> =>
        typeof item === 'object' && item !== null,
    )
    .map((item) => ({
      itemId: Number(
        item['item_id'] ?? item['itemId'] ?? item['product_id'] ?? item['productId'] ?? item['id'] ?? 0,
      ),
      name: String(item['name'] ?? item['product_name'] ?? item['productName'] ?? ''),
      sku: item['sku'] ? String(item['sku']) : undefined,
      currentQty: Number(
        item['current_qty'] ?? item['currentQty'] ?? item['quantity'] ?? item['qty'] ?? 0,
      ),
      unit: String(item['unit'] ?? 'units'),
      lowStockThreshold: Number(
        item['low_stock_threshold'] ?? item['reorder_point'] ?? item['threshold'] ?? 0,
      ),
    }))
    .filter((item) => item.itemId > 0);
}

function locationRows(raw: unknown): unknown[] {
  if (Array.isArray(raw)) return raw;
  if (!raw || typeof raw !== 'object') return [];
  const record = raw as Record<string, unknown>;
  if (Array.isArray(record.locations)) return record.locations;
  if (Array.isArray(record.branches)) return record.branches;
  return [];
}

export function mapInventoryLocations(raw: unknown): InventoryLocation[] {
  return locationRows(raw)
    .filter(
      (item): item is Record<string, unknown> =>
        typeof item === 'object' && item !== null,
    )
    .map((item) => ({
      id: Number(item.id ?? item.location_id ?? item.locationId ?? 0),
      name: String(item.name ?? item.label ?? ''),
    }))
    .filter((location) => Number.isFinite(location.id) && location.id > 0)
    .map((location) => ({
      id: Math.floor(location.id),
      name: location.name.trim() || `Location ${Math.floor(location.id)}`,
    }));
}
