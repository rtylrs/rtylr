import type { StockPayload } from '../../shared/services/posService';

export type StockAction = 'receive' | 'adjust' | 'waste' | 'transfer';

export interface InventoryItem {
  itemId: number;
  name: string;
  sku?: string;
  currentQty: number;
  unit: string;
  lowStockThreshold: number;
}

export interface StockActionParams {
  itemId: number;
  action: StockAction;
  quantityInput: string;
  reason?: string;
  note?: string;
}

export interface BuiltStockAction {
  payload: StockPayload;
  action: StockAction;
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
  reason,
  note,
}: StockActionParams): BuiltStockAction {
  if (!itemId || itemId <= 0) {
    throw new Error('Select a valid inventory item.');
  }

  const quantity = parsePositiveQuantity(quantityInput);

  return {
    action,
    payload: {
      itemId,
      quantity,
      reason: reason?.trim() || note?.trim() || undefined,
    },
  };
}

export function mapInventoryItems(raw: unknown): InventoryItem[] {
  if (!Array.isArray(raw)) return [];
  return raw
    .filter(
      (item): item is Record<string, unknown> =>
        typeof item === 'object' && item !== null,
    )
    .map((item) => ({
      itemId: Number(item['item_id'] ?? item['itemId'] ?? item['id'] ?? 0),
      name: String(item['name'] ?? ''),
      sku: item['sku'] ? String(item['sku']) : undefined,
      currentQty: Number(item['current_qty'] ?? item['currentQty'] ?? item['qty'] ?? 0),
      unit: String(item['unit'] ?? 'units'),
      lowStockThreshold: Number(item['low_stock_threshold'] ?? item['threshold'] ?? 0),
    }))
    .filter((item) => item.itemId > 0);
}
