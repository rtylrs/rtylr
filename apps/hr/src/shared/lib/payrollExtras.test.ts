import { describe, expect, it } from 'vitest';
import {
  calculatePayStubNet,
  deductionTypeLabel,
  estimateFederalTax,
  ficaEmployeeContribution,
  retirementContribution,
  roundCents,
} from './payrollExtras';

describe('roundCents', () => {
  it('rounds to two decimal places', () => {
    expect(roundCents(10.005)).toBeCloseTo(10.01);
    expect(roundCents(9.994)).toBeCloseTo(9.99);
    expect(roundCents(100)).toBe(100);
  });
});

describe('estimateFederalTax', () => {
  it('calculates tax for single filer', () => {
    const tax = estimateFederalTax(50000, 'single');
    expect(tax).toBeGreaterThan(0);
    expect(tax).toBeLessThan(50000);
  });

  it('calculates lower tax for married filer at same income', () => {
    const single = estimateFederalTax(80000, 'single');
    const married = estimateFederalTax(80000, 'married');
    expect(married).toBeLessThan(single);
  });

  it('returns 0 for 0 income', () => {
    expect(estimateFederalTax(0, 'single')).toBe(0);
  });
});

describe('ficaEmployeeContribution', () => {
  it('calculates FICA on regular wages', () => {
    const fica = ficaEmployeeContribution(5000, 0);
    expect(fica).toBeGreaterThan(0);
    expect(fica).toBeCloseTo(5000 * (0.062 + 0.0145), 1);
  });

  it('caps Social Security at wage base', () => {
    const fica = ficaEmployeeContribution(5000, 160200);
    const medicare = 5000 * 0.0145;
    expect(fica).toBeCloseTo(medicare, 1);
  });
});

describe('calculatePayStubNet', () => {
  it('calculates net pay correctly', () => {
    const net = calculatePayStubNet(5000, 500, 200, 383, 300);
    expect(net).toBeCloseTo(3617);
  });

  it('returns 0 when all deductions equal gross', () => {
    expect(calculatePayStubNet(1000, 400, 200, 200, 200)).toBeCloseTo(0);
  });
});

describe('deductionTypeLabel', () => {
  it('returns correct label for each type', () => {
    expect(deductionTypeLabel('health')).toBe('Health Insurance');
    expect(deductionTypeLabel('dental')).toBe('Dental Insurance');
    expect(deductionTypeLabel('vision')).toBe('Vision Insurance');
    expect(deductionTypeLabel('retirement')).toBe('401(k) Contribution');
    expect(deductionTypeLabel('other')).toBe('Other Deduction');
  });
});

describe('retirementContribution', () => {
  it('calculates contribution based on rate', () => {
    expect(retirementContribution(5000, 0.05)).toBeCloseTo(250);
    expect(retirementContribution(5000, 0.1)).toBeCloseTo(500);
  });

  it('clamps rate to [0, 1]', () => {
    expect(retirementContribution(5000, 1.5)).toBeCloseTo(5000);
    expect(retirementContribution(5000, -0.1)).toBeCloseTo(0);
  });
});
