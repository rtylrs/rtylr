import { z } from 'zod';

export const resourceSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  type: z.string().optional().nullable(),
  description: z.string().optional().nullable(),
  capacity: z.number().optional().nullable(),
  isActive: z.boolean().optional(),
  location: z.string().optional().nullable(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export type Resource = z.infer<typeof resourceSchema>;
