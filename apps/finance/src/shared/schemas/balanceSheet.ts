import { z } from 'zod';

export const balanceSheetSchema = z
  .object({
    assets: z.any().optional(),
    liabilities: z.any().optional(),
    equity: z.any().optional(),
    totalAssets: z.number().optional().nullable(),
    totalLiabilities: z.number().optional().nullable(),
    totalEquity: z.number().optional().nullable(),
  })
  .passthrough();

export type BalanceSheetRaw = z.infer<typeof balanceSheetSchema>;
