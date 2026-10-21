import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import Expenses from './Expenses';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  refetch: vi.fn(),
  useExpenseCategories: vi.fn(),
  useExpenses: vi.fn(),
  approve: vi.fn(),
  reject: vi.fn(),
  canApprove: false,
  exportCsv: vi.fn(),
  navigate: vi.fn(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
  Link: ({ children }: { children?: React.ReactNode }) => <a href="#">{children}</a>,
}));
vi.mock('@voxire/ui', () => ({
  EmptyState: ({ title, description }: { title: string; description?: string }) => (
    <div>
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  ),
}));
vi.mock('../../shared/components/Skeleton', () => ({
  SkeletonRow: () => <div>loading</div>,
}));
vi.mock('../../shared/lib/settings', () => ({
  useSettings: (select: (s: { currency: string }) => unknown) => select({ currency: 'USD' }),
}));
vi.mock('../../shared/lib/financePermissions', () => ({
  canFinancePermission: () => mocks.canApprove,
}));
vi.mock('../../shared/lib/exports', () => ({ exportExpensesCsv: mocks.exportCsv }));
vi.mock('../../shared/hooks/queries/useExpenses', () => ({
  useApproveExpense: () => ({ mutate: mocks.approve, isPending: false }),
  useRejectExpense: () => ({ mutate: mocks.reject, isPending: false }),
}));
vi.mock('./useExpenses', () => ({
  useExpenses: mocks.useExpenses,
  useExpenseCategories: mocks.useExpenseCategories,
}));

const expense = (overrides: Record<string, unknown>) => ({
  id: '1',
  categoryId: 11,
  description: 'Phone number',
  amount: 35,
  currency: 'USD',
  vendor: '',
  paymentMethod: 'card',
  receiptUrl: '',
  expenseDate: '2026-10-19',
  status: 'pending',
  notes: '',
  ...overrides,
});

const withExpenses = (expenses: ReturnType<typeof expense>[]) =>
  mocks.useExpenses.mockReturnValue({
    isLoading: false,
    isError: false,
    refetch: mocks.refetch,
    data: { expenses, hasMore: false, page: 1, pageSize: expenses.length },
  });

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Expenses', () => {
  // Every figure is scoped to the month in view, which starts at today.
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-10-20T10:00:00'));
    vi.clearAllMocks();
    mocks.canApprove = false;
    mocks.useExpenseCategories.mockReturnValue({
      data: [
        { id: 11, name: 'Communications', color: '#14b8a6' },
        { id: 12, name: 'Rent', color: '#8b5cf6' },
      ],
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows the month as one consistent picture, without row ids or minus signs', () => {
    withExpenses([
      expense({ id: '1', vendor: 'Alfa' }),
      expense({
        id: '2',
        description: 'Office rent',
        categoryId: 12,
        amount: 900,
        status: 'approved',
        receiptUrl: 'https://x/r.png',
      }),
      expense({ id: '3', description: 'Taxi', amount: 500, status: 'rejected' }),
      expense({
        id: '4',
        description: 'Last month',
        amount: 70,
        expenseDate: '2026-09-30',
        status: 'approved',
      }),
    ]);

    render(<Expenses />);

    expect(screen.getByText('October 2026')).toBeInTheDocument();
    // Spent counts pending and approved, never rejected or September.
    expect(screen.getByText('$935.00')).toBeInTheDocument();
    expect(screen.getByText('2 expenses')).toBeInTheDocument();
    // The breakdown agrees with the month: Rent and Communications only.
    expect(screen.getByText('96%')).toBeInTheDocument();
    expect(screen.queryByText('Last month')).not.toBeInTheDocument();
    expect(screen.queryByText(/-\$/)).not.toBeInTheDocument();

    const row = screen.getByText('Phone number').closest('tr')!;
    expect(within(row).getByText('Alfa · Card')).toBeInTheDocument();
    expect(within(row).getByText('Pending')).toBeInTheDocument();
    expect(within(row).getByText('$35.00')).toBeInTheDocument();
    expect(within(row).queryByText('1')).not.toBeInTheDocument();
  });

  it('steps back a month and explains an empty one', () => {
    withExpenses([expense({ id: '1' })]);

    render(<Expenses />);
    fireEvent.click(screen.getByRole('button', { name: 'Next month' }));

    expect(screen.getByText('No expenses in November 2026')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'This month' }));
    expect(screen.getByText('Phone number')).toBeInTheDocument();
  });

  it('filters the table by search, status and the awaiting-approval card', () => {
    withExpenses([
      expense({ id: '1' }),
      expense({
        id: '2',
        description: 'Office rent',
        categoryId: 12,
        amount: 900,
        status: 'approved',
      }),
    ]);

    render(<Expenses />);
    fireEvent.change(screen.getByRole('searchbox', { name: 'Search expenses' }), {
      target: { value: 'rent' },
    });
    expect(screen.queryByText('Phone number')).not.toBeInTheDocument();
    expect(screen.getByText('Office rent')).toBeInTheDocument();
    expect(screen.getByText('Showing 1 of 2 expenses this month.')).toBeInTheDocument();

    fireEvent.change(screen.getByRole('searchbox', { name: 'Search expenses' }), {
      target: { value: '' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Awaiting approval/ }));
    expect(screen.getByText('Phone number')).toBeInTheDocument();
    expect(screen.queryByText('Office rent')).not.toBeInTheDocument();
  });

  it('lets an approver approve a pending expense from the list', () => {
    mocks.canApprove = true;
    withExpenses([expense({ id: '7' })]);

    render(<Expenses />);
    fireEvent.click(screen.getByRole('button', { name: 'Approve Phone number' }));

    expect(mocks.approve).toHaveBeenCalledWith('7');
    expect(mocks.navigate).not.toHaveBeenCalled();
  });

  it('exports the rows in view with readable labels', () => {
    withExpenses([expense({ id: '1', vendor: 'Alfa' })]);

    render(<Expenses />);
    fireEvent.click(screen.getByRole('button', { name: /Export CSV/ }));

    expect(mocks.exportCsv).toHaveBeenCalledWith('expenses-2026-10.csv', [
      expect.objectContaining({
        description: 'Phone number',
        vendor: 'Alfa',
        category: 'Communications',
        paymentMethod: 'Card',
        status: 'Pending',
        amount: 35,
      }),
    ]);
  });

  it('shows a retryable failure instead of a zero total', () => {
    mocks.useExpenses.mockReturnValue({
      isLoading: false,
      isError: true,
      refetch: mocks.refetch,
      data: undefined,
    });

    render(<Expenses />);

    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong');
    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(mocks.refetch).toHaveBeenCalled();
    expect(screen.queryByText('$0.00')).not.toBeInTheDocument();
  });
});
