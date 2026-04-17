import { z } from 'zod';

export const integrationSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  category: z
    .enum(['payments', 'messaging', 'accounting', 'analytics', 'delivery', 'data'])
    .catch('data'),
  logo: z.string(),
  status: z.enum(['connected', 'available', 'coming_soon']).catch('available'),
  description: z.string(),
  connectedAt: z.string().optional(),
});

export type Integration = z.infer<typeof integrationSchema>;
