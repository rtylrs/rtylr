import { z } from 'zod';

export const templateSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  channel: z.enum(['email', 'sms']).catch('email'),
  category: z.string(),
  subject: z.string().nullable().optional(),
  body: z.string(),
  thumbnail: z.string().nullable().optional(),
  updatedAt: z.string(),
});

export type Template = z.infer<typeof templateSchema>;
