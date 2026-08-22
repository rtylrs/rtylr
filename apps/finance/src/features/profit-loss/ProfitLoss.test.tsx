// @vitest-environment happy-dom
import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import ProfitLoss from './ProfitLoss';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('lucide-react', () => {
  const Icon = () => null;
  return {
    Download: Icon,
    Lock: Icon,
    TrendingDown: Icon,
    TrendingUp: Icon,
  };
});

vi.mock('../../shared/hooks/useTax', () => ({
  usePLByPeriod: () => ({ data: null }),
}));

vi.mock('../../shared/lib/exports', () => ({
  exportProfitLossCsv: vi.fn(),
}));

describe('ProfitLoss close fiscal year', () => {
  it('renders an honestly disabled close-fiscal-year button with no destructive dialog', () => {
    render(<ProfitLoss />);

    const button = screen.getByRole('button', {
      name: 'finance.pl.closeFiscalYearComingSoon',
    });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute('title', 'finance.pl.closeFiscalYearHint');

    // No interactive confirm affordance may exist for the unimplemented action.
    expect(screen.queryByRole('button', { name: 'Close year' })).not.toBeInTheDocument();
  });

  it('does not render a payment reminders panel', () => {
    render(<ProfitLoss />);

    expect(screen.queryByText('Payment Reminders')).not.toBeInTheDocument();
  });
});
