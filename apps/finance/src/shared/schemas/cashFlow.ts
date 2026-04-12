import { z } from 'zod';

// Matches CashFlowPoint from CashFlowChart component
export const cashFlowPointSchema = z.object({
  month: z.string().default(''),
  inflow: z.number().default(0),
  outflow: z.number().default(0),
});

export const cashFlowSchema = cashFlowPointSchema;

export type CashFlowPointRaw = z.infer<typeof cashFlowPointSchema>;
export type CashFlowRaw = CashFlowPointRaw;
