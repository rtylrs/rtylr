import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Inventory from './Inventory';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useProducts: vi.fn(),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/config', () => ({
  cdnImageUrl: () => undefined,
}));

vi.mock('@voxire/utils', () => ({
  useDebounce: (value: string) => value,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('../products/useProducts', () => ({
  getProductStockStatus: () => 'In Stock',
  useProducts: mocks.useProducts,
}));

vi.mock('../../shared/components/QuickAddProductModal', () => ({
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="quick-add-modal-open" /> : null,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Inventory empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useProducts.mockReturnValue({
      data: { products: [], total: 0, hasMore: false },
      isLoading: false,
    });
  });

  it('shows the illustration and clears the search on CTA click', () => {
    render(<Inventory />);

    const title = screen.getByText('No products found.');
    expect(title).toBeInTheDocument();

    const searchInput = screen.getByPlaceholderText('Search products…');
    fireEvent.change(searchInput, { target: { value: 'widget' } });
    expect(searchInput).toHaveValue('widget');

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Clear Search' }));

    expect(searchInput).toHaveValue('');
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

    render(<Inventory />);

    expect(screen.queryByText('No products found.')).not.toBeInTheDocument();
  });
});
