import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Inventory from './Inventory';

const mocks = vi.hoisted(() => ({
  useProducts: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

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

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: React.PropsWithChildren<Record<string, unknown>>) => <a>{children}</a>,
}));

vi.mock('../products/useProducts', () => ({
  getProductStockStatus: () => 'In Stock',
  useProducts: mocks.useProducts,
}));

vi.mock('../../shared/components/QuickAddProductModal', () => ({
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="quick-add-modal-open" /> : null,
}));

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

    const title = screen.getByText('erp.inventory.empty');
    expect(title).toBeInTheDocument();

    const searchInput = screen.getByPlaceholderText('Search products…');
    fireEvent.change(searchInput, { target: { value: 'widget' } });
    expect(searchInput).toHaveValue('widget');

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'erp.inventory.clearSearch' }));

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

    expect(screen.queryByText('erp.inventory.empty')).not.toBeInTheDocument();
  });
});
