// Offer letter helpers for the recruit admin panel.

/**
 * formatSalaryRange returns a display string like '$50,000 - $70,000 USD'.
 */
export function formatSalaryRange(
  min: number,
  max: number,
  currency: string,
  locale = 'en-US'
): string {
  const fmt = new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  });
  return `${fmt.format(min)} - ${fmt.format(max)}`;
}

/**
 * offerExpiryDays returns the number of days between now and the expiry date.
 * Returns negative when the offer has expired.
 */
export function offerExpiryDays(expiresAt: string): number {
  const now = new Date();
  const expiry = new Date(expiresAt);
  return Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

/**
 * isOfferExpired returns true when the expiry date is in the past.
 */
export function isOfferExpired(expiresAt: string): boolean {
  return new Date(expiresAt) < new Date();
}

/**
 * offerStatusLabel returns a human-readable label for offer status.
 */
export function offerStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: 'Pending Signature',
    accepted: 'Accepted',
    declined: 'Declined',
    expired: 'Expired',
    revoked: 'Revoked',
  };
  return labels[status] ?? status;
}

/**
 * annualToMonthly converts an annual salary to monthly.
 */
export function annualToMonthly(annualSalary: number): number {
  return Math.round(annualSalary / 12);
}
