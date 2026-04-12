import { z } from 'zod';

export const expenseSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  date: z.string().default(''),
  description: z.string().default(''),
  category: z.string().default(''),
  amount: z.number().default(0),
  receipt: z.boolean().default(false),
  receiptUrl: z.string().optional(),
  approval: z.string().default('pending'),
  submitter: z.string().optional(),
  notes: z.string().optional(),
});

export const expenseCategorySchema = z.object({
  id: z.union([z.string(), z.number()]).transform((v) => Number(v)),
  name: z.string().default(''),
  color: z.string().default('#000000'),
  spent: z.number().optional().nullable(),
  budget: z.number().optional().nullable(),
  count: z.number().optional().nullable(),
});

export type ExpenseRaw = z.infer<typeof expenseSchema>;
export type ExpenseCategoryRaw = z.infer<typeof expenseCategorySchema>;
