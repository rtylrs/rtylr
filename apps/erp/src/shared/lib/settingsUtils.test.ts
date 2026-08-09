import { describe, it, expect } from 'vitest';
import {
  validateErpSettings,
  isSettingsValid,
  mergeSettings,
  currencySymbol,
  fiscalYearEndMonth,
  defaultErpSettings,
} from './settingsUtils';

describe('validateErpSettings', () => {
  it('returns no errors for valid settings', () => {
    expect(validateErpSettings({ taxRate: 10, lowStockThreshold: 5, fiscalYearStart: 1 })).toEqual(
      []
    );
  });

  it('returns error for negative tax rate', () => {
    const errors = validateErpSettings({ taxRate: -1 });
    expect(errors.some((e) => e.field === 'taxRate')).toBe(true);
  });

  it('returns error for tax rate over 100', () => {
    const errors = validateErpSettings({ taxRate: 101 });
    expect(errors.some((e) => e.field === 'taxRate')).toBe(true);
  });

  it('returns error for negative lowStockThreshold', () => {
    const errors = validateErpSettings({ lowStockThreshold: -1 });
    expect(errors.some((e) => e.field === 'lowStockThreshold')).toBe(true);
  });

  it('returns error for invalid fiscalYearStart', () => {
    const errors = validateErpSettings({ fiscalYearStart: 13 });
    expect(errors.some((e) => e.field === 'fiscalYearStart')).toBe(true);
  });

  it('returns error for negative reorderLeadDays', () => {
    const errors = validateErpSettings({ reorderLeadDays: -1 });
    expect(errors.some((e) => e.field === 'reorderLeadDays')).toBe(true);
  });
});

describe('isSettingsValid', () => {
  it('returns true for valid settings', () => {
    expect(isSettingsValid({ taxRate: 10 })).toBe(true);
  });

  it('returns false for invalid settings', () => {
    expect(isSettingsValid({ taxRate: -5 })).toBe(false);
  });
});

describe('mergeSettings', () => {
  it('merges partial settings into existing', () => {
    const current = defaultErpSettings();
    const merged = mergeSettings(current, { taxRate: 15, autoReorder: true });
    expect(merged.taxRate).toBe(15);
    expect(merged.autoReorder).toBe(true);
    expect(merged.currency).toBe('USD');
  });
});

describe('currencySymbol', () => {
  it('returns correct symbols', () => {
    expect(currencySymbol('USD')).toBe('$');
    expect(currencySymbol('EUR')).toBe('€');
    expect(currencySymbol('GBP')).toBe('£');
    expect(currencySymbol('AED')).toBe('AED');
  });
});

describe('fiscalYearEndMonth', () => {
  it('returns December for January start', () => {
    expect(fiscalYearEndMonth(1)).toBe(12);
  });

  it('returns March for April start', () => {
    expect(fiscalYearEndMonth(4)).toBe(3);
  });

  it('returns June for July start', () => {
    expect(fiscalYearEndMonth(7)).toBe(6);
  });
});

describe('defaultErpSettings', () => {
  it('returns sensible defaults', () => {
    const defaults = defaultErpSettings();
    expect(defaults.currency).toBe('USD');
    expect(defaults.taxRate).toBe(0);
    expect(defaults.autoReorder).toBe(false);
    expect(defaults.valuationMethod).toBe('WAC');
  });
});
