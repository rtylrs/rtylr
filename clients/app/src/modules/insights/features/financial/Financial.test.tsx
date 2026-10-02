import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Financial from './Financial';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useCashFlow: vi.fn(),
  useExpensesByCategory: vi.fn(),
  useInvoiceAging: vi.fn(),
  useInsightsSettings: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: React.PropsWithChildren<Record<string, unknown>>) => (
      <div {...props}>{children}</div>
    ),
  },
}));

vi.mock('../../shared/hooks/queries/useInsightsFinancial', () => ({
  useCashFlow: mocks.useCashFlow,
  useExpensesByCategory: mocks.useExpensesByCategory,
  useInvoiceAging: mocks.useInvoiceAging,
}));

vi.mock('../../shared/hooks/queries/useInsightsSettings', () => ({
  useInsightsSettings: mocks.useInsightsSettings,
}));

describe('Financial empty states', () => {
  const refetchCashFlow = vi.fn();
  const refetchExpenses = vi.fn();
  const refetchAging = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useCashFlow.mockReturnValue({
      data: { series: [] },
      isLoading: false,
      isError: false,
      refetch: refetchCashFlow,
    });
    mocks.useExpensesByCategory.mockReturnValue({
      data: { categories: [] },
      isLoading: false,
      refetch: refetchExpenses,
    });
    mocks.useInvoiceAging.mockReturnValue({
      data: { buckets: [] },
      isLoading: false,
      refetch: refetchAging,
    });
    mocks.useInsightsSettings.mockReturnValue({ data: { defaultCurrency: 'USD' } });
  });

  it('shows a distinct title, description and working refresh action for expenses', () => {
    render(<Financial />);

    expect(screen.getByText('No expense data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Expenses will appear here once they are categorized and recorded.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[0]);
    expect(refetchExpenses).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for invoice aging', () => {
    render(<Financial />);

    expect(screen.getByText('No invoice aging data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Outstanding invoices will appear here once invoices are issued.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[1]);
    expect(refetchAging).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for cash flow', () => {
    render(<Financial />);

    expect(screen.getByText('No cash flow data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Cash flow will appear here once inflows and outflows are recorded.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[2]);
    expect(refetchCashFlow).toHaveBeenCalledTimes(1);
  });
});
