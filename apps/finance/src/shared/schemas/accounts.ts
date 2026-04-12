import { z } from 'zod';

export const bankAccountSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String).optional(),
    name: z.string().default(''),
    type: z.string().optional().nullable(),
    balance: z.number().optional().nullable(),
    currency: z.string().optional().nullable(),
    institution: z.string().optional().nullable(),
    mask: z.string().optional().nullable(),
    // transaction fields (used when parsed as bank transactions)
    matched: z.boolean().optional(),
    description: z.string().optional().nullable(),
    date: z.string().optional().nullable(),
    category: z.string().optional().nullable(),
    amount: z.number().optional().nullable(),
  })
  .passthrough();

export const chartOfAccountSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String).optional(),
    name: z.string().default(''),
    code: z.string().optional().nullable(),
    type: z.string().optional().nullable(),
    balance: z.number().optional().nullable(),
  })
  .passthrough();

export type BankAccountRaw = z.infer<typeof bankAccountSchema>;
export type ChartOfAccountRaw = z.infer<typeof chartOfAccountSchema>;
