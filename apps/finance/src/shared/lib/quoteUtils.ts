import { todayIso } from './date';
/**
 * Pure utility helpers for quote state and display logic.
 */

export type QuoteStatus = 'Draft' | 'Sent' | 'Accepted' | 'Rejected' | 'Expired';

/**
 * Returns true if the quote has been accepted.
 */
export function isAccepted(status: QuoteStatus): boolean {
  return status === 'Accepted';
}

/**
 * Returns true if the quote is in a terminal state (accepted, rejected, or expired).
 */
export function isTerminal(status: QuoteStatus): boolean {
  return status === 'Accepted' || status === 'Rejected' || status === 'Expired';
}

/**
 * Returns true if a quote has passed its expiry date.
 * Compares against an optional reference date (defaults to today).
 */
export function isExpired(expiryDate: string | null | undefined, today?: string): boolean {
  if (!expiryDate) return false;
  const ref = today ?? todayIso();
  return expiryDate < ref;
}

/**
 * Returns the number of days until the quote expires (negative if already expired).
 * Returns null if no expiry date is set.
 */
export function daysUntilExpiry(
  expiryDate: string | null | undefined,
  today?: string
): number | null {
  if (!expiryDate) return null;
  const ref = today ?? todayIso();
  const msPerDay = 86_400_000;
  const diff = new Date(expiryDate).getTime() - new Date(ref).getTime();
  return Math.round(diff / msPerDay);
}

/**
 * Returns a human-readable badge label for a quote status.
 */
export function getQuoteStatusLabel(status: QuoteStatus): string {
  const labels: Record<QuoteStatus, string> = {
    Draft: 'Draft',
    Sent: 'Sent',
    Accepted: 'Accepted',
    Rejected: 'Rejected',
    Expired: 'Expired',
  };
  return labels[status] ?? status;
}

/**
 * Returns the best available expiry date from a quote object
 * (checks expiryDate then expiresAt fields).
 */
export function resolveExpiryDate(quote: {
  expiryDate?: string | null;
  expiresAt?: string | null;
}): string | null {
  return quote.expiryDate ?? quote.expiresAt ?? null;
}

/**
 * Returns the conversion rate (accepted / total) as a percentage.
 * Returns 0 if total is 0.
 */
export function quoteConversionRate(total: number, accepted: number): number {
  if (total <= 0) return 0;
  return Math.round((accepted / total) * 10000) / 100;
}

/**
 * Returns a CSS color token for a quote status badge.
 */
export function getQuoteStatusColor(status: QuoteStatus): string {
  const colors: Record<QuoteStatus, string> = {
    Draft: 'gray',
    Sent: 'blue',
    Accepted: 'green',
    Rejected: 'red',
    Expired: 'orange',
  };
  return colors[status] ?? 'gray';
}
