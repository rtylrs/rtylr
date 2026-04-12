import { z } from 'zod';

export const modifierSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string().default('Option'),
  priceDelta: z.number().default(0),
  group: z.string().optional(),
  required: z.boolean().optional().default(false),
  defaultSelected: z.boolean().optional().default(false),
});

export const modifierGroupSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string().default('Modifiers'),
  min: z.number().default(0),
  max: z.number().default(1),
  modifiers: z.array(modifierSchema).default([]),
});

export type ModifierRaw = z.infer<typeof modifierSchema>;
export type ModifierGroupRaw = z.infer<typeof modifierGroupSchema>;
