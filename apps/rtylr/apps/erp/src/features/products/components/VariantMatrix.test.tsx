import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import VariantMatrix from './VariantMatrix';
import { initTestI18n } from '../../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useVariants: vi.fn(),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('../useVariants', () => ({
  parseVariantAttributes: () => ({}),
  useVariants: mocks.useVariants,
  useCreateVariant: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateVariant: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteVariant: () => ({ mutate: vi.fn(), isPending: false }),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('VariantMatrix empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useVariants.mockReturnValue({ data: [] });
  });

  it('shows the illustration and empty copy, and adds an option group on CTA click', () => {
    render(<VariantMatrix productId="p1" productSku="SKU-1" baseSalePrice={10} />);

    const title = screen.getByText('No variants yet. Add an option below to generate them.');
    expect(title).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Add option' }));

    expect(screen.getByPlaceholderText('Option name (e.g. Size, Color)')).toBeInTheDocument();
    expect(
      screen.queryByText('No variants yet. Add an option below to generate them.')
    ).not.toBeInTheDocument();
  });

  it('does not show the empty state when variants exist', () => {
    mocks.useVariants.mockReturnValue({
      data: [{ id: 'v1', name: 'Large', sku: 'V-1', stockQuantity: 5, salePrice: 10 }],
    });

    render(<VariantMatrix productId="p1" productSku="SKU-1" baseSalePrice={10} />);

    expect(
      screen.queryByText('No variants yet. Add an option below to generate them.')
    ).not.toBeInTheDocument();
  });
});
