import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testing/i18n';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

const mocks = vi.hoisted(() => ({
  effectiveAccess: undefined as unknown,
  approveMutate: vi.fn(),
  products: [] as Array<{ id: string; name: string }>,
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
}));

const sentOrder = {
  id: 'po1',
  supplierId: 's1',
  poNumber: 'ACME-77',
  status: 'sent',
  total: 100,
  items: [],
  expectedDelivery: '2026-01-01',
};

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

vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: Object.assign(
    (
      select: (state: { user: { effectiveAccess: unknown }; effectiveAccess: unknown }) => unknown
    ) =>
      select({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    {
      getState: () => ({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    }
  ),
}));

vi.mock('./usePurchaseOrders', () => ({
  usePurchaseOrders: () => ({
    data: { purchaseOrders: [sentOrder], hasMore: false },
    isLoading: false,
  }),
  useAllPurchaseOrders: () => ({ data: [sentOrder], isLoading: false }),
  useApprovePurchaseOrder: () => ({ mutate: mocks.approveMutate, isPending: false }),
  useUpdatePurchaseOrderStatus: () => ({ mutate: vi.fn(), isPending: false }),
  useCreatePurchaseOrder: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  fetchAllPurchaseOrders: vi.fn().mockResolvedValue([]),
}));

vi.mock('../suppliers/useSuppliers', () => ({
  useSuppliers: () => ({ data: { suppliers: [] } }),
}));

vi.mock('../products/useProducts', () => ({
  useAllProducts: () => ({ data: mocks.products }),
}));

vi.mock('sonner', () => ({
  toast: { success: mocks.toastSuccess, error: mocks.toastError },
}));

vi.mock('../../shared/components/PrintPOButton', () => ({ default: () => null }));
vi.mock('../../shared/components/ExportButton', () => ({ default: () => null }));
vi.mock('../../shared/components/PaginationControls', () => ({ default: () => null }));

import Purchasing from './Purchasing';
import { useSettings } from '../../shared/lib/settingsStore';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Purchasing approval gating', () => {
  beforeEach(() => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { permissions: ['erp:purchasing:view'] });
  });

  it('hides the Approve action for an operator', () => {
    render(<Purchasing />);

    expect(screen.queryByRole('button', { name: 'Approve' })).not.toBeInTheDocument();
  });

  it('hides the Approve action for staff', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { permissions: ['erp:purchasing:view'] });
    render(<Purchasing />);

    expect(screen.queryByRole('button', { name: 'Approve' })).not.toBeInTheDocument();
  });

  it('shows the Approve action for admin', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { isAdmin: true });
    render(<Purchasing />);

    expect(screen.getByRole('button', { name: 'Approve' })).toBeInTheDocument();
  });

  it('shows the Approve action for owner', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { isOwner: true });
    render(<Purchasing />);

    expect(screen.getByRole('button', { name: 'Approve' })).toBeInTheDocument();
  });
});

describe('Purchasing list and form', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.effectiveAccess = createEffectiveAccess('erp', { isAdmin: true });
    mocks.products = [];
    useSettings.getState().setCurrency('USD');
  });

  it('shows the business PO number, translated status and business currency', () => {
    useSettings.getState().setCurrency('EUR');
    render(<Purchasing />);

    expect(screen.getByText('ACME-77')).toBeInTheDocument();
    expect(screen.getByText('€100.00')).toBeInTheDocument();
    expect(screen.queryByText('$100.00')).not.toBeInTheDocument();
    expect(screen.getAllByText('Sent').length).toBeGreaterThan(0);
  });

  it('offers every catalog product in the line-item picker, not just the first page', () => {
    mocks.products = Array.from({ length: 75 }, (_, i) => ({ id: `p${i}`, name: `Product ${i}` }));
    render(<Purchasing />);

    fireEvent.click(screen.getByRole('button', { name: 'New Order' }));
    const picker = screen.getAllByRole('combobox')[1];
    expect(within(picker).getByRole('option', { name: 'Product 74' })).toBeInTheDocument();
    expect(within(picker).getAllByRole('option')).toHaveLength(76);
  });

  it('names the order by its PO number when approval succeeds', () => {
    render(<Purchasing />);
    fireEvent.click(screen.getByRole('button', { name: 'Approve' }));

    const [id, handlers] = mocks.approveMutate.mock.calls[0];
    expect(id).toBe('po1');
    handlers.onSuccess();
    expect(mocks.toastSuccess).toHaveBeenCalledWith('ACME-77 approved');
  });
});
