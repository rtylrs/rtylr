/**
 * Pure utility helpers for expense aggregation and budget tracking.
 */

export interface Expense {
  id: string;
  category: string;
  amount: number;
  date: string;
}

/**
 * Groups an array of expenses by their category.
 * Returns a map of category name to list of expenses.
 */
export function groupExpensesByCategory(expenses: Expense[]): Record<string, Expense[]> {
  return expenses.reduce<Record<string, Expense[]>>((acc, expense) => {
    const key = expense.category || 'Uncategorized';
    if (!acc[key]) acc[key] = [];
    acc[key].push(expense);
    return acc;
  }, {});
}

/**
 * Sums the total amount across all expenses.
 */
export function sumExpenses(expenses: Expense[]): number {
  return expenses.reduce((total, e) => total + e.amount, 0);
}

/**
 * Returns the fraction of budget consumed (0-1+). Returns 0 if budget is 0.
 */
export function getBudgetUtilisation(spent: number, budget: number): number {
  if (budget <= 0) return 0;
  return spent / budget;
}

/**
 * Returns true when spent exceeds the budget.
 */
export function isOverBudget(spent: number, budget: number): boolean {
  return budget > 0 && spent > budget;
}

/**
 * Filters expenses to those within a date range (inclusive on both ends).
 * Dates are ISO strings (YYYY-MM-DD).
 */
export function filterExpensesByDateRange(
  expenses: Expense[],
  from: string,
  to: string
): Expense[] {
  return expenses.filter((e) => e.date >= from && e.date <= to);
}

/**
 * Returns the top N categories by total spend, sorted descending.
 */
export function getTopCategories(
  expenses: Expense[],
  n: number
): Array<{ category: string; total: number }> {
  const grouped = groupExpensesByCategory(expenses);
  return Object.entries(grouped)
    .map(([category, items]) => ({ category, total: sumExpenses(items) }))
    .sort((a, b) => b.total - a.total)
    .slice(0, n);
}

/**
 * Returns the percentage change between current and prior period spend.
 * Returns 0 if prior is 0.
 */
export function getExpenseGrowthPct(current: number, prior: number): number {
  if (prior === 0) return 0;
  return Math.round(((current - prior) / prior) * 10000) / 100;
}

/**
 * Sorts an array of expense-like objects by amount.
 * Direction defaults to descending (highest first).
 */
export function sortExpensesByAmount<T extends { amount: number }>(
  expenses: T[],
  direction: 'asc' | 'desc' = 'desc'
): T[] {
  return [...expenses].sort((a, b) =>
    direction === 'desc' ? b.amount - a.amount : a.amount - b.amount
  );
}
