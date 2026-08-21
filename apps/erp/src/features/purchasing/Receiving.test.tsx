import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Receiving from './Receiving';

const mocks = vi.hoisted(() => ({
  usePurchaseOrders: vi.fn(),
  navigate: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

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

    const title = screen.getByText('receiving.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('receiving.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'receiving.goToPurchasing' }));

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

    expect(screen.queryByText('receiving.empty.title')).not.toBeInTheDocument();
  });
});
