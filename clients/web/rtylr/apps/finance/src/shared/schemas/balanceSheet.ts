import { z } from 'zod';

const balanceSheetRowSchema = z.object({
  label: z.string().default(''),
  amount: z.number().optional().default(0),
  type: z.string().optional().default(''),
});

export const balanceSheetSchema = z
  .object({
    assets: z.array(balanceSheetRowSchema).optional().default([]),
    liabilities: z.array(balanceSheetRowSchema).optional().default([]),
    equity: z.array(balanceSheetRowSchema).optional().default([]),
    totalAssets: z.number().optional().nullable(),
    totalLiabilities: z.number().optional().nullable(),
    totalEquity: z.number().optional().nullable(),
  })
  .passthrough();
