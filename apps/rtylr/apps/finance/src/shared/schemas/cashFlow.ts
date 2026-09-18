import { z } from 'zod';

/**
 * A point of insights.financial.v1 CashFlowAnalytics.trend.
 *
 * Already in MAJOR units: the handler applies minorToMajor before returning
 * (insights financial_service.go), so nothing here should convert again.
 *
 * This replaces cashFlowSeriesPointSchema, which described
 * `{ date, inflow_minor_units, outflow_minor_units }` from
 * financial_cash_flow.go's internal CashFlowPointPayload. That payload is the
 * service layer's, not the HTTP response's: the only HTTP surface is
 * GetCashFlowAnalytics, which wraps the points as `analytics.trend` and renames
 * the fields. The old schema was parsing a `series` key that no response has
 * ever contained, so the chart resolved to an empty array every time (#910).
 */
export const cashFlowTrendPointSchema = z.object({
  date: z.string().default(''),
  inflow: z.number().default(0),
  outflow: z.number().default(0),
  net: z.number().default(0),
});

// Chart-facing point shape (major units, used by CashFlowChart).
export const cashFlowPointSchema = z.object({
  month: z.string().default(''),
  inflow: z.number().default(0),
  outflow: z.number().default(0),
});

export const cashFlowSchema = cashFlowPointSchema;

export type CashFlowTrendPointRaw = z.infer<typeof cashFlowTrendPointSchema>;
export type CashFlowPointRaw = z.infer<typeof cashFlowPointSchema>;
export type CashFlowRaw = CashFlowPointRaw;
