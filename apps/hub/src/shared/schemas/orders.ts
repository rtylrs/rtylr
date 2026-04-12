import { z } from 'zod';

export const hubOrderSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  ownerId: z.union([z.string(), z.number()]).transform(String).optional(),
  customerEmail: z.string().optional().default(''),
  customerName: z.string().optional().default(''),
  customerPhone: z.string().optional().default(''),
  status: z.enum(['pending', 'in_progress', 'completed', 'cancelled']).catch('pending'),
  subtotal: z.number().default(0),
  tax: z.number().default(0),
  shippingCost: z.number().default(0),
  total: z.number().default(0),
  shippingAddress: z.string().optional().default(''),
  notes: z.string().optional().default(''),
  createdAt: z.string().optional().default(''),
  updatedAt: z.string().optional().default(''),
});

export type HubOrder = z.infer<typeof hubOrderSchema>;
