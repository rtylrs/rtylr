import { z } from 'zod';

export const supplierSchema = z.object({
  id: z.number(),
  name: z.string(),
  contactName: z.string().default(''),
  email: z.string().default(''),
  phone: z.string().default(''),
  productCount: z.number().default(0),
  categories: z.array(z.string()).default([]),
  status: z.enum(['Active', 'Inactive']).catch('Active'),
  onTimeDeliveryPct: z.number().optional().nullable(),
  fulfillmentAccuracy: z.number().optional().nullable(),
  leadTimeDays: z.number().optional().nullable(),
});

export type SupplierRaw = z.infer<typeof supplierSchema>;
