import { z } from 'zod';

export const stockTransferSchema = z.object({
  id: z.number(),
  fromWarehouseId: z.number().default(0),
  toWarehouseId: z.number().default(0),
  fromName: z.string().default(''),
  toName: z.string().default(''),
  items: z.number().default(0),
  status: z.enum(['Draft', 'In Transit', 'Received', 'Cancelled']).catch('Draft'),
  createdAt: z.string().default(''),
  completedAt: z.string().optional().nullable(),
});

export type StockTransferRaw = z.infer<typeof stockTransferSchema>;
