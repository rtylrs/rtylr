import { z } from 'zod';

export const createLocationSchema = z.object({
  name: z.string().min(1),
  address: z.string().optional(),
});

export type CreateLocationForm = z.infer<typeof createLocationSchema>;
