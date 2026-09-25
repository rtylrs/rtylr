// Shared formatters for prices, currency, dates, and percentages.

// Intl.NumberFormat construction is expensive; cache one instance per
// currency instead of rebuilding it on every formatted price.
const currencyFormatters = new Map<string, Intl.NumberFormat>();

function getCurrencyFormatter(currency: string): Intl.NumberFormat | null {
  let formatter = currencyFormatters.get(currency);
  if (formatter === undefined) {
    try {
      formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency });
      currencyFormatters.set(currency, formatter);
    } catch {
      return null;
    }
  }
  return formatter;
}

export function formatCurrency(amount: number, currency = 'USD'): string {
  const formatter = getCurrencyFormatter(currency);
  if (formatter) return formatter.format(amount);
  return `$${amount.toFixed(2)}`;
}

export function formatPercent(pct: number, digits = 0): string {
  return `${(pct * 100).toFixed(digits)}%`;
}

export function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export function formatTime(iso: string): string {
  try {
    return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  } catch {
    return iso;
  }
}

export function formatShortDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString();
  } catch {
    return iso;
  }
}

export function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}
