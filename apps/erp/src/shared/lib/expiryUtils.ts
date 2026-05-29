/**
 * Expiry date classification helpers for batch/lot tracking.
 */

/** Number of days considered "expiring soon". */
export const EXPIRY_WARNING_DAYS = 30;

/** Number of days considered critically close to expiry. */
export const EXPIRY_CRITICAL_DAYS = 7;

export type ExpiryStatus = 'expired' | 'critical' | 'warning' | 'ok';

/**
 * Classify an expiry date relative to today.
 * Returns 'expired', 'critical', 'warning', or 'ok'.
 */
export function expiryStatus(expiryDate: string, today = new Date()): ExpiryStatus {
  const expiry = new Date(expiryDate);
  const diffMs = expiry.getTime() - today.getTime();
  const diffDays = diffMs / (1000 * 60 * 60 * 24);
  if (diffDays < 0) return 'expired';
  if (diffDays <= EXPIRY_CRITICAL_DAYS) return 'critical';
  if (diffDays <= EXPIRY_WARNING_DAYS) return 'warning';
  return 'ok';
}

/** Return true when the item is already expired. */
export function isExpired(expiryDate: string, today = new Date()): boolean {
  return expiryStatus(expiryDate, today) === 'expired';
}

/** Return true when the item needs attention (critical or warning). */
export function needsAttention(expiryDate: string, today = new Date()): boolean {
  const status = expiryStatus(expiryDate, today);
  return status === 'critical' || status === 'warning';
}

/**
 * Sort expiry items so the soonest-to-expire come first.
 * Items with invalid or missing dates sort to the end.
 */
export function sortByExpiry<T extends { expiryDate: string }>(items: T[]): T[] {
  return [...items].sort((a, b) => {
    const aMs = new Date(a.expiryDate).getTime();
    const bMs = new Date(b.expiryDate).getTime();
    if (Number.isNaN(aMs)) return 1;
    if (Number.isNaN(bMs)) return -1;
    return aMs - bMs;
  });
}
