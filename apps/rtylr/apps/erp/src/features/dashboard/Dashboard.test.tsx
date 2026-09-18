import React from 'react';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeAll, describe, expect, it, vi } from 'vitest';
import Dashboard from './Dashboard';
import { initTestI18n } from '../../shared/testing/i18n';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

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

vi.mock('../../shared/hooks/queries/useDashboardStats', () => ({
  useDashboardStats: () => ({
    data: {
      totalProducts: 234,
      lowStockCount: 1,
      pendingOrders: 1,
      totalSuppliers: 1,
      inventoryValue: 245561.2,
      lowStockItems: [
        {
          productId: 1,
          name: 'Espresso Beans',
          sku: 'ESP-01',
          stock: 4,
          minStock: 8,
          unit: 'kg',
          warehouseName: '',
        },
      ],
      recentOrders: [
        {
          id: 17,
          poNumber: 'PO-17',
          status: 'Pending Approval',
          supplier: 'North Roast',
          total: 120,
          date: '2026-06-05',
        },
      ],
      topCategories: [],
    },
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
  useCreateProduct: () => ({
    isPending: false,
    mutate: vi.fn(),
  }),
  useUpdateProduct: () => ({
    isPending: false,
    mutate: vi.fn(),
  }),
  useProductAttributes: () => ({
    data: undefined,
    isLoading: false,
  }),
  useProductFull: () => ({
    data: undefined,
    isLoading: false,
  }),
}));

vi.mock('../suppliers/useSuppliers', () => ({
  useSuppliers: () => ({
    data: { suppliers: [{ id: 1, name: 'North Roast' }] },
    isLoading: false,
  }),
}));

vi.mock('../categories/useCategories', () => ({
  useCategories: () => ({
    data: [],
    isLoading: false,
  }),
}));

vi.mock('../brands/useBrands', () => ({
  useBrands: () => ({
    data: [],
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

function renderDashboard() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <Dashboard />
    </QueryClientProvider>
  );
}

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Dashboard layout', () => {
  it('uses the full content width on large screens', () => {
    const { container } = renderDashboard();

    expect(screen.getByRole('heading', { name: 'Inventory Overview' })).toBeInTheDocument();
    expect(screen.getByText('234')).toBeInTheDocument();
    expect(container.firstElementChild).toHaveClass('w-full');
    expect(container.firstElementChild).toHaveClass('min-w-0');
  });
});
