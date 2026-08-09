// Locale and i18n helpers for the recruit public pages.
// The recruit app is a white-label SPA, so owners can target different
// locales by passing locale config through the customization object.
// These helpers pick the right format for dates, numbers, and currency.

/**
 * DEFAULT_LOCALE is used when no locale is configured.
 */
export const DEFAULT_LOCALE = 'en-US';

/**
 * formatLocaleCurrency formats a number as currency using the given
 * locale and currency code.
 */
export function formatLocaleCurrency(
  amount: number,
  currency = 'USD',
  locale = DEFAULT_LOCALE
): string {
  try {
    return amount.toLocaleString(locale, {
      style: 'currency',
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    });
  } catch {
    return `${currency} ${amount}`;
  }
}

/**
 * formatLocaleDate formats an ISO date string using the locale's
 * preferred date representation.
 */
export function formatLocaleDate(iso: string, locale = DEFAULT_LOCALE): string {
  try {
    const date = new Date(iso + 'T00:00:00Z');
    return date.toLocaleDateString(locale, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      timeZone: 'UTC',
    });
  } catch {
    return iso;
  }
}

/**
 * formatLocaleNumber formats a number with locale-specific separators.
 */
export function formatLocaleNumber(value: number, locale = DEFAULT_LOCALE): string {
  try {
    return value.toLocaleString(locale);
  } catch {
    return String(value);
  }
}

/**
 * isRTLLocale returns true for locales that use right-to-left text
 * direction (Arabic, Hebrew, etc.).
 */
export function isRTLLocale(locale: string): boolean {
  const rtlPrefixes = ['ar', 'he', 'fa', 'ur', 'yi', 'dv'];
  const prefix = locale.split('-')[0].toLowerCase();
  return rtlPrefixes.includes(prefix);
}
