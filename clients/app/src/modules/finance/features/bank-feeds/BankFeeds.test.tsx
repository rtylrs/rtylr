// @vitest-environment happy-dom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { bankTransactionsMock, bankAccountsMock } = vi.hoisted(() => ({
  bankTransactionsMock: vi.fn(),
  bankAccountsMock: vi.fn(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: { user: null; role: string }) => unknown) =>
    selector({ user: null, role: 'viewer' }),
}));

vi.mock('../../shared/lib/financePermissions', () => ({
  canFinancePermission: () => false,
}));

vi.mock('../../shared/hooks/useBankAccounts', () => ({
  useBankAccounts: () => bankAccountsMock(),
  useBankTransactions: (...args: unknown[]) => bankTransactionsMock(...args),
  useCreateBankAccount: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ subtitle, actions }: { subtitle?: React.ReactNode; actions?: React.ReactNode }) => (
    <div>
      <div>{subtitle}</div>
      {actions}
    </div>
  ),
}));

vi.mock('../../shared/components/MoneyDisplay', () => ({
  default: ({ amount }: { amount?: number }) => <span>{(amount ?? 0).toFixed(2)}</span>,
}));

vi.mock('../../shared/components/Skeleton', () => ({
  SkeletonRow: () => <div role="status" aria-label="loading" />,
}));

vi.mock('../../shared/components/Modal', () => ({
  default: ({ open, children }: { open: boolean; children?: React.ReactNode }) =>
    open ? <div>{children}</div> : null,
}));

import BankFeeds from './BankFeeds';

const oneAccount = {
  data: [
    { id: '1', name: 'Main', type: 'checking', balance: 100, institution: 'Bank', mask: '0001' },
  ],
  isLoading: false,
  isError: false,
};

describe('BankFeeds transaction states', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    bankAccountsMock.mockReturnValue(oneAccount);
  });

  it('claims no account count while the accounts are still loading', () => {
    bankAccountsMock.mockReturnValue({ data: undefined, isLoading: true, isError: false });
    bankTransactionsMock.mockReturnValue({ data: [], isLoading: false, isError: false });

    render(<BankFeeds />);

    expect(screen.queryByText(/financeBankAccounts.accountCount/)).not.toBeInTheDocument();
  });

  it('explains an empty account list instead of an empty grid', () => {
    bankAccountsMock.mockReturnValue({ data: [], isLoading: false, isError: false });
    bankTransactionsMock.mockReturnValue({ data: [], isLoading: false, isError: false });

    render(<BankFeeds />);

    expect(screen.getByText('financeBankAccounts.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('financeBankAccounts.emptyDescription')).toBeInTheDocument();
  });

  it('shows the list failure rather than zero accounts', () => {
    bankAccountsMock.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: vi.fn(),
    });
    bankTransactionsMock.mockReturnValue({ data: [], isLoading: false, isError: false });

    render(<BankFeeds />);

    expect(screen.getByRole('alert')).toHaveTextContent('financeBankAccounts.loadFailed');
    expect(screen.queryByText(/accountCount/)).not.toBeInTheDocument();
  });

  it('renders a retryable error instead of an empty transaction table', () => {
    const refetch = vi.fn();
    bankTransactionsMock.mockReturnValue({
      data: [],
      isLoading: false,
      isError: true,
      refetch,
    });

    render(<BankFeeds />);

    expect(screen.getByRole('alert')).toHaveTextContent('financeErrors.generic');
    expect(screen.queryByText('financeCommon.noData')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'financeCommon.refresh' })).toBeInTheDocument();
  });

  it('renders normalized bank-feed transactions returned by the hook', () => {
    bankTransactionsMock.mockReturnValue({
      data: [
        {
          id: 'tx-1',
          matched: true,
          description: 'Acme payment',
          date: '2026-09-20',
          category: null,
          amount: 12.5,
        },
      ],
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });

    render(<BankFeeds />);

    expect(screen.getByText('Acme payment')).toBeInTheDocument();
    expect(screen.getByText('2026-09-20')).toBeInTheDocument();
    expect(screen.getByText('12.50')).toBeInTheDocument();
    expect(screen.getByText('Matched')).toBeInTheDocument();
  });
});
