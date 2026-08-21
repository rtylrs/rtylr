import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Categories from './Categories';

const mocks = vi.hoisted(() => ({
  useCategories: vi.fn(),
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

vi.mock('./useCategories', () => ({
  useCategories: mocks.useCategories,
  useCreateCategory: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteCategory: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateCategory: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe('Categories empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useCategories.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and focuses the new-category input on CTA click', () => {
    render(<Categories />);

    const title = screen.getByText('categories.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('categories.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'categories.addCategory' }));

    expect(screen.getByPlaceholderText('New category name…')).toHaveFocus();
  });

  it('does not show the empty state when categories exist', () => {
    mocks.useCategories.mockReturnValue({
      data: [{ id: 'c1', name: 'Beverages' }],
      isLoading: false,
    });

    render(<Categories />);

    expect(screen.queryByText('categories.empty.title')).not.toBeInTheDocument();
  });
});
