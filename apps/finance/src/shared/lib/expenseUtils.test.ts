import { describe, it, expect } from 'vitest';
import {
  groupExpensesByCategory,
  sumExpenses,
  getBudgetUtilisation,
  isOverBudget,
  filterExpensesByDateRange,
  getTopCategories,
} from './expenseUtils';

const EXPENSES = [
  { id: '1', category: 'Travel', amount: 300, date: '2026-05-01' },
  { id: '2', category: 'Travel', amount: 200, date: '2026-05-15' },
  { id: '3', category: 'Software', amount: 100, date: '2026-05-10' },
  { id: '4', category: 'Food', amount: 50, date: '2026-06-01' },
];

describe('groupExpensesByCategory', () => {
  it('groups expenses by category', () => {
    const groups = groupExpensesByCategory(EXPENSES);
    expect(groups['Travel']).toHaveLength(2);
    expect(groups['Software']).toHaveLength(1);
  });

  it('uses Uncategorized for empty category', () => {
    const groups = groupExpensesByCategory([
      { id: '5', category: '', amount: 10, date: '2026-01-01' },
    ]);
    expect(groups['Uncategorized']).toHaveLength(1);
  });

  it('returns empty object for empty input', () => {
    expect(groupExpensesByCategory([])).toEqual({});
  });
});

describe('sumExpenses', () => {
  it('sums total amount', () => {
    expect(sumExpenses(EXPENSES)).toBe(650);
  });

  it('returns 0 for empty array', () => {
    expect(sumExpenses([])).toBe(0);
  });

  it('handles single expense', () => {
    expect(sumExpenses([{ id: '1', category: 'A', amount: 75, date: '' }])).toBe(75);
  });
});

describe('getBudgetUtilisation', () => {
  it('returns fraction of budget used', () => {
    expect(getBudgetUtilisation(500, 1000)).toBeCloseTo(0.5);
  });

  it('returns 0 when budget is 0', () => {
    expect(getBudgetUtilisation(100, 0)).toBe(0);
  });

  it('returns value above 1 when over budget', () => {
    expect(getBudgetUtilisation(1500, 1000)).toBeCloseTo(1.5);
  });
});

describe('isOverBudget', () => {
  it('returns true when spent exceeds budget', () => {
    expect(isOverBudget(1100, 1000)).toBe(true);
  });

  it('returns false when within budget', () => {
    expect(isOverBudget(900, 1000)).toBe(false);
  });

  it('returns false when exactly at budget', () => {
    expect(isOverBudget(1000, 1000)).toBe(false);
  });

  it('returns false when budget is 0', () => {
    expect(isOverBudget(500, 0)).toBe(false);
  });
});

describe('filterExpensesByDateRange', () => {
  it('returns expenses within the range', () => {
    const filtered = filterExpensesByDateRange(EXPENSES, '2026-05-01', '2026-05-31');
    expect(filtered).toHaveLength(3);
  });

  it('includes endpoints', () => {
    const filtered = filterExpensesByDateRange(EXPENSES, '2026-05-01', '2026-05-01');
    expect(filtered).toHaveLength(1);
    expect(filtered[0]?.id).toBe('1');
  });

  it('returns empty when no match', () => {
    const filtered = filterExpensesByDateRange(EXPENSES, '2025-01-01', '2025-12-31');
    expect(filtered).toHaveLength(0);
  });
});

describe('getTopCategories', () => {
  it('returns categories sorted by total spend descending', () => {
    const top = getTopCategories(EXPENSES, 2);
    expect(top[0]?.category).toBe('Travel');
    expect(top[0]?.total).toBe(500);
    expect(top[1]?.category).toBe('Software');
  });

  it('limits to N categories', () => {
    const top = getTopCategories(EXPENSES, 1);
    expect(top).toHaveLength(1);
    expect(top[0]?.category).toBe('Travel');
  });

  it('returns empty for empty expenses', () => {
    expect(getTopCategories([], 3)).toEqual([]);
  });
});
