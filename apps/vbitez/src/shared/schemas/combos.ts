import { z } from 'zod';

const comboItemSchema = z.object({
  itemId: z.union([z.string(), z.number()]).transform(String),
  qty: z.number().default(1),
  name: z.string().optional(),
});

export const comboSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string().default('Combo'),
  price: z.number().default(0),
  description: z.string().optional().nullable(),
  items: z.array(comboItemSchema).default([]),
});

export type ComboRaw = z.infer<typeof comboSchema>;
