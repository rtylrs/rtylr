import { todayIso } from './dateUtils';

/**
 * Whether a document's expiry has already passed.
 *
 * A document expiring **today** is not expired: it is valid for the whole of
 * today. Comparing a parsed date to the current instant said otherwise from
 * 00:00 UTC, so a valid document was flagged red for its entire last day
 * (rtylr#1077).
 */
export function isExpired(expiryDateIso: string): boolean {
  return expiryDateIso.split('T')[0] < todayIso();
}
