import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import VariantMatrix from './VariantMatrix';

const mocks = vi.hoisted(() => ({
  useVariants: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('../useVariants', () => ({
  parseVariantAttributes: () => ({}),
  useVariants: mocks.useVariants,
  useCreateVariant: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateVariant: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteVariant: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe('VariantMatrix empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useVariants.mockReturnValue({ data: [] });
  });

  it('shows the illustration and empty copy, and adds an option group on CTA click', () => {
    render(<VariantMatrix productId="p1" productSku="SKU-1" baseSalePrice={10} />);

    const title = screen.getByText('erp.products.variants.empty');
    expect(title).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(
      within(emptyState).getByRole('button', { name: 'erp.products.variants.addOption' })
    );

    expect(screen.getByPlaceholderText('erp.products.variants.optionName')).toBeInTheDocument();
    expect(screen.queryByText('erp.products.variants.empty')).not.toBeInTheDocument();
  });

  it('does not show the empty state when variants exist', () => {
    mocks.useVariants.mockReturnValue({
      data: [{ id: 'v1', name: 'Large', sku: 'V-1', stockQuantity: 5, salePrice: 10 }],
    });

    render(<VariantMatrix productId="p1" productSku="SKU-1" baseSalePrice={10} />);

    expect(screen.queryByText('erp.products.variants.empty')).not.toBeInTheDocument();
  });
});
