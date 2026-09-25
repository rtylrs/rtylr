import { describe, it, expect } from 'vitest';
import { findExactMatches, ReconcileTransaction, ReconcileCandidate } from './reconcileHelper';

describe('reconcileHelper - findExactMatches', () => {
  it('matches transaction with exact invoice amount', () => {
    const transactions: ReconcileTransaction[] = [
      { id: 'tx-1', description: 'Payment from Acme', amount: 500 },
      { id: 'tx-2', description: 'Payment from Globex', amount: 1200 },
    ];
    const candidates: ReconcileCandidate[] = [
      { id: 'inv-101', label: 'INV-101 - Acme', amount: 500 },
      { id: 'inv-102', label: 'INV-102 - Beta', amount: 350 },
    ];

    const result = findExactMatches(transactions, candidates);
    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({
      transactionId: 'tx-1',
      candidate: { id: 'inv-101', label: 'INV-101 - Acme', amount: 500 },
    });
  });

  it('matches negative transaction amounts against positive invoice totals', () => {
    const transactions: ReconcileTransaction[] = [
      { id: 'tx-1', description: 'Stripe Payout', amount: -750 },
    ];
    const candidates: ReconcileCandidate[] = [{ id: 'inv-201', label: 'INV-201', amount: 750 }];

    const result = findExactMatches(transactions, candidates);
    expect(result).toHaveLength(1);
    expect(result[0].candidate.id).toBe('inv-201');
  });

  it('returns empty array when no amounts match', () => {
    const transactions: ReconcileTransaction[] = [
      { id: 'tx-1', description: 'Unknown', amount: 99.99 },
    ];
    const candidates: ReconcileCandidate[] = [{ id: 'inv-301', label: 'INV-301', amount: 500 }];

    const result = findExactMatches(transactions, candidates);
    expect(result).toHaveLength(0);
  });
});
