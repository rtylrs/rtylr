import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Variants from './Variants';

const mocks = vi.hoisted(() => ({
  useVariants: vi.fn(),
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

vi.mock('../products/useVariants', () => ({
  parseVariantAttributes: () => ({}),
  useVariants: mocks.useVariants,
  useCreateVariant: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

describe('Variants empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useVariants.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and opens the create form on CTA click', () => {
    render(<Variants />);

    const title = screen.getByText('variants.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('variants.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'variants.addVariant' }));

    expect(screen.getByPlaceholderText('Product ID')).toBeInTheDocument();
  });

  it('does not show the empty state when variants exist', () => {
    mocks.useVariants.mockReturnValue({
      data: [{ id: 'v1', name: 'Large', sku: 'V-1', stockQuantity: 5, salePrice: 10 }],
      isLoading: false,
    });

    render(<Variants />);

    expect(screen.queryByText('variants.empty.title')).not.toBeInTheDocument();
  });
});
