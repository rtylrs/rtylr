import { z } from 'zod';

const kdsOrderItemSchema = z.object({
  name: z.string().default('Item'),
  qty: z.number().default(1),
  course: z.string().optional().nullable(),
  notes: z.string().optional().nullable(),
});

export const kdsOrderSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  number: z.union([z.string(), z.number()]).transform(String).optional(),
  station: z.string().default('Kitchen'),
  items: z.array(kdsOrderItemSchema).default([]),
  status: z.enum(['new', 'cooking', 'ready', 'served']).catch('new'),
  createdAt: z.string().default(() => new Date().toISOString()),
  elapsedSeconds: z.number().optional(),
});

export type KDSOrderRaw = z.infer<typeof kdsOrderSchema>;
