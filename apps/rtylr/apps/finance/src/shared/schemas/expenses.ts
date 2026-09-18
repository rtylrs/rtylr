import { z } from 'zod';

/**
 * Mirrors the deployed onion expenses `Expense` payload (proto finance/v1
 * Expense): id, ownerId, categoryId?, amount, currency, description, vendor,
 * receiptUrl, expenseDate, status, createdAt, updatedAt. Keys are camelCase
 * because the @voxire/api interceptor converts the snake_case wire shape.
 */
export const expenseSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  ownerId: z.number().optional().nullable(),
  categoryId: z.number().optional().nullable(),
  description: z.string().default(''),
  amount: z.number().default(0),
  currency: z.string().optional().nullable(),
  vendor: z.string().optional().nullable(),
  receiptUrl: z.string().optional().nullable(),
  expenseDate: z.string().default(''),
  status: z.string().default('pending'),
  notes: z.string().optional().nullable(),
  createdAt: z.string().optional().nullable(),
  updatedAt: z.string().optional().nullable(),
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
