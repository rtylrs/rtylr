import { z } from 'zod';

export const documentItemSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  name: z.string(),
  type: z.enum(['I-9', 'W4', 'Contract', 'NDA', 'Offer Letter', 'Other']).catch('Other'),
  uploadedAt: z.string(),
  expiresAt: z.string().optional(),
  sizeKb: z.number(),
  url: z.string().optional(),
});

export type DocumentItem = z.infer<typeof documentItemSchema>;
