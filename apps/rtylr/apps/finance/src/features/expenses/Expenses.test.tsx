import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
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
  expenseDate: '2026-09-01',
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

  it('keeps the monthly total and category breakdown global while paging rows', () => {
    render(<Expenses />);

    expect(screen.getAllByText(/\$35/).length).toBeGreaterThan(0);
    expect(screen.getByTitle('Travel: $10')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.next' }));

    expect(screen.getByText('Expense 26')).toBeInTheDocument();
    expect(screen.getAllByText(/\$35/).length).toBeGreaterThan(0);
  });

  it('shows a retryable failure instead of asserting a zero monthly total', () => {
    useExpensesMock.mockReturnValue({
      isLoading: false,
      isError: true,
      refetch: refetchMock,
      data: undefined,
    });

    render(<Expenses />);

    expect(screen.getByRole('alert')).toHaveTextContent('errors.generic');
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
    expect(screen.queryByText(/Monthly total/)).not.toBeInTheDocument();
    expect(screen.queryByText('$0')).not.toBeInTheDocument();
  });
});
