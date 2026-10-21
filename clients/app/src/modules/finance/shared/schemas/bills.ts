import { z } from 'zod';

export const billSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String),
    supplier: z.string().default(''),
    // supplierName, currency and description exist so an update can round-trip
    // the stored record. UpdateBillBody marks currency, dueDate, status and
    // description @required, so a caller that cannot read them back has no
    // choice but to invent values, and inventing them overwrites real data.
    supplierName: z.string().optional().nullable(),
    currency: z.string().optional().nullable(),
    description: z.string().optional().nullable(),
    vendor: z.string().optional().nullable(),
    category: z.string().optional().nullable(),
    amount: z.number().optional().nullable(),
    status: z.string().optional().nullable(),
    dueDate: z.string().optional().nullable(),
    issuedDate: z.string().optional().nullable(),
  })
  .passthrough();
