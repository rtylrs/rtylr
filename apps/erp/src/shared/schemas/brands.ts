import { z } from 'zod';

export const brandSchema = z.object({
  id: z.number(),
  name: z.string(),
  logoUrl: z.string().optional().nullable(),
  description: z.string().optional().nullable(),
  productCount: z.number().default(0),
});

export type BrandRaw = z.infer<typeof brandSchema>;
