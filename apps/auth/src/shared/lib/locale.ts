/**
 * Locale and internationalisation helpers for the auth app.
 */

/**
 * Returns the best-match locale from a list of supported locales.
 * Falls back to the default when no match is found.
 */
export function resolveLocale(
  preferred: string[],
  supported: string[],
  defaultLocale = 'en'
): string {
  for (const lang of preferred) {
    const base = lang.split('-')[0];
    const exact = supported.find((s) => s.toLowerCase() === lang.toLowerCase());
    if (exact) return exact;
    const partial = supported.find((s) => s.split('-')[0].toLowerCase() === base.toLowerCase());
    if (partial) return partial;
  }
  return defaultLocale;
}

/**
 * Returns the browser's preferred languages, or an empty array in non-browser environments.
 */
export function getBrowserLocales(): string[] {
  if (typeof navigator === 'undefined') return [];
  return Array.from(navigator.languages ?? [navigator.language].filter(Boolean));
}

/**
 * Formats a number using the given locale.
 */
export function localiseNumber(value: number, locale: string): string {
  return new Intl.NumberFormat(locale).format(value);
}

/**
 * Formats a date using the given locale and options.
 */
export function localiseDate(
  date: Date | number | string,
  locale: string,
  options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'long', day: 'numeric' }
): string {
  return new Intl.DateTimeFormat(locale, options).format(new Date(date));
}

/**
 * Returns a collator for locale-aware string comparison.
 */
export function getCollator(locale: string): Intl.Collator {
  return new Intl.Collator(locale, { sensitivity: 'base' });
}
