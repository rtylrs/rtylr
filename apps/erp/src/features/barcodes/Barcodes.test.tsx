import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Barcodes from './Barcodes';

const mocks = vi.hoisted(() => ({
  useProducts: vi.fn(),
  navigate: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

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

    const title = screen.getByText('barcodes.previewEmpty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('barcodes.previewEmpty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'barcodes.selectFirst' }));

    expect(screen.queryByText('barcodes.previewEmpty.title')).not.toBeInTheDocument();
  });

  it('navigates to products on CTA click when there are no products at all', () => {
    mocks.useProducts.mockReturnValue({
      data: { products: [] },
      isLoading: false,
    });

    render(<Barcodes />);

    const title = screen.getByText('barcodes.previewEmpty.title');
    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'barcodes.goToProducts' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/products' });
  });
});
