import { z } from 'zod';

export const webhookSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  url: z.string(),
  events: z.array(z.string()),
  status: z.enum(['active', 'paused', 'failing']).catch('active'),
  lastDelivery: z.string().optional(),
  secretPreview: z.string().optional(),
});

export type Webhook = z.infer<typeof webhookSchema>;
