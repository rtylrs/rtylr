import { z } from 'zod';

export const stockTakeSchema = z.object({
  id: z.number(),
  name: z.string(),
  warehouseId: z.number().default(0),
  warehouseName: z.string().default(''),
  status: z.enum(['Draft', 'In Progress', 'Completed']).catch('Draft'),
  itemsCounted: z.number().default(0),
  totalItems: z.number().default(0),
  variance: z.number().default(0),
  createdAt: z.string().default(''),
});

export type StockTakeRaw = z.infer<typeof stockTakeSchema>;
