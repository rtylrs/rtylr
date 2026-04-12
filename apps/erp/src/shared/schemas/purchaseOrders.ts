import { z } from 'zod';

export const purchaseOrderSchema = z.object({
  id: z.number(),
  supplier: z.string().default(''),
  supplierId: z.number().default(0),
  date: z.string().default(''),
  expectedDate: z.string().default(''),
  items: z.number().default(0),
  total: z.string().default('0'),
  status: z.enum(['Draft', 'Sent', 'Received', 'Pending Approval', 'Approved']).catch('Draft'),
});

export type PurchaseOrderRaw = z.infer<typeof purchaseOrderSchema>;
