import { z } from 'zod';

export const appContractSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  isEnabled: z.boolean(),
});

export type AppContract = z.infer<typeof appContractSchema>;
