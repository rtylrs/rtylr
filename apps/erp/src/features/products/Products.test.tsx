import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Products from './Products';

const mocks = vi.hoisted(() => ({
  useProducts: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/config', async (importOriginal) => ({
  ...(await importOriginal<Record<string, unknown>>()),
  cdnImageUrl: () => undefined,
}));

vi.mock('@voxire/utils', () => ({
  useDebounce: (value: string) => value,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (select: (state: { user: { role: string } }) => unknown) =>
    select({ user: { role: 'admin' } }),
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: React.PropsWithChildren<Record<string, unknown>>) => <a>{children}</a>,
}));

vi.mock('./useProducts', () => ({
  fetchAllProducts: vi.fn(),
  getProductStockStatus: () => 'In Stock',
  useProducts: mocks.useProducts,
  useDeleteProduct: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useCreateProduct: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('./importProducts', () => ({ importProducts: vi.fn() }));
vi.mock('../categories/useCategories', () => ({ useCategories: () => ({ data: [] }) }));
vi.mock('../../shared/components/ProductCard', () => ({ default: () => null }));
vi.mock('../../shared/components/StockBadge', () => ({ default: () => null }));
vi.mock('../../shared/components/FilterDrawer', () => ({
  default: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
}));
vi.mock('../../shared/components/BulkActionBar', () => ({ default: () => null }));
vi.mock('../../shared/components/ExportButton', () => ({ default: () => null }));
vi.mock('../../shared/components/ImportCSV', () => ({ default: () => null }));
vi.mock('../../shared/components/PaginationControls', () => ({ default: () => null }));
vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));
vi.mock('../../shared/components/QuickAddProductModal', () => ({
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="quick-add-modal-open" /> : null,
}));

describe('Products empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useProducts.mockReturnValue({
      data: { products: [], total: 0, hasMore: false },
      isLoading: false,
    });
  });

  it('shows the illustration and empty copy, and opens the create modal on CTA click', () => {
    render(<Products />);

    const title = screen.getByText('erp.products.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('erp.products.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'erp.products.create' }));

    expect(screen.getByTestId('quick-add-modal-open')).toBeInTheDocument();
  });

  it('does not show the empty state when products exist', () => {
    mocks.useProducts.mockReturnValue({
      data: {
        products: [{ id: 'p1', name: 'Widget', sku: 'W-1', stockQuantity: 5 }],
        total: 1,
        hasMore: false,
      },
      isLoading: false,
    });

    render(<Products />);

    expect(screen.queryByText('erp.products.empty.title')).not.toBeInTheDocument();
  });
});
