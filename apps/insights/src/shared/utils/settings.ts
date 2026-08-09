// Settings helpers for the insights app configuration.

import type { InsightsSettingsResponse } from '../hooks/queries/types';

/**
 * currencySymbol returns the symbol for a given ISO currency code.
 * Falls back to the code itself for unknown currencies.
 */
export function currencySymbol(currencyCode: string): string {
  const symbols: Record<string, string> = {
    USD: '$',
    EUR: '€',
    GBP: '£',
    JPY: '¥',
    CAD: 'CA$',
    AUD: 'A$',
    CHF: 'Fr',
    CNY: '¥',
    INR: '₹',
    MXN: 'MX$',
  };
  return symbols[currencyCode.toUpperCase()] ?? currencyCode;
}

/**
 * fiscalYearLabel returns a formatted fiscal year string, e.g. 'FY 2026'.
 */
export function fiscalYearLabel(settings: InsightsSettingsResponse, calendarYear: number): string {
  const [month] = settings.fiscalYearStart.split('-').map(Number);
  const fy = month > 1 ? calendarYear + 1 : calendarYear;
  return `FY ${fy}`;
}

/**
 * timezoneOffset returns the UTC offset string for display purposes.
 * E.g. 'America/New_York' -> '(UTC)' placeholder (full tz conversion requires a lib).
 */
export function timezoneDisplayName(timezone: string): string {
  return timezone.replace(/_/g, ' ');
}

/**
 * defaultCurrencyForDisplay returns a display string like 'USD ($)'.
 */
export function defaultCurrencyForDisplay(settings: InsightsSettingsResponse): string {
  const symbol = currencySymbol(settings.defaultCurrency);
  return `${settings.defaultCurrency} (${symbol})`;
}
