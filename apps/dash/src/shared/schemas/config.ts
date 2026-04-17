import { z } from 'zod';

export const configSchema = z
  .object({
    maintanance: z.any().optional(),
    aboutContent: z.any().optional(),
    count: z.any().optional(),
    contactEmail: z.string().nullable().default(null),
    contactNumber: z.string().nullable().default(null),
    faqs: z.any().optional(),
    news: z.any().optional(),
  })
  .passthrough();

export type Config = z.infer<typeof configSchema>;
