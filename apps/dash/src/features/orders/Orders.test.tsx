import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: { status?: string }) => (opts?.status ? `${key}:${opts.status}` : key),
  }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const ordersData = vi.hoisted(() => ({
  data: [] as { id: string; status: string }[],
  isLoading: false,
  isError: false,
  refetch: vi.fn(),
}));

const updateStatusMutate = vi.hoisted(() => vi.fn());

vi.mock('../../shared/hooks/queries/useOrders', () => ({
  useOrders: () => ordersData,
  useUpdateOrderStatus: () => ({ mutate: updateStatusMutate, isPending: false }),
}));

import Orders from './Orders';

beforeEach(() => {
  vi.clearAllMocks();
  ordersData.data = [];
  ordersData.isLoading = false;
  ordersData.isError = false;
});

describe('Orders table empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there are no orders', () => {
    render(<Orders />);

    expect(screen.getByText('orders.emptyTitleAll')).toBeInTheDocument();
    expect(screen.getByText('orders.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches the orders list when the refresh call-to-action is clicked', () => {
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(ordersData.refetch).toHaveBeenCalledTimes(1);
  });

  it('renders the filtered-tab empty title with the status interpolated when a status tab is empty', () => {
    render(<Orders />);

    fireEvent.click(screen.getByRole('tab', { name: /Pending/i }));

    expect(screen.getByText('orders.emptyTitleFiltered:pending')).toBeInTheDocument();
  });

  it('does not show the empty state when there are orders', () => {
    ordersData.data = [
      {
        id: '1',
        status: 'pending',
        subtotal: 10,
        total: 11,
        createdAt: new Date().toISOString(),
      } as unknown as { id: string; status: string },
    ];
    render(<Orders />);

    expect(screen.queryByText('orders.emptyTitleAll')).not.toBeInTheDocument();
  });

  it('renders minor-unit subtotal and total through the canonical display formatter', () => {
    ordersData.data = [
      {
        id: '7',
        status: 'pending',
        subtotal: 1000,
        total: 1250,
        createdAt: new Date().toISOString(),
      } as unknown as { id: string; status: string },
    ];
    render(<Orders />);

    expect(screen.getByText('$10.00')).toBeInTheDocument();
    expect(screen.getByText('$12.50')).toBeInTheDocument();
  });

  it('falls back to the raw status label in the detail dialog for unknown statuses', () => {
    ordersData.data = [
      {
        id: '9',
        status: 'refunded',
        subtotal: 500,
        total: 500,
        createdAt: new Date().toISOString(),
      } as unknown as { id: string; status: string },
    ];
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'View' }));

    // Table badge and dialog badge both fall back to the raw status label
    // instead of rendering an empty badge.
    expect(screen.getAllByText('refunded')).toHaveLength(2);
  });

  it('shows formatted totals and line items in the detail dialog', () => {
    ordersData.data = [
      {
        id: '8',
        status: 'completed',
        subtotal: 1000,
        total: 2500,
        createdAt: new Date().toISOString(),
        items: [{ id: 'i1', name: 'Kebab', quantity: 2, price: 1000 }],
      } as unknown as { id: string; status: string },
    ];
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'View' }));

    expect(screen.getAllByText('$10.00')).toHaveLength(2);
    expect(screen.getAllByText('$25.00')).toHaveLength(2);
    expect(screen.getByText('Kebab').closest('div.flex')).toHaveTextContent('20.00');
  });
});
