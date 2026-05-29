// Conversion funnel helpers for the insights analytics feature.

export interface FunnelStep {
  name: string;
  count: number;
}

/**
 * conversionRate returns the percentage of step count relative to the first step.
 */
export function conversionRate(step: FunnelStep, funnel: FunnelStep[]): number {
  const base = funnel[0]?.count ?? 0;
  if (base === 0) return 0;
  return Math.round((step.count / base) * 1000) / 10;
}

/**
 * stepDropOff returns the percentage drop between consecutive steps.
 */
export function stepDropOff(current: FunnelStep, previous: FunnelStep): number {
  if (previous.count === 0) return 0;
  const dropped = previous.count - current.count;
  return Math.round((dropped / previous.count) * 1000) / 10;
}

/**
 * overallConversionRate returns the conversion rate from first to last step.
 */
export function overallConversionRate(funnel: FunnelStep[]): number {
  if (funnel.length < 2) return 0;
  const first = funnel[0].count;
  const last = funnel[funnel.length - 1].count;
  if (first === 0) return 0;
  return Math.round((last / first) * 1000) / 10;
}

/**
 * biggestDropOffStep returns the step with the highest absolute drop from the previous step.
 */
export function biggestDropOffStep(funnel: FunnelStep[]): FunnelStep | undefined {
  if (funnel.length < 2) return undefined;
  let maxDrop = -1;
  let result: FunnelStep | undefined;
  for (let i = 1; i < funnel.length; i++) {
    const drop = funnel[i - 1].count - funnel[i].count;
    if (drop > maxDrop) {
      maxDrop = drop;
      result = funnel[i];
    }
  }
  return result;
}

/**
 * funnelStepNames returns the names of all steps in the funnel.
 */
export function funnelStepNames(funnel: FunnelStep[]): string[] {
  return funnel.map((s) => s.name);
}
