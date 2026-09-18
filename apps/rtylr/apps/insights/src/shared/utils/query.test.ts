import { describe, it, expect } from 'vitest';
import { insightsKeys } from './query';

describe('insightsKeys', () => {
  it('all starts with "insights"', () => {
    expect(insightsKeys.all[0]).toBe('insights');
  });

  it('dashboard key includes all prefix', () => {
    expect(insightsKeys.dashboard()).toContain('insights');
    expect(insightsKeys.dashboard()).toContain('dashboard');
  });

  it('revenueSeries key includes params', () => {
    const params = { from: '2026-01-01', to: '2026-01-31', granularity: 'daily' };
    const key = insightsKeys.revenueSeries(params);
    expect(key[key.length - 1]).toEqual(params);
  });

  it('topCustomers key includes limit when provided', () => {
    const key = insightsKeys.topCustomers(5);
    const last = key[key.length - 1] as { limit: number };
    expect(last.limit).toBe(5);
  });

  it('customerKpi key contains customers scope', () => {
    expect(insightsKeys.customerKpi()).toContain('customers');
    expect(insightsKeys.customerKpi()).toContain('kpi');
  });

  it('cashFlow key contains params object', () => {
    const params = { from: '2026-01-01', to: '2026-03-31' };
    const key = insightsKeys.cashFlow(params);
    expect(key[key.length - 1]).toEqual(params);
  });

  it('reportTemplates key contains reports scope', () => {
    expect(insightsKeys.reportTemplates()).toContain('reports');
    expect(insightsKeys.reportTemplates()).toContain('templates');
  });

  it('settings key contains settings scope', () => {
    expect(insightsKeys.settings()).toContain('settings');
  });

  it('different resources produce different keys', () => {
    expect(insightsKeys.dashboard()).not.toEqual(insightsKeys.customerKpi());
    expect(insightsKeys.revenue()).not.toEqual(insightsKeys.financial());
  });
});
