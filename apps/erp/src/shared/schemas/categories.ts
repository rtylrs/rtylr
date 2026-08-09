import { z } from 'zod';

export const categorySchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string().optional().nullable(),
  parentId: z.number().optional().nullable(),
  productCount: z.number().default(0),
  sortOrder: z.number().default(0),
});

export type CategoryRaw = z.infer<typeof categorySchema>;
