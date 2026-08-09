import { z } from 'zod';

/**
 * Mirrors the deployed insights cash-flow series point
 * (insights module, financial_cash_flow.go CashFlowPointPayload):
 * { date, inflow_minor_units, outflow_minor_units } -> camelCased by the
 * @voxire/api interceptor.
 */
export const cashFlowSeriesPointSchema = z.object({
  date: z.string().default(''),
  inflowMinorUnits: z.number().default(0),
  outflowMinorUnits: z.number().default(0),
});

// Chart-facing point shape (major units, used by CashFlowChart).
export const cashFlowPointSchema = z.object({
  month: z.string().default(''),
  inflow: z.number().default(0),
  outflow: z.number().default(0),
});

export const cashFlowSchema = cashFlowPointSchema;

export type CashFlowSeriesPointRaw = z.infer<typeof cashFlowSeriesPointSchema>;
export type CashFlowPointRaw = z.infer<typeof cashFlowPointSchema>;
export type CashFlowRaw = CashFlowPointRaw;
