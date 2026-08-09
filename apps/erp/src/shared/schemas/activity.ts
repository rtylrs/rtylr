import { z } from 'zod';

export const activityItemSchema = z.object({
  id: z.number(),
  type: z.enum(['product', 'order', 'transfer', 'adjustment', 'supplier']).catch('product'),
  title: z.string().default(''),
  description: z.string().default(''),
  user: z.string().default(''),
  timestamp: z.string().default(''),
});

export type ActivityItemRaw = z.infer<typeof activityItemSchema>;
