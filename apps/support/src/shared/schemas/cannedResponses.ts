import { z } from 'zod';

export const cannedResponseSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  title: z.string(),
  body: z.string(),
  category: z.string().optional(),
  usageCount: z.number().optional(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export type CannedResponse = z.infer<typeof cannedResponseSchema>;
