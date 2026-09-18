import { z } from 'zod';

const resourceIdSchema = z.union([z.string(), z.number()]).transform(String);

export const bankAccountSchema = z
  .object({
    id: resourceIdSchema,
    name: z.string(),
    type: z.string().nullable(),
    balance: z.number().nullable(),
    currency: z.string().nullable(),
    institution: z.string().nullable(),
    mask: z.string().nullable(),
  })
  .strict();

export const bankTransactionSchema = z
  .object({
    id: resourceIdSchema,
    matched: z.boolean(),
    description: z.string().nullable(),
    date: z.string().nullable(),
    category: z.string().nullable(),
    amount: z.number().nullable(),
  })
  .strict();

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
export type BankTransactionRaw = z.infer<typeof bankTransactionSchema>;
export type ChartOfAccountRaw = z.infer<typeof chartOfAccountSchema>;
