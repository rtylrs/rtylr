import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import BalanceSheet from './BalanceSheet';
import { initTestI18n } from '../../shared/testing/i18n';

const { balanceSheetMock } = vi.hoisted(() => ({
  balanceSheetMock: vi.fn(),
}));

vi.mock('./useBalanceSheet', () => ({
  useBalanceSheet: (asOf: string) => balanceSheetMock(asOf),
}));

vi.mock('../../shared/hooks/queries/useErpInventoryValuation', () => ({
  useErpInventoryValuation: () => ({ data: null }),
}));

vi.mock('../../shared/lib/settings', () => ({
  useSettings: (select: (s: { currency: string }) => unknown) => select({ currency: 'USD' }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children?: React.ReactNode }) => <span>{children}</span>,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

const sheet = {
  asOf: '2026-10-31',
  lines: [{ accountId: '7', code: '1010', name: 'Bank', type: 'asset', amount: 465 }],
  netIncome: 465,
  totalAssets: 465,
  totalLiabilities: 0,
  totalEquity: 465,
  cashBalance: 8000,
  accountsReceivable: 300,
  accountsPayable: 0,
};

describe('Balance sheet', () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-10-31T10:00:00'));
    balanceSheetMock.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('lists ledger accounts, shows net income once inside equity, and checks the balance', () => {
    balanceSheetMock.mockReturnValue({ data: sheet, isLoading: false });

    render(<BalanceSheet />);

    expect(screen.getByText('Bank')).toBeInTheDocument();
    expect(screen.getByText('1010')).toBeInTheDocument();
    expect(screen.getAllByText('Net income to date')).toHaveLength(1);
    // Equity is a section with a labelled total, never a row repeating the total.
    expect(screen.getByText('Total equity')).toBeInTheDocument();
    expect(screen.getByText('Total assets')).toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveTextContent(
      'Balanced: assets $465.00 equal liabilities and equity $465.00.'
    );
    // Figures from other screens sit apart from the statement.
    expect(screen.getByText('Open invoices still owed to you')).toBeInTheDocument();
  });

  it('asks for the date the reader picks', () => {
    balanceSheetMock.mockReturnValue({ data: sheet, isLoading: false });

    const { container } = render(<BalanceSheet />);
    expect(balanceSheetMock).toHaveBeenLastCalledWith('2026-10-31');

    const picker = container.querySelector('input[type="date"]') as HTMLInputElement;
    fireEvent.change(picker, { target: { value: '2026-06-30' } });

    expect(balanceSheetMock).toHaveBeenLastCalledWith('2026-06-30');
  });

  it('explains an empty ledger instead of a page of zeros', () => {
    balanceSheetMock.mockReturnValue({
      data: { ...sheet, lines: [], netIncome: 0, totalAssets: 0, totalEquity: 0 },
      isLoading: false,
    });

    render(<BalanceSheet />);

    expect(screen.getByText(/Nothing is in the ledger by October 31, 2026/)).toBeInTheDocument();
  });
});
