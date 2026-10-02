import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Products from './Products';
import { initTestI18n } from '../../shared/testing/i18n';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  useProducts: vi.fn(),
  effectiveAccess: undefined as unknown,
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

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

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Products empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.effectiveAccess = createEffectiveAccess('erp', { isOwner: true });
    mocks.useProducts.mockReturnValue({
      data: { products: [], total: 0, hasMore: false },
      isLoading: false,
    });
  });

  it('shows the illustration and empty copy, and opens the create modal on CTA click', () => {
    render(<Products />);

    const title = screen.getByText('No products found');
    expect(title).toBeInTheDocument();
    expect(
      screen.getByText('Try adjusting your filters or create your first product.')
    ).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Create Product' }));

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

    expect(screen.queryByText('No products found')).not.toBeInTheDocument();
  });
  it('does not offer creation to a member with product view access only', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { permissions: ['erp:products:view'] });
    render(<Products />);
    expect(screen.getByTestId('products-new-button')).toBeDisabled();
    expect(screen.queryByRole('button', { name: 'Create Product' })).toBeNull();
  });

  it('shows a retryable failure instead of an empty catalogue', () => {
    const refetch = vi.fn();
    mocks.useProducts.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch,
    });
    render(<Products />);
    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load products.');
    expect(screen.queryByText('No products found')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(refetch).toHaveBeenCalledOnce();
  });
});
