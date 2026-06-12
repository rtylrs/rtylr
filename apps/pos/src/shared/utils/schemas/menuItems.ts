import { z } from 'zod';

const safeNumber = (fallback: number) => z.coerce.number().catch(fallback);
const optionalNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === '') return undefined;
  const numeric = Number(value);
  return Number.isFinite(numeric) ? numeric : undefined;
}, z.number().optional());

function normalizeMenuItemAliases(value: unknown): unknown {
  if (!value || typeof value !== 'object') return value;
  const record = value as Record<string, unknown>;

  return {
    ...record,
    price: record.price ?? record.salePrice ?? record.sale_price,
    category: record.category ?? record.categoryName ?? record.category_name,
    image: record.image ?? record.imageUrl ?? record.image_url,
    available: record.available ?? record.isActive ?? record.is_active,
    taxRate: record.taxRate ?? record.tax_rate,
  };
}

export const menuItemSchema = z.preprocess(
  normalizeMenuItemAliases,
  z.object({
    id: z.union([z.string(), z.number()]).transform(String),
    name: z.string().default('Item'),
    price: safeNumber(0).default(0),
    category: z.string().default('Other'),
    image: z.string().optional().nullable(),
    available: z.boolean().optional().default(true),
    sku: z.string().optional().nullable(),
    stock: optionalNumber,
    description: z.string().optional().nullable(),
    taxRate: optionalNumber,
    barcode: z.string().optional().nullable(),
  }),
);

export type MenuItemRaw = z.infer<typeof menuItemSchema>;
