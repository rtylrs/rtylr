import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import ExpenseDetail from './ExpenseDetail';
import { initTestI18n } from '../../shared/testing/i18n';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const expense = {
  id: 'EXP-1',
  categoryId: null,
  description: 'Office supplies',
  amount: 42,
  currency: 'USD',
  vendor: 'Acme',
  receiptUrl: null,
  paymentMethod: 'Card',
  isTaxDeductible: true,
  taxAmount: 3.21,
  expenseDate: '2026-08-01',
  status: 'pending',
  notes: 'Keep this note',
};

const mocks = vi.hoisted(() => ({
  effectiveAccess: undefined as unknown,
  updateExpense: vi.fn(),
  refetchExpense: vi.fn(),
  expenseQuery: {
    data: undefined as unknown,
    isLoading: false,
    isError: false,
    error: null as unknown,
    isFetching: false,
  },
}));

vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: Object.assign(
    (
      selector: (state: { user: { effectiveAccess: unknown }; effectiveAccess: unknown }) => unknown
    ) =>
      selector({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    {
      getState: () => ({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    }
  ),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ updateExpense: mocks.updateExpense }),
  getOneKitErrorMessage: (error: unknown, fallback: string) =>
    error instanceof Error ? error.message : fallback,
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('../../shared/lib/settings', () => ({
  useSettings: (select: (s: { currency: string }) => unknown) => select({ currency: 'USD' }),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: 'EXP-1' }),
  useNavigate: () => vi.fn(),
}));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('./useExpenseById', () => ({
  useExpenseById: () => ({
    ...mocks.expenseQuery,
    refetch: mocks.refetchExpense,
  }),
}));

vi.mock('./useExpenses', () => ({
  useExpenseCategories: () => ({ data: [] }),
  useDeleteExpense: () => ({ mutate: vi.fn(), isPending: false }),
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
vi.mock('../../shared/components/ReceiptUpload', () => ({
  default: ({ onChange }: { onChange: (url: string | null) => void }) => (
    <button type="button" onClick={() => onChange('https://cdn.example/receipt.png')}>
      Attach receipt
    </button>
  ),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

let queryClient: QueryClient;
const renderDetail = () =>
  render(
    <QueryClientProvider client={queryClient}>
      <ExpenseDetail />
    </QueryClientProvider>
  );

describe('ExpenseDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    mocks.effectiveAccess = createEffectiveAccess('finance', {
      permissions: ['finance:expenses:view'],
    });
    mocks.updateExpense.mockResolvedValue({});
    mocks.refetchExpense.mockResolvedValue({ data: expense });
    Object.assign(mocks.expenseQuery, {
      data: expense,
      isLoading: false,
      isError: false,
      error: null,
      isFetching: false,
    });
  });

  it('hides Approve/Reject for a role without finance:expenses:approve', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', {
      permissions: ['finance:expenses:view'],
    });
    renderDetail();

    expect(screen.queryByText('Approve')).not.toBeInTheDocument();
    expect(screen.queryByText('Reject')).not.toBeInTheDocument();
  });

  it('shows Approve/Reject for admin', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', { isAdmin: true });
    renderDetail();

    expect(screen.getByText('Approve')).toBeInTheDocument();
    expect(screen.getByText('Reject')).toBeInTheDocument();
  });

  it('shows Approve/Reject for owner', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', { isOwner: true });
    renderDetail();

    expect(screen.getByText('Approve')).toBeInTheDocument();
    expect(screen.getByText('Reject')).toBeInTheDocument();
  });

  it('persists a receipt change before updating the displayed value', async () => {
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');
    renderDetail();

    fireEvent.click(screen.getByRole('button', { name: 'Attach receipt' }));

    await waitFor(() =>
      expect(mocks.updateExpense).toHaveBeenCalledWith({
        id: 'EXP-1',
        body: {
          categoryId: undefined,
          amount: 42,
          currency: 'USD',
          description: 'Office supplies',
          vendor: 'Acme',
          receiptUrl: 'https://cdn.example/receipt.png',
          paymentMethod: 'Card',
          isTaxDeductible: true,
          taxAmount: 3.21,
          notes: 'Keep this note',
          expenseDate: '2026-08-01',
        },
      })
    );
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['finance', 'expenses'] });
    expect(mocks.refetchExpense).toHaveBeenCalled();
  });

  it('shows a retry state instead of an endless skeleton when detail loading fails', () => {
    Object.assign(mocks.expenseQuery, {
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error('Expense not found'),
      isFetching: false,
    });

    renderDetail();

    expect(screen.getByRole('alert')).toHaveTextContent('Expense not found');
    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(mocks.refetchExpense).toHaveBeenCalledTimes(1);
  });
});
