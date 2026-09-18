import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Barcodes from './Barcodes';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useProducts: vi.fn(),
  navigate: vi.fn(),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('../products/useProducts', () => ({
  useProducts: mocks.useProducts,
}));

vi.mock('../../shared/components/BarcodeSvg', () => ({ default: () => null }));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Barcodes preview empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('selects the first product on CTA click when products are available', () => {
    mocks.useProducts.mockReturnValue({
      data: { products: [{ id: 'p1', name: 'Widget', sku: 'W-1', salePrice: 10 }] },
      isLoading: false,
    });

    render(<Barcodes />);

    const title = screen.getByText('Select products to preview labels');
    expect(title).toBeInTheDocument();
    expect(
      screen.getByText(
        'Choose one or more products from the list to see their barcode labels here.'
      )
    ).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Select First Product' }));

    expect(screen.queryByText('Select products to preview labels')).not.toBeInTheDocument();
  });

  it('navigates to products on CTA click when there are no products at all', () => {
    mocks.useProducts.mockReturnValue({
      data: { products: [] },
      isLoading: false,
    });

    render(<Barcodes />);

    const title = screen.getByText('Select products to preview labels');
    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Go to Products' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/products' });
  });
});
