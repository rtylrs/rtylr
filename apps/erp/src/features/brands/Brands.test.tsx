import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Brands from './Brands';

const mocks = vi.hoisted(() => ({
  useBrands: vi.fn(),
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

vi.mock('./useBrands', () => ({
  useBrands: mocks.useBrands,
  useCreateBrand: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteBrand: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

describe('Brands empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useBrands.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and opens the create form on CTA click', () => {
    render(<Brands />);

    const title = screen.getByText('brands.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('brands.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'brands.newBrand' }));

    expect(screen.getByPlaceholderText('Brand name')).toBeInTheDocument();
  });

  it('does not show the empty state when brands exist', () => {
    mocks.useBrands.mockReturnValue({
      data: [{ id: 'b1', name: 'Acme', productCount: 3 }],
      isLoading: false,
    });

    render(<Brands />);

    expect(screen.queryByText('brands.empty.title')).not.toBeInTheDocument();
  });
});
