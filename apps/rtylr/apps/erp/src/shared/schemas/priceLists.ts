import { z } from 'zod';

export const priceListSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string().default(''),
  tier: z.enum(['Retail', 'Wholesale', 'VIP', 'Custom']).catch('Retail'),
  multiplier: z.number().default(1),
  activeCount: z.number().default(0),
});

export type PriceListRaw = z.infer<typeof priceListSchema>;
