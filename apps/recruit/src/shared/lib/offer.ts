// Offer management helpers for the recruit admin panel.

export type OfferStatus =
  'draft' | 'pending_approval' | 'sent' | 'accepted' | 'declined' | 'expired';

export interface Offer {
  id: string;
  candidateId: string;
  positionTitle: string;
  baseSalary: number;
  currency: string;
  status: OfferStatus;
  sentAt?: string;
  expiresAt?: string;
  respondedAt?: string;
}

/**
 * offerStatusLabel returns a display-friendly label for an offer status.
 */
export function offerStatusLabel(status: OfferStatus | string): string {
  const labels: Record<string, string> = {
    draft: 'Draft',
    pending_approval: 'Pending Approval',
    sent: 'Sent',
    accepted: 'Accepted',
    declined: 'Declined',
    expired: 'Expired',
  };
  return labels[status] ?? status;
}

/**
 * isOfferActive returns true when the offer has been sent but not yet responded to.
 */
export function isOfferActive(offer: Offer): boolean {
  return offer.status === 'sent';
}

/**
 * isOfferFinal returns true when the offer has a terminal status.
 */
export function isOfferFinal(offer: Offer): boolean {
  return offer.status === 'accepted' || offer.status === 'declined' || offer.status === 'expired';
}

/**
 * daysUntilExpiry returns the number of days until an offer expires.
 * Returns null when expiresAt is not set.
 */
export function daysUntilExpiry(offer: Offer, referenceDate: string): number | null {
  if (!offer.expiresAt) return null;
  const ref = new Date(referenceDate).getTime();
  const exp = new Date(offer.expiresAt).getTime();
  return Math.ceil((exp - ref) / (1000 * 60 * 60 * 24));
}

/**
 * acceptanceRate returns the percentage of sent offers that were accepted.
 */
export function acceptanceRate(offers: Offer[]): number {
  const sent = offers.filter(
    (o) => o.status === 'accepted' || o.status === 'declined' || o.status === 'expired'
  );
  if (sent.length === 0) return 0;
  const accepted = offers.filter((o) => o.status === 'accepted').length;
  return Math.round((accepted / sent.length) * 1000) / 10;
}

/**
 * filterOffersByStatus returns offers matching a given status.
 */
export function filterOffersByStatus(offers: Offer[], status: OfferStatus): Offer[] {
  return offers.filter((o) => o.status === status);
}
