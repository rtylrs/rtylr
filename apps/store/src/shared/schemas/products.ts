import { z } from 'zod';

export const storeProductSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  sku: z.string(),
  category: z.string().optional().default(''),
  categoryId: z.number().optional(),
  price: z.number(),
  compareAtPrice: z.number().optional(),
  stock: z.number(),
  status: z.enum(['active', 'draft', 'archived']).catch('draft'),
  imageUrl: z.string().optional(),
  description: z.string().optional(),
  tags: z.array(z.string()).optional(),
  weight: z.number().optional(),
  weightUnit: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type StoreProductParsed = z.infer<typeof storeProductSchema>;
