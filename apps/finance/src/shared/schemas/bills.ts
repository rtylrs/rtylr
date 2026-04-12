import { z } from 'zod';

export const billSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String),
    supplier: z.string().default(''),
    vendor: z.string().optional().nullable(),
    category: z.string().optional().nullable(),
    amount: z.number().optional().nullable(),
    status: z.string().optional().nullable(),
    dueDate: z.string().optional().nullable(),
    issuedDate: z.string().optional().nullable(),
  })
  .passthrough();

export type BillRaw = z.infer<typeof billSchema>;
