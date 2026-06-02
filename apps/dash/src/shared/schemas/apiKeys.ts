import { z } from 'zod';

// Mirrors the onion hub_api_keys payload. @voxire/api converts the wire
// snake_case (key_prefix, created_at, …) to camelCase before this parse runs,
// so the keys here are camelCase. Only key_prefix is returned (first 8 chars);
// the full key is shown once at creation time and never again.
export const apiKeySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  keyPrefix: z.string().optional().default(''),
  scopes: z.array(z.string()).optional().default([]),
  createdAt: z.string().optional().default(''),
  lastUsedAt: z.string().optional(),
  isActive: z.boolean().optional(),
});

export type ApiKey = z.infer<typeof apiKeySchema>;
