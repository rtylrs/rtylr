import { z } from 'zod';

/**
 * Stock take contract — mirrors the onion backend `StockTakePayload`.
 * Backend marshals snake_case; @voxire/api converts response keys to camelCase,
 * so the schema uses the camelCase form of the Go json tags.
 *
 * Status is a lowercase backend enum: 'draft' | 'in_progress' | 'completed'.
 */
export const STOCK_TAKE_STATUSES = ['draft', 'in_progress', 'completed'] as const;

export type StockTakeStatus = (typeof STOCK_TAKE_STATUSES)[number];

export const stockTakeItemSchema = z.object({
  id: z.number().default(0),
  stockTakeId: z.number().default(0),
  productId: z.number().default(0),
  expectedQty: z.number().default(0),
  countedQty: z.number().default(0),
  variance: z.number().default(0),
  createdAt: z.string().default(''),
});

export const stockTakeSchema = z.object({
  id: z.number(),
  ownerId: z.number().default(0),
  warehouseId: z.number().default(0),
  status: z.enum(STOCK_TAKE_STATUSES).catch('draft'),
  notes: z.string().default(''),
  items: z.array(stockTakeItemSchema).default([]),
  startedAt: z.string().default(''),
  completedAt: z.string().default(''),
  createdAt: z.string().default(''),
});

export type StockTakeItem = z.infer<typeof stockTakeItemSchema>;
export type StockTake = z.infer<typeof stockTakeSchema>;

export interface CreateStockTakeItemInput {
  productId: number;
  expectedQty: number;
  countedQty: number;
}

export interface CreateStockTakeInput {
  warehouseId: number;
  notes?: string;
  items: CreateStockTakeItemInput[];
}

export interface UpdateStockTakeInput {
  id: number;
  status: StockTakeStatus;
  notes?: string;
  items: CreateStockTakeItemInput[];
}

const STATUS_LABELS: Record<StockTakeStatus, string> = {
  draft: 'Draft',
  in_progress: 'In Progress',
  completed: 'Completed',
};

export function stockTakeStatusLabel(status: string): string {
  return STATUS_LABELS[status as StockTakeStatus] ?? status;
}

/** Counting progress derived from items (backend does not send totals). */
export function stockTakeProgress(items: Array<Pick<StockTakeItem, 'countedQty'>>): {
  counted: number;
  total: number;
  pct: number;
} {
  const total = items.length;
  const counted = items.filter((item) => item.countedQty >= 0).length;
  const pct = total > 0 ? Math.round((counted / total) * 100) : 0;
  return { counted, total, pct };
}
