import { describe, it, expect } from 'vitest';
import {
  currencySymbol,
  fiscalYearLabel,
  timezoneDisplayName,
  defaultCurrencyForDisplay,
} from './settings';
import type { InsightsSettingsResponse } from '../hooks/queries/types';

function makeSettings(overrides: Partial<InsightsSettingsResponse> = {}): InsightsSettingsResponse {
  return {
    defaultCurrency: 'USD',
    fiscalYearStart: '01-01',
    timezone: 'America/New_York',
    ...overrides,
  };
}

describe('currencySymbol', () => {
  it('returns $ for USD', () => {
    expect(currencySymbol('USD')).toBe('$');
  });

  it('returns the code itself for unknown currency', () => {
    expect(currencySymbol('XYZ')).toBe('XYZ');
  });

  it('is case-insensitive', () => {
    expect(currencySymbol('eur')).toBe('€');
  });

  it('returns correct symbol for GBP', () => {
    expect(currencySymbol('GBP')).toBe('£');
  });
});

describe('fiscalYearLabel', () => {
  it('returns FY for current year when fiscal year starts in January', () => {
    const settings = makeSettings({ fiscalYearStart: '01-01' });
    expect(fiscalYearLabel(settings, 2026)).toBe('FY 2026');
  });

  it('returns next year when fiscal year starts mid-year', () => {
    const settings = makeSettings({ fiscalYearStart: '07-01' });
    expect(fiscalYearLabel(settings, 2026)).toBe('FY 2027');
  });
});

describe('timezoneDisplayName', () => {
  it('replaces underscores with spaces', () => {
    expect(timezoneDisplayName('America/New_York')).toBe('America/New York');
  });

  it('handles timezone without underscores', () => {
    expect(timezoneDisplayName('UTC')).toBe('UTC');
  });
});

describe('defaultCurrencyForDisplay', () => {
  it('returns currency with symbol', () => {
    expect(defaultCurrencyForDisplay(makeSettings({ defaultCurrency: 'USD' }))).toBe('USD ($)');
  });

  it('returns currency code twice for unknown currency', () => {
    const result = defaultCurrencyForDisplay(makeSettings({ defaultCurrency: 'XYZ' }));
    expect(result).toBe('XYZ (XYZ)');
  });
});
