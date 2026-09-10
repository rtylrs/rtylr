import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Receiving from './Receiving';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  usePurchaseOrders: vi.fn(),
  navigate: vi.fn(),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('./usePurchaseOrders', () => ({
  usePurchaseOrders: mocks.usePurchaseOrders,
  useUpdatePurchaseOrderStatus: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../suppliers/useSuppliers', () => ({
  useSuppliers: () => ({ data: { suppliers: [] } }),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Receiving empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.usePurchaseOrders.mockReturnValue({
      data: { purchaseOrders: [] },
      isLoading: false,
    });
  });

  it('shows the illustration and empty copy, and navigates to purchasing on CTA click', () => {
    render(<Receiving />);

    const title = screen.getByText('No open orders to receive');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('All orders have been received')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Go to Purchase Orders' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/purchasing' });
  });

  it('does not show the empty state when there are open orders', () => {
    mocks.usePurchaseOrders.mockReturnValue({
      data: {
        purchaseOrders: [
          {
            id: 'po1',
            supplierId: 's1',
            status: 'sent',
            items: [],
            expectedDelivery: '2026-01-01',
          },
        ],
      },
      isLoading: false,
    });

    render(<Receiving />);

    expect(screen.queryByText('No open orders to receive')).not.toBeInTheDocument();
  });
});
