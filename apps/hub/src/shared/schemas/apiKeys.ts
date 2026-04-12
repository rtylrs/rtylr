import { z } from 'zod';

export const apiKeySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  prefix: z.string(),
  lastFour: z.string(),
  createdAt: z.string(),
  lastUsedAt: z.string().optional(),
  scopes: z.array(z.string()),
});

export type ApiKey = z.infer<typeof apiKeySchema>;
