import type { AdjustStockRequest, BranchInventory, ReceiveStockRequest, TransferStockRequest, WasteStockRequest } from '@voxire/voxie/pos/stock/v1/types';
import type { MyLocation } from '@voxire/voxie/pos/access/v1/types';
import type { Location } from '@voxire/voxie/pos/location/v1/types';

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

export interface LowStockSummary {
  count: number;
  names: string[];
  hiddenCount: number;
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
  | { payload: ReceiveStockRequest; action: 'receive' }
  | { payload: AdjustStockRequest; action: 'adjust' }
  | { payload: WasteStockRequest; action: 'waste' }
  | { payload: TransferStockRequest; action: 'transfer' };

/**
 * One stock movement, identified by the operator action that recorded it.
 *
 * `Idempotency-Key` is a required header on the transfer route (the server
 * answers 428 without it) and is also what dedups a retry, so it belongs to the
 * action rather than to the request: every attempt of this action sends the
 * same key, and a later, different movement gets its own. It travels in the
 * same object as the request so no transport can drop it on the floor.
 */
export type StockActionSubmission =
  | { payload: ReceiveStockRequest; action: 'receive'; idempotencyKey: string }
  | { payload: AdjustStockRequest; action: 'adjust'; idempotencyKey: string }
  | { payload: WasteStockRequest; action: 'waste'; idempotencyKey: string }
  | { payload: TransferStockRequest; action: 'transfer'; idempotencyKey: string };

/**
 * Hands one operator action to the stock client under a single key.
 *
 * The key is minted ONCE, above every retry. Generating it per attempt would
 * give a double-tap or a retried submission two keys, so both would execute
 * and the server's fingerprint check could not tell them apart from two real
 * movements.
 */
export function submitStockAction(
  action: BuiltStockAction,
  send: (submission: StockActionSubmission) => Promise<unknown>,
): Promise<unknown> {
  const idempotencyKey = crypto.randomUUID();
  return send({ ...action, idempotencyKey });
}

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
        fromBranchId: String(branchId),
        toBranchId: String(toBranchId),
        productId: String(itemId),
        quantity,
        note: trimmedNote ?? '',
        employeeId: String(employeeId ?? 0),
      },
    };
  }

  if (action === 'receive') {
    return {
      action: 'receive',
      payload: {
        branchId: String(branchId),
        productId: String(itemId),
        quantity,
        note: trimmedNote ?? '',
        employeeId: String(employeeId ?? 0),
      },
    };
  }

  if (action === 'waste') {
    return {
      action: 'waste',
      payload: {
        branchId: String(branchId),
        productId: String(itemId),
        quantity,
        note: trimmedNote ?? '',
        employeeId: String(employeeId ?? 0),
      },
    };
  }

  return {
    action: 'adjust',
    payload: {
      branchId: String(branchId),
      productId: String(itemId),
      quantity,
      note: trimmedNote ?? '',
      employeeId: String(employeeId ?? 0),
    },
  };
}

export function mapInventoryItems(items: BranchInventory[]): InventoryItem[] {
  return items
    .map((item) => ({
      itemId: Number(item.itemId || item.productId || item.id || 0),
      name: item.name ?? '',
      sku: item.sku || undefined,
      currentQty: item.currentQty || item.quantity || 0,
      unit: item.unit || 'units',
      lowStockThreshold: item.lowStockThreshold || item.reorderPoint || 0,
    }))
    .filter((item) => item.itemId > 0);
}

export function summarizeLowStockItems(
  items: Pick<InventoryItem, 'name'>[],
  visibleItemCount = 4,
): LowStockSummary {
  const visibleCount = Math.max(0, Math.floor(visibleItemCount));
  const names = items
    .slice(0, visibleCount)
    .map((item) => item.name.trim())
    .filter(Boolean);

  return {
    count: items.length,
    names,
    hiddenCount: Math.max(0, items.length - visibleCount),
  };
}

function inventoryLocationId(item: MyLocation | Location): number {
  if ('locationId' in item && item.locationId !== undefined) {
    return Number(item.locationId);
  }
  if ('id' in item && item.id !== undefined) {
    return Number(item.id);
  }
  return NaN;
}

export function mapInventoryLocations(items: Array<MyLocation | Location>): InventoryLocation[] {
  return items
    .map((item) => ({
      id: inventoryLocationId(item),
      name: item.name ?? '',
    }))
    .filter((location) => Number.isFinite(location.id) && location.id > 0)
    .map((location) => ({
      id: Math.floor(location.id),
      name: location.name.trim() || `Location ${Math.floor(location.id)}`,
    }));
}
