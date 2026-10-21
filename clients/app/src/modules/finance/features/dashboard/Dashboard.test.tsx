import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import Dashboard from './Dashboard';
import { initTestI18n } from '../../shared/testing/i18n';

const { dashboardMock } = vi.hoisted(() => ({
  dashboardMock: vi.fn(),
}));

vi.mock('./useDashboard', () => ({
  useFinanceDashboard: () => dashboardMock(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('@voxire/ui', () => ({
  ErrorBoundary: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children?: React.ReactNode }) => <a href="#">{children}</a>,
}));
vi.mock('../../shared/lib/settings', () => ({
  useSettings: (select: (s: { currency: string }) => unknown) => select({ currency: 'USD' }),
}));
vi.mock('../../shared/components/OverdueInvoicesWidget', () => ({ default: () => null }));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Dashboard query states', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders a failure instead of default zero KPIs and empty-data messages', () => {
    dashboardMock.mockReturnValue({
      data: undefined,
      isPending: false,
      isError: true,
      isFetching: false,
      refetch: vi.fn(),
    });

    render(<Dashboard />);

    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong. Please try again.');
    expect(screen.queryByText('Revenue this month')).not.toBeInTheDocument();
  });

  const loaded = (overrides: Record<string, unknown>) =>
    dashboardMock.mockReturnValue({
      data: {
        totalRevenueMtd: 0,
        totalExpensesMtd: 0,
        netProfitMtd: 0,
        totalReceivable: 0,
        overdueAmount: 0,
        pendingExpenses: 0,
        pendingExpenseCount: 0,
        monthly: [],
        recentActivity: [],
        overdueInvoices: [],
        overdueInvoiceCount: 0,
        ...overrides,
      },
      isPending: false,
      isError: false,
      isFetching: false,
      refetch: vi.fn(),
    });

  it('explains why a new business sees zeros and empty sections', () => {
    loaded({ pendingExpenses: 35, pendingExpenseCount: 1 });

    render(<Dashboard />);

    expect(screen.getByRole('status')).toHaveTextContent(
      '1 expense ($35.00) is waiting for approval. The figures below count approved expenses only.'
    );
    expect(screen.getByText(/Nothing to chart yet/)).toBeInTheDocument();
    expect(screen.getByText(/No activity yet/)).toBeInTheDocument();
    // One row of figures, two decimals, no invented trend arrows.
    expect(screen.getAllByText('$0.00')).toHaveLength(4);
    expect(screen.queryByText(/vs last month/)).not.toBeInTheDocument();
  });

  it('lists recent expenses and payments with their direction', () => {
    loaded({
      monthly: [{ month: '2026-10', revenue: 2000, expenses: 35 }],
      recentActivity: [
        {
          kind: 'payment',
          id: '9',
          description: 'INV-9 · Acme',
          amount: 2000,
          date: '2026-10-18',
          status: '',
        },
        {
          kind: 'expense',
          id: '1',
          description: 'Phone number',
          amount: 35,
          date: '2026-10-19',
          status: 'approved',
        },
      ],
    });

    render(<Dashboard />);

    expect(screen.getByRole('img', { name: 'Revenue vs Expenses' })).toBeInTheDocument();
    expect(screen.getByText('INV-9 · Acme')).toBeInTheDocument();
    expect(screen.getByText('+$2,000.00')).toBeInTheDocument();
    expect(screen.getByText('Phone number')).toBeInTheDocument();
    expect(screen.getByText('Oct 19 · Approved')).toBeInTheDocument();
    expect(screen.getByText('Oct 18 · Invoice paid')).toBeInTheDocument();
  });
});
