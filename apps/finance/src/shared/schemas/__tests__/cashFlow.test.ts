import { describe, it, expect } from 'vitest';
import { cashFlowPointSchema } from '../cashFlow';

describe('cashFlowPointSchema', () => {
  it('defaults all fields when missing', () => {
    const result = cashFlowPointSchema.parse({});
    expect(result.month).toBe('');
    expect(result.inflow).toBe(0);
    expect(result.outflow).toBe(0);
  });

  it('parses a valid cash flow point', () => {
    const result = cashFlowPointSchema.parse({ month: 'Jan', inflow: 5000, outflow: 3000 });
    expect(result.month).toBe('Jan');
    expect(result.inflow).toBe(5000);
    expect(result.outflow).toBe(3000);
  });

  it('computes net flow manually from inflow and outflow', () => {
    const result = cashFlowPointSchema.parse({ month: 'Feb', inflow: 8000, outflow: 6000 });
    const net = result.inflow - result.outflow;
    expect(net).toBe(2000);
  });

  it('accepts zero values', () => {
    const result = cashFlowPointSchema.parse({ month: 'Mar', inflow: 0, outflow: 0 });
    expect(result.inflow).toBe(0);
    expect(result.outflow).toBe(0);
  });
});
