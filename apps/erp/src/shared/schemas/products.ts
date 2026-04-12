import { z } from 'zod';

export const productSchema = z.object({
  id: z.number(),
  name: z.string(),
  sku: z.string(),
  category: z.string(),
  stock: z.number().default(0),
  lowStock: z.number().default(0),
  costPrice: z.number().default(0),
  salePrice: z.number().default(0),
  status: z.enum(['In Stock', 'Low Stock', 'Out of Stock']).catch('Out of Stock'),
  brand: z.string().optional().nullable(),
  barcode: z.string().optional().nullable(),
  imageUrl: z.string().optional().nullable(),
  description: z.string().optional().nullable(),
  unit: z.string().optional().nullable(),
  tags: z.array(z.string()).optional().nullable(),
  warehouseId: z.number().optional().nullable(),
  reorderPoint: z.number().optional().nullable(),
  leadTimeDays: z.number().optional().nullable(),
});

export const productVariantSchema = z.object({
  id: z.number(),
  productId: z.number(),
  name: z.string(),
  sku: z.string(),
  attributes: z.record(z.string(), z.string()).default({}),
  stock: z.number().default(0),
  price: z.number().default(0),
});

export type ProductRaw = z.infer<typeof productSchema>;
export type ProductVariantRaw = z.infer<typeof productVariantSchema>;
