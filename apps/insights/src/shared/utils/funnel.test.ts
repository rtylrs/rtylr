import { describe, it, expect } from 'vitest';
import {
  conversionRate,
  stepDropOff,
  overallConversionRate,
  biggestDropOffStep,
  funnelStepNames,
} from './funnel';
import type { FunnelStep } from './funnel';

function makeStep(name: string, count: number): FunnelStep {
  return { name, count };
}

const funnel = [
  makeStep('Visit', 1000),
  makeStep('Signup', 400),
  makeStep('Activate', 200),
  makeStep('Purchase', 100),
];

describe('conversionRate', () => {
  it('calculates rate relative to first step', () => {
    expect(conversionRate(funnel[1], funnel)).toBe(40);
  });

  it('returns 0 when base is 0', () => {
    const empty = [makeStep('A', 0), makeStep('B', 10)];
    expect(conversionRate(empty[1], empty)).toBe(0);
  });
});

describe('stepDropOff', () => {
  it('calculates percentage drop between steps', () => {
    expect(stepDropOff(funnel[1], funnel[0])).toBe(60);
  });

  it('returns 0 when previous count is 0', () => {
    expect(stepDropOff(makeStep('B', 10), makeStep('A', 0))).toBe(0);
  });
});

describe('overallConversionRate', () => {
  it('returns conversion from first to last step', () => {
    expect(overallConversionRate(funnel)).toBe(10);
  });

  it('returns 0 for single step funnel', () => {
    expect(overallConversionRate([makeStep('A', 100)])).toBe(0);
  });
});

describe('biggestDropOffStep', () => {
  it('returns step with highest absolute drop', () => {
    expect(biggestDropOffStep(funnel)?.name).toBe('Signup');
  });

  it('returns undefined for single step', () => {
    expect(biggestDropOffStep([makeStep('A', 100)])).toBeUndefined();
  });
});

describe('funnelStepNames', () => {
  it('returns all step names', () => {
    expect(funnelStepNames(funnel)).toEqual(['Visit', 'Signup', 'Activate', 'Purchase']);
  });
});
