import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import CashFlow from './CashFlow';
import { initTestI18n } from '../../shared/testing/i18n';

const { cashFlowMock } = vi.hoisted(() => ({
  cashFlowMock: vi.fn(),
}));

vi.mock('./useCashFlow', () => ({
  useCashFlow: () => cashFlowMock(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

vi.mock('../../shared/components/PeriodSelector', () => ({ default: () => null }));
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
});
