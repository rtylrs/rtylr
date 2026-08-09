import { describe, it, expect } from 'vitest';
import {
  sumDebits,
  sumCredits,
  isBalanced,
  netAmount,
  linesForAccount,
  accountBalance,
} from './journalUtils';
import type { JournalLine } from './journalUtils';

const lines: JournalLine[] = [
  { accountId: 'cash', type: 'debit', amount: 1000 },
  { accountId: 'revenue', type: 'credit', amount: 1000 },
  { accountId: 'cash', type: 'debit', amount: 500 },
  { accountId: 'expense', type: 'debit', amount: 200 },
  { accountId: 'cash', type: 'credit', amount: 200 },
];

describe('sumDebits', () => {
  it('sums all debit lines', () => {
    expect(sumDebits(lines)).toBe(1700);
  });

  it('returns 0 for empty', () => {
    expect(sumDebits([])).toBe(0);
  });
});

describe('sumCredits', () => {
  it('sums all credit lines', () => {
    expect(sumCredits(lines)).toBe(1200);
  });
});

describe('isBalanced', () => {
  it('returns false when debits and credits differ beyond tolerance', () => {
    expect(isBalanced(lines)).toBe(false);
  });

  it('returns true when balanced', () => {
    const balanced: JournalLine[] = [
      { accountId: 'cash', type: 'debit', amount: 500 },
      { accountId: 'revenue', type: 'credit', amount: 500 },
    ];
    expect(isBalanced(balanced)).toBe(true);
  });

  it('respects custom tolerance', () => {
    const almost: JournalLine[] = [
      { accountId: 'a', type: 'debit', amount: 100.005 },
      { accountId: 'b', type: 'credit', amount: 100 },
    ];
    expect(isBalanced(almost, 0.01)).toBe(true);
    expect(isBalanced(almost, 0.001)).toBe(false);
  });
});

describe('netAmount', () => {
  it('returns debits minus credits', () => {
    expect(netAmount(lines)).toBe(500);
  });
});

describe('linesForAccount', () => {
  it('filters lines by accountId', () => {
    const cashLines = linesForAccount(lines, 'cash');
    expect(cashLines).toHaveLength(3);
  });

  it('returns empty array for unknown account', () => {
    expect(linesForAccount(lines, 'unknown')).toHaveLength(0);
  });
});

describe('accountBalance', () => {
  it('returns debit minus credit for account', () => {
    expect(accountBalance(lines, 'cash')).toBe(1300);
  });

  it('returns negative when credits exceed debits', () => {
    expect(accountBalance(lines, 'revenue')).toBe(-1000);
  });

  it('returns 0 for unknown account', () => {
    expect(accountBalance(lines, 'nonexistent')).toBe(0);
  });
});
