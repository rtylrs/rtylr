import { describe, it, expect } from 'vitest';
import {
  offerStatusLabel,
  isOfferActive,
  isOfferFinal,
  daysUntilExpiry,
  acceptanceRate,
  filterOffersByStatus,
} from './offer';
import type { Offer } from './offer';

function makeOffer(
  status: 'draft' | 'pending_approval' | 'sent' | 'accepted' | 'declined' | 'expired',
  expiresAt?: string
): Offer {
  return {
    id: 'o1',
    candidateId: 'c1',
    positionTitle: 'Engineer',
    baseSalary: 95000,
    currency: 'USD',
    status,
    expiresAt,
  };
}

const offers = [
  makeOffer('accepted'),
  makeOffer('accepted'),
  makeOffer('declined'),
  makeOffer('sent'),
  makeOffer('draft'),
];

describe('offerStatusLabel', () => {
  it('returns display labels for known statuses', () => {
    expect(offerStatusLabel('accepted')).toBe('Accepted');
    expect(offerStatusLabel('pending_approval')).toBe('Pending Approval');
  });

  it('returns raw value for unknown status', () => {
    expect(offerStatusLabel('withdrawn')).toBe('withdrawn');
  });
});

describe('isOfferActive', () => {
  it('returns true for sent offers', () => {
    expect(isOfferActive(makeOffer('sent'))).toBe(true);
  });

  it('returns false for other statuses', () => {
    expect(isOfferActive(makeOffer('accepted'))).toBe(false);
  });
});

describe('isOfferFinal', () => {
  it('returns true for terminal statuses', () => {
    expect(isOfferFinal(makeOffer('accepted'))).toBe(true);
    expect(isOfferFinal(makeOffer('declined'))).toBe(true);
    expect(isOfferFinal(makeOffer('expired'))).toBe(true);
  });

  it('returns false for non-terminal statuses', () => {
    expect(isOfferFinal(makeOffer('sent'))).toBe(false);
  });
});

describe('daysUntilExpiry', () => {
  it('returns days until expiry', () => {
    expect(daysUntilExpiry(makeOffer('sent', '2026-06-10'), '2026-06-01')).toBe(9);
  });

  it('returns null when expiresAt is not set', () => {
    expect(daysUntilExpiry(makeOffer('sent'), '2026-06-01')).toBeNull();
  });
});

describe('acceptanceRate', () => {
  it('calculates acceptance rate from finalized offers', () => {
    expect(acceptanceRate(offers)).toBeCloseTo(66.7, 0);
  });

  it('returns 0 for empty array', () => {
    expect(acceptanceRate([])).toBe(0);
  });
});

describe('filterOffersByStatus', () => {
  it('filters offers by status', () => {
    expect(filterOffersByStatus(offers, 'accepted')).toHaveLength(2);
    expect(filterOffersByStatus(offers, 'draft')).toHaveLength(1);
  });
});
