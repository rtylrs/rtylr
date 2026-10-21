import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  effectiveAccess: undefined as unknown,
  deleteMutate: vi.fn(),
  navigate: vi.fn(),
  product: undefined as Record<string, unknown> | undefined,
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: 'p1' }),
  useNavigate: () => mocks.navigate,
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

const FOUND_PRODUCT = {
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
};

vi.mock('./useProducts', () => ({
  useProductFull: () => ({
    data: mocks.product,
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
vi.mock('../../shared/components/Skeleton', () => ({ Skeleton: () => null }));
vi.mock('../../shared/components/QuickAddProductModal', () => ({ default: () => null }));
vi.mock('../../shared/components/ConfirmModal', () => ({ default: () => null }));
vi.mock('../categories/useCategories', () => ({ useCategories: () => ({ data: [] }) }));
vi.mock('../brands/useBrands', () => ({ useBrands: () => ({ data: [] }) }));

import ProductDetail from './ProductDetail';
import { initTestI18n } from '../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('ProductDetail permission gating', () => {
  beforeEach(() => {
    mocks.product = FOUND_PRODUCT;
    mocks.effectiveAccess = createEffectiveAccess('erp', { permissions: ['erp:products:view'] });
  });

  it('hides the delete button and inventory adjust control for an operator', () => {
    render(<ProductDetail />);

    expect(screen.queryByTestId('product-delete-button')).not.toBeInTheDocument();
    expect(screen.queryByTestId('quick-stock-adjust')).not.toBeInTheDocument();
  });

  it('hides the delete button and inventory adjust control for staff', () => {
    render(<ProductDetail />);

    expect(screen.queryByTestId('product-delete-button')).not.toBeInTheDocument();
    expect(screen.queryByTestId('quick-stock-adjust')).not.toBeInTheDocument();
  });

  it('shows the delete button and inventory adjust control for admin', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { isAdmin: true });
    render(<ProductDetail />);

    expect(screen.getByTestId('product-delete-button')).toBeInTheDocument();
    expect(screen.getByTestId('quick-stock-adjust')).toBeInTheDocument();
  });

  it('shows the delete button and inventory adjust control for owner', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { isOwner: true });
    render(<ProductDetail />);

    expect(screen.getByTestId('product-delete-button')).toBeInTheDocument();
    expect(screen.getByTestId('quick-stock-adjust')).toBeInTheDocument();
  });
});

describe('ProductDetail not-found empty state', () => {
  it('shows the illustration and empty copy, and navigates back to inventory on CTA click', () => {
    mocks.product = undefined;
    render(<ProductDetail />);

    const title = screen.getByText('Product not found');
    expect(title).toBeInTheDocument();
    expect(
      screen.getByText('The product you are looking for does not exist or has been deleted.')
    ).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Back to Inventory' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/erp/inventory' });
  });
});
