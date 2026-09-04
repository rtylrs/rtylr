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

  it('renders subtotal and total as the major units the POS sends', () => {
    // pos_orders.subtotal and .total are DECIMAL(10,2) in onion, read through
    // sqlutil.DecimalToFloat and declared float64 on the wire, so 10 is ten
    // dollars. This test previously fed 1000 and expected $10.00, which pinned
    // the cents assumption and made a real $1,000 order read as $10.00.
    ordersData.data = [
      {
        id: '7',
        status: 'pending',
        subtotal: 10,
        total: 12.5,
        createdAt: new Date().toISOString(),
      } as unknown as { id: string; status: string },
    ];
    render(<Orders />);

    expect(screen.getByText('$10.00')).toBeInTheDocument();
    expect(screen.getByText('$12.50')).toBeInTheDocument();
  });

  it('does not divide the total by a hundred', () => {
    ordersData.data = [
      {
        id: '11',
        status: 'pending',
        subtotal: 1000,
        total: 1250,
        createdAt: new Date().toISOString(),
      } as unknown as { id: string; status: string },
    ];
    render(<Orders />);

    expect(screen.getByText('$1,000.00')).toBeInTheDocument();
    expect(screen.getByText('$1,250.00')).toBeInTheDocument();
    expect(screen.queryByText('$10.00')).not.toBeInTheDocument();
    expect(screen.queryByText('$12.50')).not.toBeInTheDocument();
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
        subtotal: 10,
        total: 25,
        createdAt: new Date().toISOString(),
        items: [{ id: 'i1', name: 'Kebab', quantity: 2, price: 10 }],
      } as unknown as { id: string; status: string },
    ];
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'View' }));

    expect(screen.getAllByText('$10.00')).toHaveLength(2);
    expect(screen.getAllByText('$25.00')).toHaveLength(2);
    // The line total is unit_price times quantity, both major units.
    expect(screen.getByText('Kebab').closest('div.flex')).toHaveTextContent('$20.00');
  });

  it('renders the detail line total without dividing by a hundred', () => {
    ordersData.data = [
      {
        id: '12',
        status: 'completed',
        subtotal: 1500,
        total: 1500,
        createdAt: new Date().toISOString(),
        items: [{ id: 'i1', name: 'Catering tray', quantity: 3, price: 500 }],
      } as unknown as { id: string; status: string },
    ];
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'View' }));

    expect(screen.getByText('Catering tray').closest('div.flex')).toHaveTextContent('$1,500.00');
  });
});
