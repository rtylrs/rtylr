import { z } from 'zod';

export const warehouseSchema = z.object({
  id: z.number(),
  name: z.string(),
  address: z.string().default(''),
  manager: z.string().default(''),
  capacity: z.number().default(0),
  used: z.number().default(0),
  isDefault: z.boolean().default(false),
});

export type WarehouseRaw = z.infer<typeof warehouseSchema>;
