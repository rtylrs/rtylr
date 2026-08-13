import { describe, expect, it } from 'vitest';
import { isPayrollConfigured, type PayrollConfigSource } from './payrollConfig';

describe('isPayrollConfigured', () => {
  it('accepts the legacy period-only config', () => {
    expect(isPayrollConfigured({ payrollPeriod: 'monthly' })).toBe(true);
  });

  it('accepts enabled and configured capability flags without a period', () => {
    const config: PayrollConfigSource = {
      payrollEnabled: true,
      payrollConfigured: true,
    };
    expect(isPayrollConfigured(config)).toBe(true);
  });

  it('reads snake_case capability flags from the API response', () => {
    const config: PayrollConfigSource = {
      payroll_enabled: true,
      payroll_configured: true,
    };
    expect(isPayrollConfigured(config)).toBe(true);
  });

  it('fails closed when payroll is disabled or unconfigured', () => {
    expect(isPayrollConfigured({ payrollPeriod: 'monthly', payrollEnabled: false })).toBe(false);
    expect(isPayrollConfigured({ payrollPeriod: 'monthly', payrollConfigured: false })).toBe(false);
    expect(isPayrollConfigured(null)).toBe(false);
    expect(isPayrollConfigured({})).toBe(false);
  });
});
