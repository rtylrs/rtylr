import { z } from 'zod';

export const expiryItemSchema = z.object({
  productId: z.number(),
  name: z.string(),
  sku: z.string(),
  batchId: z.string().default(''),
  expiryDate: z.string().default(''),
  quantity: z.number().default(0),
  warehouseName: z.string().default(''),
});

export type ExpiryItemRaw = z.infer<typeof expiryItemSchema>;
