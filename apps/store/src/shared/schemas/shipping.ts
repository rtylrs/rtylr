import { z } from 'zod';

export const shippingMethodSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  type: z.enum(['flat_rate', 'free', 'weight_based', 'price_based']).catch('flat_rate'),
  cost: z.number(),
  minOrderAmount: z.number().optional(),
  estimatedDays: z.string(),
  enabled: z.boolean(),
});

export const shippingZoneSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  regions: z.array(z.string()),
  methods: z.array(shippingMethodSchema),
});

export type ShippingMethodParsed = z.infer<typeof shippingMethodSchema>;
export type ShippingZoneParsed = z.infer<typeof shippingZoneSchema>;
