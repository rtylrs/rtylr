import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import CashFlow from './CashFlow';
import { initTestI18n } from '../../shared/testing/i18n';

const { cashFlowMock } = vi.hoisted(() => ({
  cashFlowMock: vi.fn(),
}));

vi.mock('./useCashFlow', () => ({
  useCashFlow: (options: unknown) => cashFlowMock(options),
}));

vi.mock('../../shared/lib/settings', () => ({
  useSettings: (select: (s: { currency: string }) => unknown) => select({ currency: 'USD' }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

vi.mock('../../shared/components/MoneyDisplay', () => ({
  default: ({ amount }: { amount: number }) => <span>{amount}</span>,
}));
vi.mock('../../shared/components/CashFlowChart', () => ({
  default: () => <div data-testid="cash-flow-chart" />,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('CashFlow query states', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders a skeleton instead of zero-valued cards while loading', () => {
    cashFlowMock.mockReturnValue({
      data: undefined,
      isPending: true,
      isError: false,
      isFetching: true,
      refetch: vi.fn(),
    });

    render(<CashFlow />);

    expect(screen.getByRole('status')).toHaveAttribute('aria-label', 'Loading');
    expect(screen.queryByText('Total inflow')).not.toBeInTheDocument();
    expect(screen.queryByTestId('cash-flow-chart')).not.toBeInTheDocument();
  });

  it('renders a retryable failure instead of zero-valued cards', () => {
    cashFlowMock.mockReturnValue({
      data: undefined,
      isPending: false,
      isError: true,
      isFetching: false,
      refetch: vi.fn(),
    });

    render(<CashFlow />);

    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong. Please try again.');
    expect(screen.queryByText('Total inflow')).not.toBeInTheDocument();
    expect(screen.queryByTestId('cash-flow-chart')).not.toBeInTheDocument();
  });

  it('asks for the period the reader picks', () => {
    cashFlowMock.mockReturnValue({ data: [], isPending: false, isError: false });

    render(<CashFlow />);
    expect(cashFlowMock).toHaveBeenLastCalledWith({ period: 'month' });

    fireEvent.click(screen.getByRole('tab', { name: 'This quarter' }));
    expect(cashFlowMock).toHaveBeenLastCalledWith({ period: 'quarter' });
    fireEvent.click(screen.getByRole('tab', { name: 'This year' }));
    expect(cashFlowMock).toHaveBeenLastCalledWith({ period: 'year' });
  });

  it('says why the chart is empty instead of drawing a flat line', () => {
    cashFlowMock.mockReturnValue({
      data: [{ month: '2026-10-01', inflow: 0, outflow: 0 }],
      isPending: false,
      isError: false,
    });

    render(<CashFlow />);

    expect(screen.getByText(/No money in or out this month/)).toBeInTheDocument();
    expect(screen.queryByTestId('cash-flow-chart')).not.toBeInTheDocument();
  });
});
