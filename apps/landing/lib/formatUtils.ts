/**
 * Display formatting helpers for the landing app.
 * Converts raw values into user-facing strings: numbers, dates, durations,
 * byte sizes, and abbreviated counts.
 */

/**
 * Formats a number as a locale-aware currency string.
 * Defaults to USD but callers can pass any ISO 4217 currency code.
 */
export function formatCurrency(amount: number, currency = 'USD', locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
}

/**
 * Formats a number with thousands separators and an optional decimal count.
 */
export function formatNumber(value: number, decimals = 0, locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value);
}

/**
 * Abbreviates large numbers: 1200 -> "1.2K", 2500000 -> "2.5M".
 * Values below 1000 are returned as-is.
 */
export function abbreviateNumber(value: number): string {
  const abs = Math.abs(value);
  const sign = value < 0 ? '-' : '';
  if (abs >= 1_000_000_000) {
    return `${sign}${(abs / 1_000_000_000).toFixed(1).replace(/\.0$/, '')}B`;
  }
  if (abs >= 1_000_000) {
    return `${sign}${(abs / 1_000_000).toFixed(1).replace(/\.0$/, '')}M`;
  }
  if (abs >= 1_000) {
    return `${sign}${(abs / 1_000).toFixed(1).replace(/\.0$/, '')}K`;
  }
  return `${value}`;
}

/**
 * Formats a duration in seconds as a human-readable string:
 * < 60s -> "42s", < 3600s -> "5m 30s", >= 3600s -> "1h 5m".
 */
export function formatDuration(seconds: number): string {
  if (seconds < 0) return '0s';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  if (h > 0) return m > 0 ? `${h}h ${m}m` : `${h}h`;
  if (m > 0) return s > 0 ? `${m}m ${s}s` : `${m}m`;
  return `${s}s`;
}

/**
 * Formats a byte count as a human-readable file size.
 * formatBytes(1536) -> "1.5 KB"
 */
export function formatBytes(bytes: number): string {
  if (bytes < 0) return '0 B';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

/**
 * Converts a decimal fraction to a percentage string.
 * formatPercent(0.1234) -> "12.3%"
 */
export function formatPercent(fraction: number, decimals = 1): string {
  return `${(fraction * 100).toFixed(decimals)}%`;
}

/**
 * Returns an ordinal suffix for an integer: 1 -> "st", 2 -> "nd", 3 -> "rd", 4+ -> "th".
 */
export function ordinalSuffix(n: number): string {
  const abs = Math.abs(n);
  const mod100 = abs % 100;
  if (mod100 >= 11 && mod100 <= 13) return 'th';
  switch (abs % 10) {
    case 1:
      return 'st';
    case 2:
      return 'nd';
    case 3:
      return 'rd';
    default:
      return 'th';
  }
}

/**
 * Formats an integer with its ordinal suffix: 1 -> "1st", 22 -> "22nd".
 */
export function formatOrdinal(n: number): string {
  return `${n}${ordinalSuffix(n)}`;
}
