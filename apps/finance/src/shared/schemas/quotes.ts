import { z } from 'zod';

export const quoteSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String),
    client: z.string().default(''),
    amount: z.number().optional().nullable(),
    status: z.string().optional().nullable(),
    date: z.string().optional().nullable(),
    issuedDate: z.string().optional().nullable(),
    expiryDate: z.string().optional().nullable(),
    expiresAt: z.string().optional().nullable(),
  })
  .passthrough();

export type QuoteRaw = z.infer<typeof quoteSchema>;
