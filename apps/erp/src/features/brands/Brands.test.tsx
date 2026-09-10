import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Brands from './Brands';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useBrands: vi.fn(),
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

vi.mock('./useBrands', () => ({
  useBrands: mocks.useBrands,
  useCreateBrand: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteBrand: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Brands empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useBrands.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and opens the create form on CTA click', () => {
    render(<Brands />);

    const title = screen.getByText('No brands yet');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('Add your first brand to group products.')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'New Brand' }));

    expect(screen.getByPlaceholderText('Brand name')).toBeInTheDocument();
  });

  it('does not show the empty state when brands exist', () => {
    mocks.useBrands.mockReturnValue({
      data: [{ id: 'b1', name: 'Acme', productCount: 3 }],
      isLoading: false,
    });

    render(<Brands />);

    expect(screen.queryByText('No brands yet')).not.toBeInTheDocument();
  });
});
