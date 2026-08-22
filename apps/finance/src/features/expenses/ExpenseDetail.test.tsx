// @vitest-environment happy-dom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import ExpenseDetail from './ExpenseDetail';

const expense: {
  id: string;
  categoryId: null;
  description: string;
  amount: number;
  currency: string;
  vendor: string;
  receiptUrl: string | null;
  expenseDate: string;
  status: string;
  notes: string | null;
} = {
  id: 'EXP-1',
  categoryId: null,
  description: 'Office supplies',
  amount: 42,
  currency: 'USD',
  vendor: 'Acme',
  receiptUrl: null,
  expenseDate: '2026-08-01',
  status: 'pending',
  notes: '',
};

const mocks = vi.hoisted(() => ({ role: 'staff' as string | null }));

vi.mock('@voxire/auth', () => ({
  useUserStore: (
    selector: (state: { user: { role: string | null }; role: string | null }) => unknown
  ) => selector({ user: { role: mocks.role }, role: mocks.role }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: 'EXP-1' }),
}));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('./useExpenseById', () => ({
  useExpenseById: () => ({ data: expense, isLoading: false }),
}));

vi.mock('./useExpenses', () => ({
  useExpenseCategories: () => ({ data: [] }),
  useUpdateExpenseReceipt: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useExpenses', () => ({
  useApproveExpense: () => ({ mutate: vi.fn(), isPending: false }),
  useRejectExpense: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

vi.mock('../../shared/components/MoneyDisplay', () => ({
  default: ({ amount }: { amount: number }) => <span>{amount}</span>,
}));

vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));
vi.mock('../../shared/components/ReceiptUpload', () => ({ default: () => null }));

describe('ExpenseDetail approval actions', () => {
  beforeEach(() => {
    mocks.role = 'staff';
  });

  it('hides Approve/Reject for a role without finance:expenses:approve', () => {
    mocks.role = 'staff';
    render(<ExpenseDetail />);

    expect(screen.queryByText('Approve')).not.toBeInTheDocument();
    expect(screen.queryByText('Reject')).not.toBeInTheDocument();
  });

  it('shows Approve/Reject for admin', () => {
    mocks.role = 'admin';
    render(<ExpenseDetail />);

    expect(screen.getByText('Approve')).toBeInTheDocument();
    expect(screen.getByText('Reject')).toBeInTheDocument();
  });

  it('shows Approve/Reject for owner', () => {
    mocks.role = 'owner';
    render(<ExpenseDetail />);

    expect(screen.getByText('Approve')).toBeInTheDocument();
    expect(screen.getByText('Reject')).toBeInTheDocument();
  });
});

describe('ExpenseDetail notes read-path', () => {
  it('hides the notes block when the expense has no notes', () => {
    render(<ExpenseDetail />);

    expect(screen.queryByText('expenses.notes')).not.toBeInTheDocument();
  });

  it('renders persisted notes under a Notes label', () => {
    expense.notes = 'Client dinner for Q3 planning';
    render(<ExpenseDetail />);

    expect(screen.getByText('expenses.notes')).toBeInTheDocument();
    expect(screen.getByText('Client dinner for Q3 planning')).toBeInTheDocument();
  });
});
