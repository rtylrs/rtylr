import { z } from 'zod';

export const recurringInvoiceSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String),
    client: z.string().default(''),
    amount: z.number().optional().nullable(),
    cadence: z.string().optional().nullable(),
    frequency: z.string().optional().nullable(),
    nextDate: z.string().optional().nullable(),
    status: z.string().optional().nullable(),
  })
  .passthrough();

export type RecurringInvoiceRaw = z.infer<typeof recurringInvoiceSchema>;
