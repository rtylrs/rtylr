import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeAll, describe, expect, it, vi } from 'vitest';
import BalanceSheet from './BalanceSheet';
import { initTestI18n } from '../../shared/testing/i18n';

/**
 * finance.statement.v1.GetBalanceSheet is point-in-time only: it has no as-of
 * parameter. The page used to offer an unrestricted date picker and label the
 * unchanged current figures as historical, so the label must now reflect when
 * the data was fetched and no date control may be offered.
 */
const { balanceSheetMock } = vi.hoisted(() => ({
  balanceSheetMock: vi.fn(),
}));

vi.mock('./useBalanceSheet', () => ({
  useBalanceSheet: () => balanceSheetMock(),
}));

vi.mock('../../shared/hooks/queries/useErpInventoryValuation', () => ({
  useErpInventoryValuation: () => ({ data: null }),
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

describe('Balance sheet as-of labelling', () => {
  it('offers no date control and dates the figures from the fetch', () => {
    // Local 31 Jan 2026, so any timezone maps it to the same calendar day.
    const fetchedAt = new Date(2026, 0, 31, 9, 30).getTime();
    balanceSheetMock.mockReturnValue({
      data: {
        assets: [{ label: 'Cash', amount: 8000, type: 'asset' }],
        liabilities: [{ label: 'Accounts Payable', amount: 0, type: 'liability' }],
        equity: [{ label: 'Equity', amount: 10000, type: 'equity' }],
        totalAssets: 10000,
        totalLiabilities: 0,
        totalEquity: 10000,
      },
      isLoading: false,
      dataUpdatedAt: fetchedAt,
    });

    const { container } = render(<BalanceSheet />);

    expect(container.querySelector('input[type="date"]')).toBeNull();
    expect(screen.getByText('As of Date: 2026-01-31')).toBeInTheDocument();
  });
});
