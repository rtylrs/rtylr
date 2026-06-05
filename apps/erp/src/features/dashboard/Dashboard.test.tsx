import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import Dashboard from './Dashboard';

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    children,
    to,
    className,
    style,
    ...props
  }: React.PropsWithChildren<{
    to: string;
    className?: string;
    style?: React.CSSProperties;
  }>) => (
    <a href={to} className={className} style={style} {...props}>
      {children}
    </a>
  ),
}));

vi.mock('../inventory/useInventory', () => ({
  useLowStock: () => ({
    data: [
      {
        productName: 'Espresso Beans',
        sku: 'ESP-01',
        stockQuantity: 4,
        lowStockThreshold: 8,
        unit: 'kg',
      },
    ],
    isLoading: false,
  }),
}));

vi.mock('../purchasing/usePurchaseOrders', () => ({
  usePurchaseOrders: () => ({
    data: {
      orders: [
        {
          id: 17,
          status: 'Pending Approval',
          supplier: 'North Roast',
          total: '$120.00',
          date: '2026-06-05',
        },
      ],
    },
    isLoading: false,
  }),
}));

vi.mock('../products/useProducts', () => ({
  useProducts: () => ({
    data: {
      products: [{ id: 1, name: 'Espresso Beans', salePrice: 25, stock: 12 }],
    },
    isLoading: false,
  }),
}));

vi.mock('../suppliers/useSuppliers', () => ({
  useSuppliers: () => ({
    data: { suppliers: [{ id: 1, name: 'North Roast' }] },
    isLoading: false,
  }),
}));

vi.mock('../../shared/hooks/useActivity', () => ({
  useActivity: () => ({
    data: [
      {
        id: 'a1',
        type: 'product',
        title: 'Espresso Beans updated',
        timestamp: new Date().toISOString(),
      },
    ],
  }),
}));

describe('Dashboard layout', () => {
  it('uses the full content width on large screens', () => {
    const { container } = render(<Dashboard />);

    expect(screen.getByRole('heading', { name: 'Inventory Overview' })).toBeInTheDocument();
    expect(container.firstElementChild).toHaveClass('w-full');
    expect(container.firstElementChild).toHaveClass('min-w-0');
  });
});
