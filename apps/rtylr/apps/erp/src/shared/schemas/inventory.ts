import { z } from 'zod';

export const inventoryReportItemSchema = z.object({
  productId: z.number(),
  name: z.string(),
  sku: z.string(),
  currentStock: z.number().default(0),
  value: z.number().default(0),
});

export const lowStockItemSchema = z.object({
  name: z.string(),
  sku: z.string(),
  stock: z.number().default(0),
  minStock: z.number().default(0),
  unit: z.string().default(''),
});

export type InventoryReportItemRaw = z.infer<typeof inventoryReportItemSchema>;
export type LowStockItemRaw = z.infer<typeof lowStockItemSchema>;
