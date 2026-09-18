import { z } from 'zod';

/**
 * Backend contract: /erp/inventory/report and /erp/inventory/low-stock both
 * return `{ items: InventoryReportItem[], ... }`. After @voxire/api converts
 * snake_case -> camelCase the row shape matches erp/v1 InventoryReportItem.
 * Keep these schemas aligned with the proto so .parse() never throws on a list.
 */
export const inventoryReportItemSchema = z.object({
  productId: z.number(),
  productName: z.string().default(''),
  sku: z.string().default(''),
  unit: z.string().default(''),
  stockQuantity: z.number().default(0),
  lowStockThreshold: z.number().default(0),
  reorderQuantity: z.number().default(0),
  costPrice: z.number().default(0),
  salePrice: z.number().default(0),
  stockValue: z.number().default(0),
  isLowStock: z.boolean().default(false),
});

export type InventoryReportItem = z.infer<typeof inventoryReportItemSchema>;
