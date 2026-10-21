import React from 'react';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Categories from './Categories';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useCategories: vi.fn(),
  createCategory: vi.fn(),
  createState: { isPending: false },
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

vi.mock('./useCategories', () => ({
  useCategories: mocks.useCategories,
  useCreateCategory: () => ({
    mutateAsync: mocks.createCategory,
    // The real mutation object is stable and reports live pending state.
    get isPending() {
      return mocks.createState.isPending;
    },
  }),
  useDeleteCategory: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateCategory: () => ({ mutate: vi.fn(), isPending: false }),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Categories empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.createCategory.mockReset();
    mocks.createState.isPending = false;
    mocks.useCategories.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and focuses the new-category input on CTA click', () => {
    render(<Categories />);

    const title = screen.getByText('No categories yet');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('Create your first category to organize products')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Add Category' }));

    expect(screen.getByPlaceholderText('New category name…')).toHaveFocus();
  });

  it('does not show the empty state when categories exist', () => {
    mocks.useCategories.mockReturnValue({
      data: [{ id: 'c1', name: 'Beverages' }],
      isLoading: false,
    });

    render(<Categories />);

    expect(screen.queryByText('No categories yet')).not.toBeInTheDocument();
  });
});

describe('Categories create submission', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.createCategory.mockReset();
    mocks.createState.isPending = false;
    mocks.useCategories.mockReturnValue({ data: [], isLoading: false });
  });

  it('creates only one category when Enter is pressed again while the first create is in flight', async () => {
    let resolveCreate!: (value: unknown) => void;
    mocks.createCategory.mockImplementation(() => {
      mocks.createState.isPending = true;
      return new Promise((resolve) => {
        resolveCreate = resolve;
      });
    });

    render(<Categories />);
    const input = screen.getByPlaceholderText('New category name…');
    fireEvent.change(input, { target: { value: 'Beverages' } });

    fireEvent.keyDown(input, { key: 'Enter' });
    fireEvent.keyDown(input, { key: 'Enter' });

    expect(mocks.createCategory).toHaveBeenCalledTimes(1);

    resolveCreate({ categoryId: 'c2' });
    await waitFor(() => expect(input).toHaveValue(''));
  });
});
