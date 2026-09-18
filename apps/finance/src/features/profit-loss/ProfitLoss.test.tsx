// @vitest-environment happy-dom
import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi, beforeAll } from 'vitest';
import ProfitLoss from './ProfitLoss';
import { initTestI18n } from '../../shared/testing/i18n';

const toastInfo = vi.hoisted(() => vi.fn());

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
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

vi.mock('sonner', () => ({ toast: { info: toastInfo } }));

vi.mock('../../shared/hooks/useTax', () => ({
  usePLByPeriod: () => ({ data: null }),
}));

vi.mock('../../shared/components/PaymentReminderPanel', () => ({ default: () => null }));

vi.mock('../../shared/components/ConfirmDialog', () => ({
  default: ({
    open,
    onConfirm,
    confirmLabel,
  }: {
    open: boolean;
    onConfirm: () => void;
    confirmLabel: string;
  }) =>
    open ? (
      <button type="button" onClick={onConfirm}>
        {confirmLabel}
      </button>
    ) : null,
}));

vi.mock('../../shared/lib/exports', () => ({
  exportProfitLossCsv: vi.fn(),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('ProfitLoss close fiscal year', () => {
  it('uses the documented read-only toast copy after confirmation', () => {
    render(<ProfitLoss />);

    fireEvent.click(screen.getByRole('button', { name: 'Close Fiscal Year' }));
    fireEvent.click(screen.getByRole('button', { name: 'Close year' }));

    expect(toastInfo).toHaveBeenCalledWith('Fiscal year period locking is coming soon');
  });
});
