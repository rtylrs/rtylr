import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Warehouses from './Warehouses';

const mocks = vi.hoisted(() => ({
  useWarehouses: vi.fn(),
  createMutateAsync: vi.fn(),
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

vi.mock('./useWarehouses', () => ({
  useWarehouses: mocks.useWarehouses,
  useCreateWarehouse: () => ({ mutateAsync: mocks.createMutateAsync, isPending: false }),
  useDeleteWarehouse: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

describe('Warehouses empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useWarehouses.mockReturnValue({ data: [], isLoading: false });
  });

  it('shows the illustration and empty copy, and opens the create form on CTA click', () => {
    render(<Warehouses />);

    const title = screen.getByText('warehouses.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('warehouses.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'warehouses.new' }));

    expect(screen.getByPlaceholderText('warehouses.fields.name')).toBeInTheDocument();
  });

  it('does not show the empty state when warehouses exist', () => {
    mocks.useWarehouses.mockReturnValue({
      data: [{ id: '1', name: 'Main', isDefault: true, isActive: true }],
      isLoading: false,
    });

    render(<Warehouses />);

    expect(screen.queryByText('warehouses.empty.title')).not.toBeInTheDocument();
  });
});
