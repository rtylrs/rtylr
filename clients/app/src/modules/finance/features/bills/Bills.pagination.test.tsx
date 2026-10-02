// @vitest-environment happy-dom
import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { useBillsMock } = vi.hoisted(() => ({ useBillsMock: vi.fn() }));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: { count?: number }) =>
      options?.count === undefined ? key : `${key}:${options.count}`,
  }),
}));
vi.mock('@voxire/ui', () => ({ EmptyState: () => null }));
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}));
vi.mock('./useBills', () => ({ BILLS_PAGE_SIZE: 50, useBills: useBillsMock }));
vi.mock('../../shared/hooks/queries/useBills', () => ({
  useCreateBill: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateBill: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));
vi.mock('../../shared/hooks/queries/useErpReceivedPurchaseOrders', () => ({
  useErpReceivedPurchaseOrders: () => ({ data: [] }),
}));
vi.mock('../../shared/lib/settings', () => ({
  useSettings: (selector: (state: { currency: string }) => unknown) =>
    selector({ currency: 'USD' }),
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
vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));
vi.mock('../../shared/components/Modal', () => ({
  default: ({ open, children }: { open: boolean; children?: React.ReactNode }) =>
    open ? <div>{children}</div> : null,
}));
vi.mock('../../shared/components/CurrencyInput', () => ({
  default: () => <input aria-label="amount" />,
}));

import Bills from './Bills';

const bills = Array.from({ length: 51 }, (_, index) => ({
  id: `bill-${index}`,
  supplier: '',
  supplierName: `Supplier ${index}`,
  vendor: `Supplier ${index}`,
  description: `Bill ${index}`,
  amount: 1,
  currency: 'USD',
  status: 'pending',
  dueDate: '2026-09-30',
}));

describe('Bills pagination', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useBillsMock.mockReturnValue({ data: bills, isLoading: false, isError: false });
  });

  it('pages rows without truncating the global outstanding total', () => {
    const { container } = render(<Bills />);

    expect(screen.getByText('bill-0')).toBeInTheDocument();
    expect(screen.queryByText('bill-50')).not.toBeInTheDocument();
    expect(container).toHaveTextContent('financeBills.outstanding:51');

    fireEvent.click(screen.getByRole('button', { name: 'financeCommon.next' }));

    expect(screen.getByText('Page 2')).toBeInTheDocument();
    expect(screen.getByText('bill-50')).toBeInTheDocument();
    expect(screen.queryByText('bill-0')).not.toBeInTheDocument();
    expect(container).toHaveTextContent('financeBills.outstanding:51');
  });
});
