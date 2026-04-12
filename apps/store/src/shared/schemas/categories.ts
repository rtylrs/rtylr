import { z } from 'zod';

export const storeCategorySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  slug: z.string(),
  description: z.string().optional(),
  parentId: z.union([z.number(), z.null()]).optional().nullable(),
  productCount: z.number(),
  imageUrl: z.string().optional(),
  sortOrder: z.number(),
});

export type StoreCategoryParsed = z.infer<typeof storeCategorySchema>;
