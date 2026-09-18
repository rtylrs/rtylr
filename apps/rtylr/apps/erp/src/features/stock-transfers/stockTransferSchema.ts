import { z } from 'zod';

/**
 * Stock transfer contract — mirrors the onion backend `StockTransferPayload`.
 * @voxire/api converts snake_case responses to camelCase, so the schema uses
 * the camelCase form of the Go json tags.
 *
 * Status is a lowercase backend enum: 'draft' | 'in_transit' | 'received' | 'cancelled'.
 */
export const STOCK_TRANSFER_STATUSES = ['draft', 'in_transit', 'received', 'cancelled'] as const;

export type StockTransferStatus = (typeof STOCK_TRANSFER_STATUSES)[number];

export const stockTransferItemSchema = z.object({
  id: z.number().default(0),
  transferId: z.number().default(0),
  productId: z.number().default(0),
  quantity: z.number().default(0),
  createdAt: z.string().default(''),
});

export const stockTransferSchema = z.object({
  id: z.number(),
  ownerId: z.number().default(0),
  fromWarehouseId: z.number().default(0),
  toWarehouseId: z.number().default(0),
  status: z.enum(STOCK_TRANSFER_STATUSES).catch('draft'),
  notes: z.string().default(''),
  items: z.array(stockTransferItemSchema).default([]),
  createdAt: z.string().default(''),
  updatedAt: z.string().default(''),
});

export type StockTransferItem = z.infer<typeof stockTransferItemSchema>;
export type StockTransfer = z.infer<typeof stockTransferSchema>;

export interface CreateStockTransferItemInput {
  productId: number;
  quantity: number;
}

export interface StockTransferDraftItem {
  productId: string;
  quantity: string;
}

export interface StockTransferItemInput {
  productId: string;
  quantity: number;
}

export function toStockTransferItems(items: StockTransferDraftItem[]): StockTransferItemInput[] {
  return items.flatMap((item) => {
    const productId = item.productId.trim();
    const quantity = Number(item.quantity);
    if (!productId || !Number.isFinite(quantity) || quantity <= 0) return [];
    return [{ productId, quantity }];
  });
}

export interface CreateStockTransferInput {
  fromWarehouseId: number;
  toWarehouseId: number;
  notes?: string;
  items: CreateStockTransferItemInput[];
}

export interface UpdateStockTransferStatusInput {
  id: number;
  status: StockTransferStatus;
}

const STATUS_LABELS: Record<StockTransferStatus, string> = {
  draft: 'Draft',
  in_transit: 'In Transit',
  received: 'Received',
  cancelled: 'Cancelled',
};

export function stockTransferStatusLabel(status: string): string {
  return STATUS_LABELS[status as StockTransferStatus] ?? status;
}
