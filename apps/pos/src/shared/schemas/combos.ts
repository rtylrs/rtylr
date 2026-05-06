import { z } from 'zod';

const safeNumber = (fallback: number) => z.coerce.number().catch(fallback);

const comboItemSchema = z.object({
  itemId: z.union([z.string(), z.number()]).transform(String),
  qty: safeNumber(1).default(1),
  name: z.string().optional(),
});

export const comboSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string().default('Combo'),
  price: safeNumber(0).default(0),
  description: z.string().optional().nullable(),
  items: z.array(comboItemSchema).default([]),
});

export type ComboRaw = z.infer<typeof comboSchema>;
