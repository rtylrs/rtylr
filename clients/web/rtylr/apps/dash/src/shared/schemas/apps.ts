import { z } from 'zod';

export const appContractSchema = z
  .object({
    id: z.string(),
    name: z.string(),
    description: z.string().optional(),
    isEnabled: z.boolean().optional(),
    is_enabled: z.boolean().optional(),
    icon: z.string().optional(),
  })
  .transform((a) => ({
    ...a,
    isEnabled: a.isEnabled ?? a.is_enabled ?? true,
  }));
