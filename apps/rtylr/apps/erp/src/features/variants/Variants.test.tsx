import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Variants from './Variants';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useVariants: vi.fn(),
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

vi.mock('../products/useVariants', () => ({
  parseVariantAttributes: () => ({}),
  useVariants: mocks.useVariants,
  useCreateVariant: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Variants empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useVariants.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and opens the create form on CTA click', () => {
    render(<Variants />);

    const title = screen.getByText('No variants');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('Add variants to differentiate your products')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Add Variant' }));

    expect(screen.getByPlaceholderText('Product ID')).toBeInTheDocument();
  });

  it('does not show the empty state when variants exist', () => {
    mocks.useVariants.mockReturnValue({
      data: [{ id: 'v1', name: 'Large', sku: 'V-1', stockQuantity: 5, salePrice: 10 }],
      isLoading: false,
    });

    render(<Variants />);

    expect(screen.queryByText('No variants')).not.toBeInTheDocument();
  });
});
