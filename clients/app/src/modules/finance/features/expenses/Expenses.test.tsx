import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Expenses from './Expenses';

const { refetchMock, useExpenseCategoriesMock, useExpensesMock } = vi.hoisted(() => ({
  refetchMock: vi.fn(),
  useExpenseCategoriesMock: vi.fn(),
  useExpensesMock: vi.fn(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => vi.fn() }));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('../../shared/lib/pageTitle', () => ({ pageTitle: (key: string) => key }));
vi.mock('../../shared/components/Skeleton', () => ({
  SkeletonRow: () => <div>loading</div>,
}));
vi.mock('@voxire/api', () => ({
  safeRatio: (value: number, total: number) => (total === 0 ? 0 : value / total),
  toFiniteNumber: (value: unknown) => Number(value ?? 0),
}));
vi.mock('./useExpenses', () => ({
  useExpenses: useExpensesMock,
  useExpenseCategories: useExpenseCategoriesMock,
}));

const allExpenses = Array.from({ length: 30 }, (_, index) => ({
  id: String(index + 1),
  categoryId: index < 25 ? 1 : 2,
  description: `Expense ${index + 1}`,
  amount: index < 25 ? 1 : 2,
  currency: 'USD' as const,
  vendor: '',
  receiptUrl: '',
  // Five of the office rows sit in a previous month. They are the only rows a
  // month-scoped total must exclude, and they are what makes the assertion
  // below a regression test rather than a restatement of the fixture.
  expenseDate: index >= 20 && index < 25 ? '2026-08-15' : '2026-09-01',
  status: 'approved',
  createdAt: '',
  updatedAt: '',
}));

const resolvedQuery = {
  isLoading: false,
  isError: false,
  refetch: refetchMock,
};

describe('Expenses aggregates and load states', () => {
  // The total is scoped to the month the reader is standing in, so the system
  // clock decides which rows count. Pin it inside the fixture's month.
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-09-15T10:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  beforeEach(() => {
    vi.clearAllMocks();
    useExpenseCategoriesMock.mockReturnValue({
      data: [
        { id: 1, name: 'Office', color: '' },
        { id: 2, name: 'Travel', color: '' },
      ],
    });
    useExpensesMock.mockImplementation(
      ({ page = 1, pageSize = 25 }: { page?: number; pageSize?: number }) => {
        if (pageSize > 25) {
          return {
            ...resolvedQuery,
            data: { expenses: allExpenses, hasMore: false, page: 1, pageSize },
          };
        }
        const start = (page - 1) * pageSize;
        return {
          ...resolvedQuery,
          data: {
            expenses: allExpenses.slice(start, start + pageSize),
            hasMore: start + pageSize < allExpenses.length,
            page,
            pageSize,
          },
        };
      }
    );
  });

  it('scopes the monthly total to the current month while paging rows', () => {
    render(<Expenses />);

    // 20 office rows in September ($20) plus 5 travel rows ($10). The five
    // August rows add $5 to the ledger but must not reach a figure labelled
    // "Monthly total" in all three locales.
    expect(screen.getByText('$30')).toBeInTheDocument();
    expect(screen.queryByText('$35')).not.toBeInTheDocument();
    expect(screen.getByTitle('Travel: $10')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'financeCommon.next' }));

    expect(screen.getByText('Expense 26')).toBeInTheDocument();
    expect(screen.getByText('$30')).toBeInTheDocument();
  });

  it('shows a retryable failure instead of asserting a zero monthly total', () => {
    useExpensesMock.mockReturnValue({
      isLoading: false,
      isError: true,
      refetch: refetchMock,
      data: undefined,
    });

    render(<Expenses />);

    expect(screen.getByRole('alert')).toHaveTextContent('financeErrors.generic');
    expect(screen.getByRole('button', { name: 'financeCommon.refresh' })).toBeInTheDocument();
    expect(screen.queryByText(/Monthly total/)).not.toBeInTheDocument();
    expect(screen.queryByText('$0')).not.toBeInTheDocument();
  });
});
