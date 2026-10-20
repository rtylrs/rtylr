// @vitest-environment happy-dom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi, beforeAll, beforeEach } from 'vitest';
import ProfitLoss from './ProfitLoss';
import { initTestI18n } from '../../shared/testing/i18n';

const plMock = vi.hoisted(() => vi.fn());

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children?: React.ReactNode }) => <a>{children}</a>,
}));

vi.mock('../../shared/hooks/useTax', () => ({
  usePLByPeriod: () => plMock(),
}));

vi.mock('../../shared/lib/exports', () => ({
  exportProfitLossCsv: vi.fn(),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

const report = {
  period: 'month',
  revenue: 0,
  cogs: 0,
  operatingExpenses: 0,
  grossProfit: 0,
  netProfit: 0,
  expensesByCategory: [],
  pendingExpenses: 0,
  pendingExpenseCount: 0,
};

describe('ProfitLoss', () => {
  beforeEach(() => {
    plMock.mockReset();
  });

  it('says how much is waiting for approval instead of silently showing $0', () => {
    plMock.mockReturnValue({
      data: { ...report, pendingExpenses: 35, pendingExpenseCount: 1 },
      isPending: false,
    });

    render(<ProfitLoss />);

    expect(screen.getByRole('status')).toHaveTextContent(
      '1 expense ($35.00) in this period is waiting for approval and is not counted below.'
    );
    expect(screen.getByText('Review expenses')).toBeInTheDocument();
    // An empty period reads $0.00, never "-$0".
    expect(screen.queryByText(/-\$0/)).not.toBeInTheDocument();
  });

  it('lists approved expenses by category under operating expenses', () => {
    plMock.mockReturnValue({
      data: {
        ...report,
        revenue: 2000,
        operatingExpenses: 950,
        expensesByCategory: [
          { categoryId: 2, name: 'Rent', color: '#8b5cf6', amount: 900 },
          { categoryId: 1, name: 'Communications', color: '#14b8a6', amount: 40 },
          { name: '', color: '', amount: 10 },
        ],
      },
      isPending: false,
    });

    render(<ProfitLoss />);

    expect(screen.getByText('Rent')).toBeInTheDocument();
    expect(screen.getByText('$900.00')).toBeInTheDocument();
    expect(screen.getByText('Communications')).toBeInTheDocument();
    expect(screen.getByText('Uncategorized')).toBeInTheDocument();
    expect(screen.getAllByText('$1,050.00').length).toBeGreaterThan(0);
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('no longer offers a fiscal-year close that did nothing', () => {
    plMock.mockReturnValue({ data: report, isPending: false });

    render(<ProfitLoss />);

    expect(screen.queryByRole('button', { name: /fiscal year/i })).not.toBeInTheDocument();
    expect(screen.queryByText('Payment Reminders')).not.toBeInTheDocument();
  });
});
