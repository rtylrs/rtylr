import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import QuickAddProductModal from './QuickAddProductModal';
import { initTestI18n } from '../testing/i18n';
import type { FullProduct, Product } from '../../features/products/useProducts';

const mocks = vi.hoisted(() => ({
  useProductFull: vi.fn(),
  emptyAttributes: [] as unknown[],
}));

vi.mock('../../features/products/useProducts', () => ({
  useProductFull: mocks.useProductFull,
  useProductAttributes: () => ({ data: mocks.emptyAttributes }),
  useCreateProduct: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateProduct: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../features/categories/useCategories', () => ({
  useCategories: () => ({ data: [] }),
}));
vi.mock('../../features/suppliers/useSuppliers', () => ({
  useSuppliers: () => ({ data: { suppliers: [] } }),
}));
vi.mock('../../features/brands/useBrands', () => ({ useBrands: () => ({ data: [] }) }));
vi.mock('@voxire/api', () => ({ uploadAsset: vi.fn() }));
vi.mock('@voxire/config', () => ({ cdnImageUrl: (url: string) => url }));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

const listProduct: Product = {
  id: '1',
  name: 'List name',
  sku: 'SKU-1',
  salePrice: 5,
  stockQuantity: 3,
};

const fullProduct: FullProduct = {
  id: '1',
  name: 'Full name',
  sku: 'SKU-1',
  salePrice: 5,
  stockQuantity: 3,
  compareAtPrice: 7,
  taxRate: 10,
};

beforeAll(async () => {
  await initTestI18n('en');
});

describe('QuickAddProductModal edit seeding', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useProductFull.mockReset();
  });

  it('keeps in-progress input when the full product payload arrives late or refetches', () => {
    mocks.useProductFull.mockReturnValue({ data: undefined });
    const { rerender } = render(
      <QuickAddProductModal open onClose={vi.fn()} editing={listProduct} />
    );

    fireEvent.change(screen.getByTestId('qap-name'), { target: { value: 'Unsaved name' } });
    fireEvent.change(screen.getByLabelText('Sale price'), { target: { value: '9' } });

    mocks.useProductFull.mockReturnValue({ data: fullProduct });
    rerender(<QuickAddProductModal open onClose={vi.fn()} editing={listProduct} />);

    expect(screen.getByTestId('qap-name')).toHaveValue('Unsaved name');
    expect(screen.getByLabelText('Sale price')).toHaveValue(9);
  });

  it('seeds the PIM fields from the first full product payload while untouched', () => {
    mocks.useProductFull.mockReturnValue({ data: undefined });
    const { rerender } = render(
      <QuickAddProductModal open onClose={vi.fn()} editing={listProduct} />
    );
    expect(screen.getByTestId('qap-name')).toHaveValue('List name');

    mocks.useProductFull.mockReturnValue({ data: fullProduct });
    rerender(<QuickAddProductModal open onClose={vi.fn()} editing={listProduct} />);

    expect(screen.getByTestId('qap-name')).toHaveValue('Full name');
    expect(screen.getByLabelText('Compare-at price')).toHaveValue(7);
  });

  it('re-seeds when a different product is loaded for editing', () => {
    mocks.useProductFull.mockReturnValue({ data: undefined });
    const { rerender } = render(
      <QuickAddProductModal open onClose={vi.fn()} editing={listProduct} />
    );
    fireEvent.change(screen.getByTestId('qap-name'), { target: { value: 'Typed' } });

    rerender(
      <QuickAddProductModal
        open
        onClose={vi.fn()}
        editing={{ ...listProduct, id: '2', name: 'Second name' }}
      />
    );

    expect(screen.getByTestId('qap-name')).toHaveValue('Second name');
  });
});
