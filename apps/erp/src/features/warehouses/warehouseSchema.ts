import { z } from 'zod';

/**
 * Warehouse contract — mirrors the onion backend `WarehousePayload`.
 * @voxire/api converts snake_case responses to camelCase, so the schema uses
 * the camelCase form of the Go json tags (e.g. manager_name -> managerName).
 *
 * The backend does NOT return any capacity-usage figure, so there is no `used`
 * field; the page shows configured capacity only.
 */
export const warehouseSchema = z.object({
  id: z.number(),
  ownerId: z.number().default(0),
  name: z.string(),
  address: z.string().default(''),
  city: z.string().default(''),
  country: z.string().default(''),
  managerName: z.string().default(''),
  capacity: z.number().default(0),
  isDefault: z.boolean().default(false),
  isActive: z.boolean().default(true),
  createdAt: z.string().default(''),
  updatedAt: z.string().default(''),
});

export type Warehouse = z.infer<typeof warehouseSchema>;

export interface CreateWarehouseInput {
  name: string;
  address?: string;
  city?: string;
  country?: string;
  managerName?: string;
  capacity?: number;
  isDefault?: boolean;
  isActive?: boolean;
}

export interface UpdateWarehouseInput extends CreateWarehouseInput {
  id: number;
}
