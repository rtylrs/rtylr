import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testing/i18n';
import { createEffectiveAccess } from '../../../../../../../test/effectiveAccess';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

const mocks = vi.hoisted(() => ({
  effectiveAccess: undefined as unknown,
  approveMutate: vi.fn(),
}));

const sentOrder = {
  id: 'po1',
  supplierId: 's1',
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
  useProducts: () => ({ data: { products: [] } }),
}));

vi.mock('../../shared/components/PrintPOButton', () => ({ default: () => null }));
vi.mock('../../shared/components/ExportButton', () => ({ default: () => null }));
vi.mock('../../shared/components/PaginationControls', () => ({ default: () => null }));

import Purchasing from './Purchasing';

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
