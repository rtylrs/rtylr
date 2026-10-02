import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Bills from './Bills';

const { createMutateMock, updateMutateAsyncMock, useBillsMock } = vi.hoisted(() => ({
  createMutateMock: vi.fn(),
  updateMutateAsyncMock: vi.fn(),
  useBillsMock: vi.fn(),
}));

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
  useCreateBill: () => ({ mutate: createMutateMock, isPending: false }),
  useUpdateBill: () => ({ mutateAsync: updateMutateAsyncMock, isPending: false }),
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
  default: ({
    open,
    children,
    footer,
  }: {
    open: boolean;
    children?: React.ReactNode;
    footer?: React.ReactNode;
  }) =>
    open ? (
      <div>
        {children}
        {footer}
      </div>
    ) : null,
}));

const pendingBill = {
  id: '7',
  supplier: '42',
  supplierName: 'Acme Corp',
  vendor: 'Acme Corp',
  description: 'March invoice',
  amount: 125,
  currency: 'USD',
  status: 'pending',
  dueDate: '2026-09-30',
};

describe('Bills pending status', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useBillsMock.mockReturnValue({ data: [pendingBill], isLoading: false });
  });

  it('creates bills in the backend-supported pending state', () => {
    render(<Bills />);

    fireEvent.click(screen.getByRole('button', { name: 'financeBills.new' }));
    fireEvent.change(screen.getByPlaceholderText('financeBills.vendorPlaceholder'), {
      target: { value: 'Acme Corp' },
    });
    fireEvent.change(screen.getByPlaceholderText('0.00'), { target: { value: '125' } });
    fireEvent.click(screen.getByRole('button', { name: 'financeBills.create' }));

    expect(createMutateMock).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'pending', supplierName: 'Acme Corp', amount: 125 }),
      expect.any(Object)
    );
  });

  it('includes pending bills in outstanding metrics, actions, and badge styling', () => {
    const { container } = render(<Bills />);

    expect(container).toHaveTextContent('financeBills.outstanding:1');
    expect(container).toHaveTextContent('125.00');
    expect(screen.getByRole('button', { name: 'financeBills.schedule' })).toBeDisabled();
    expect(screen.getByText('pending').style.background).toBe('var(--finance-neutral-soft)');
  });
});
