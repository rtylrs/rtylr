import { z } from 'zod';

const amount = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

const lineSchema = z.object({
  accountId: z.union([z.string(), z.number()]).transform(String).default(''),
  code: z.string().default(''),
  name: z.string().default(''),
  type: z.enum(['asset', 'liability', 'equity']).catch('asset'),
  amount: amount.default(0),
});

/** Mirrors finance.statement.v1 GetBalanceSheetResponse. */
export const balanceSheetSchema = z.object({
  asOf: z.string().default(''),
  lines: z.array(lineSchema).default([]),
  netIncome: amount.default(0),
  totalAssets: amount.default(0),
  totalLiabilities: amount.default(0),
  totalEquity: amount.default(0),
  cashBalance: amount.default(0),
  accountsReceivable: amount.default(0),
  accountsPayable: amount.default(0),
});

export type BalanceSheetData = z.infer<typeof balanceSheetSchema>;
export type BalanceSheetLine = BalanceSheetData['lines'][number];
