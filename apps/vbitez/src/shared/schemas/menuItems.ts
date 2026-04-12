import { z } from 'zod';

export const menuItemSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string().default('Item'),
  price: z.number().default(0),
  category: z.string().default('Other'),
  image: z.string().optional().nullable(),
  available: z.boolean().optional().default(true),
  sku: z.string().optional().nullable(),
  stock: z.number().optional().nullable(),
  description: z.string().optional().nullable(),
  taxRate: z.number().optional().nullable(),
  barcode: z.string().optional().nullable(),
});

export type MenuItemRaw = z.infer<typeof menuItemSchema>;
