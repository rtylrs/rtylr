import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Warehouses from './Warehouses';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useWarehouses: vi.fn(),
  createMutateAsync: vi.fn(),
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

vi.mock('./useWarehouses', () => ({
  useWarehouses: mocks.useWarehouses,
  useCreateWarehouse: () => ({ mutateAsync: mocks.createMutateAsync, isPending: false }),
  useDeleteWarehouse: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Warehouses empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useWarehouses.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and opens the create form on CTA click', () => {
    render(<Warehouses />);

    const title = screen.getByText('No warehouses');
    expect(title).toBeInTheDocument();
    expect(
      screen.getByText('Create your first warehouse to track stock locations')
    ).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'New Warehouse' }));

    expect(screen.getByPlaceholderText('Warehouse name')).toBeInTheDocument();
  });

  it('does not show the empty state when warehouses exist', () => {
    mocks.useWarehouses.mockReturnValue({
      data: [{ id: '1', name: 'Main', isDefault: true, isActive: true }],
      isLoading: false,
    });

    render(<Warehouses />);

    expect(screen.queryByText('No warehouses')).not.toBeInTheDocument();
  });
});
