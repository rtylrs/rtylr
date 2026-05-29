/**
 * Conversion rate and funnel analytics helpers for CRM.
 */

export interface FunnelStep {
  name: string;
  count: number;
}

/** Returns the conversion rate from one count to another (0-1). */
export function conversionRate(from: number, to: number): number {
  if (from === 0) return 0;
  return Math.min(1, to / from);
}

/** Returns the drop-off count between two funnel steps. */
export function dropOff(from: number, to: number): number {
  return Math.max(0, from - to);
}

/** Returns the drop-off percentage between two funnel steps. */
export function dropOffPct(from: number, to: number): number {
  if (from === 0) return 0;
  return (dropOff(from, to) / from) * 100;
}

/** Calculates conversion rates for each step in a funnel. */
export function funnelRates(
  steps: FunnelStep[]
): Array<FunnelStep & { rate: number; dropOffPct: number }> {
  return steps.map((step, i) => {
    const prev = i === 0 ? step.count : steps[i - 1].count;
    return {
      ...step,
      rate: conversionRate(prev, step.count),
      dropOffPct: dropOffPct(prev, step.count),
    };
  });
}

/** Returns the overall conversion rate from first to last funnel step. */
export function overallConversion(steps: FunnelStep[]): number {
  if (steps.length < 2) return steps.length === 1 ? 1 : 0;
  return conversionRate(steps[0].count, steps[steps.length - 1].count);
}
