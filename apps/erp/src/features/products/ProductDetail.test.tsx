import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  role: 'operator' as string | null,
  deleteMutate: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: 'p1' }),
  useNavigate: () => vi.fn(),
  Link: ({ children, to }: React.PropsWithChildren<{ to: string }>) => <a href={to}>{children}</a>,
}));

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

vi.mock('@voxire/auth', () => ({
  useUserStore: (select: (state: { user: { role: string | null } }) => unknown) =>
    select({ user: { role: mocks.role } }),
}));

vi.mock('./useProducts', () => ({
  useProductFull: () => ({
    data: {
      id: 'p1',
      name: 'Espresso Beans',
      sku: 'ESP-01',
      stockQuantity: 12,
      lowStockThreshold: 5,
      categoryId: '1',
      brandId: '0',
      description: 'Roasted beans',
      costPrice: 5,
      salePrice: 10,
      leadTimeDays: 7,
      reorderQuantity: 10,
      unit: 'kg',
      barcode: '123456',
    },
    isLoading: false,
  }),
  useDeleteProduct: () => ({ mutate: mocks.deleteMutate, isPending: false }),
}));

vi.mock('./components/VariantMatrix', () => ({ default: () => null }));
vi.mock('../../shared/components/Breadcrumbs', () => ({ default: () => null }));
vi.mock('../../shared/components/StockBadge', () => ({ default: () => null }));
vi.mock('../../shared/components/BarcodeSvg', () => ({ default: () => null }));
vi.mock('../../shared/components/QuickStockAdjust', () => ({
  default: () => <div data-testid="quick-stock-adjust" />,
}));
vi.mock('@voxire/ui', () => ({ EmptyState: () => null }));
vi.mock('../../shared/components/Skeleton', () => ({ Skeleton: () => null }));
vi.mock('../../shared/components/QuickAddProductModal', () => ({ default: () => null }));
vi.mock('../../shared/components/ConfirmModal', () => ({ default: () => null }));
vi.mock('../categories/useCategories', () => ({ useCategories: () => ({ data: [] }) }));
vi.mock('../brands/useBrands', () => ({ useBrands: () => ({ data: [] }) }));

import ProductDetail from './ProductDetail';

describe('ProductDetail permission gating', () => {
  it('hides the delete button and inventory adjust control for an operator', () => {
    mocks.role = 'operator';
    render(<ProductDetail />);

    expect(screen.queryByTestId('product-delete-button')).not.toBeInTheDocument();
    expect(screen.queryByTestId('quick-stock-adjust')).not.toBeInTheDocument();
  });

  it('hides the delete button and inventory adjust control for staff', () => {
    mocks.role = 'staff';
    render(<ProductDetail />);

    expect(screen.queryByTestId('product-delete-button')).not.toBeInTheDocument();
    expect(screen.queryByTestId('quick-stock-adjust')).not.toBeInTheDocument();
  });

  it('shows the delete button and inventory adjust control for admin', () => {
    mocks.role = 'admin';
    render(<ProductDetail />);

    expect(screen.getByTestId('product-delete-button')).toBeInTheDocument();
    expect(screen.getByTestId('quick-stock-adjust')).toBeInTheDocument();
  });

  it('shows the delete button and inventory adjust control for owner', () => {
    mocks.role = 'owner';
    render(<ProductDetail />);

    expect(screen.getByTestId('product-delete-button')).toBeInTheDocument();
    expect(screen.getByTestId('quick-stock-adjust')).toBeInTheDocument();
  });
});
