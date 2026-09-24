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

type Order = {
  id: string;
  status: string;
  subtotal: number;
  total: number;
  createdAt: string;
  customerName?: string;
  customerEmail?: string;
  items?: Array<{
    id?: string;
    name: string;
    quantity: number;
    price: number;
    total?: number;
  }>;
};

type OrderPage = {
  orders: Order[];
  total: number;
  hasMore: boolean;
  page: number;
  pageSize: number;
};

const ordersData = vi.hoisted(() => ({
  data: { orders: [], total: 0, hasMore: false, page: 1, pageSize: 50 } as OrderPage,
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

const setOrders = (orders: Order[], metadata: Partial<OrderPage> = {}) => {
  ordersData.data = {
    orders,
    total: metadata.total ?? orders.length,
    hasMore: metadata.hasMore ?? false,
    page: metadata.page ?? 1,
    pageSize: metadata.pageSize ?? 50,
  };
};

beforeEach(() => {
  vi.clearAllMocks();
  setOrders([]);
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
    setOrders([
      {
        id: '1',
        status: 'pending',
        subtotal: 10,
        total: 11,
        createdAt: new Date().toISOString(),
      },
    ]);
    render(<Orders />);

    expect(screen.queryByText('orders.emptyTitleAll')).not.toBeInTheDocument();
  });

  it('renders subtotal and total as the major units the POS sends', () => {
    setOrders([
      {
        id: '7',
        status: 'pending',
        subtotal: 10,
        total: 12.5,
        createdAt: new Date().toISOString(),
      },
    ]);
    render(<Orders />);

    expect(screen.getByText('$10.00')).toBeInTheDocument();
    expect(screen.getByText('$12.50')).toBeInTheDocument();
  });

  it('does not divide the total by a hundred', () => {
    setOrders([
      {
        id: '11',
        status: 'pending',
        subtotal: 1000,
        total: 1250,
        createdAt: new Date().toISOString(),
      },
    ]);
    render(<Orders />);

    expect(screen.getByText('$1,000.00')).toBeInTheDocument();
    expect(screen.getByText('$1,250.00')).toBeInTheDocument();
    expect(screen.queryByText('$10.00')).not.toBeInTheDocument();
    expect(screen.queryByText('$12.50')).not.toBeInTheDocument();
  });

  it('falls back to the raw status label in the detail dialog for unknown statuses', () => {
    setOrders([
      {
        id: '9',
        status: 'refunded',
        subtotal: 500,
        total: 500,
        createdAt: new Date().toISOString(),
      },
    ]);
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'View' }));

    expect(screen.getAllByText('refunded')).toHaveLength(2);
  });

  it('renders the detail line total without dividing by a hundred', () => {
    setOrders([
      {
        id: '8',
        status: 'completed',
        subtotal: 10,
        total: 25,
        createdAt: new Date().toISOString(),
        items: [{ id: 'i1', name: 'Kebab', quantity: 2, price: 10 }],
      },
    ]);
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'View' }));

    expect(screen.getAllByText('$10.00')).toHaveLength(2);
    expect(screen.getAllByText('$25.00')).toHaveLength(2);
    expect(screen.getByText('Kebab').closest('div.flex')).toHaveTextContent('$20.00');
  });

  it('uses the persisted line total for discounted or taxed items', () => {
    setOrders([
      {
        id: '12',
        status: 'completed',
        subtotal: 9,
        total: 9.9,
        createdAt: new Date().toISOString(),
        items: [{ id: 'i1', name: 'Discounted item', quantity: 1, price: 10, total: 9 }],
      },
    ]);
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'View' }));

    expect(screen.getByText('Discounted item').closest('div.flex')).toHaveTextContent('$9.00');
  });

  it('renders the detail line total without dividing by a hundred', () => {
    setOrders([
      {
        id: '13',
        status: 'completed',
        subtotal: 1500,
        total: 1500,
        createdAt: new Date().toISOString(),
        items: [{ id: 'i1', name: 'Catering tray', quantity: 3, price: 500 }],
      },
    ]);
    render(<Orders />);

    fireEvent.click(screen.getByRole('button', { name: 'View' }));

    expect(screen.getByText('Catering tray').closest('div.flex')).toHaveTextContent('$1,500.00');
  });

  it('uses the server continuation flag instead of guessing from a full page', () => {
    setOrders(
      Array.from({ length: 50 }, (_, index) => ({
        id: String(index + 1),
        status: 'pending',
        subtotal: 1,
        total: 1,
        createdAt: new Date().toISOString(),
      })),
      { hasMore: true, total: 51 }
    );
    render(<Orders />);

    expect(screen.getByRole('button', { name: 'common.pagination.next' })).toBeEnabled();
  });
});
