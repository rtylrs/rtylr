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

  it('never binds two transactions to the same candidate', () => {
    const transactions: ReconcileTransaction[] = [
      { id: 'tx-1', description: 'Duplicate deposit', amount: 500 },
      { id: 'tx-2', description: 'Duplicate deposit', amount: 500 },
    ];
    const candidates: ReconcileCandidate[] = [{ id: 'inv-401', label: 'INV-401', amount: 500 }];

    const result = findExactMatches(transactions, candidates);
    expect(result).toHaveLength(1);
    expect(result[0].transactionId).toBe('tx-1');
  });

  it('excludes candidates already used by earlier manual matches', () => {
    const transactions: ReconcileTransaction[] = [
      { id: 'tx-1', description: 'Payment from Acme', amount: 500 },
    ];
    const candidates: ReconcileCandidate[] = [
      { id: 'inv-101', label: 'INV-101 - Acme', amount: 500 },
    ];

    const result = findExactMatches(transactions, candidates, new Set(['inv-101']));
    expect(result).toHaveLength(0);
  });

  it('treats amounts within one cent as exact and beyond that as different', () => {
    const transactions: ReconcileTransaction[] = [
      { id: 'tx-1', description: 'Rounding cent', amount: 499.995 },
      { id: 'tx-2', description: 'Two cents off', amount: 500.02 },
    ];
    const candidates: ReconcileCandidate[] = [
      { id: 'inv-501', label: 'INV-501 - Cent', amount: 500 },
      { id: 'inv-502', label: 'INV-502 - Cents', amount: 500 },
    ];

    const result = findExactMatches(transactions, candidates);
    // tx-1 (within $0.01 of both) takes inv-501; tx-2 ($0.02 off) matches nothing.
    expect(result).toHaveLength(1);
    expect(result[0]).toMatchObject({ transactionId: 'tx-1', candidate: { id: 'inv-501' } });
  });
});
