import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import StockTakes from './StockTakes';

const mocks = vi.hoisted(() => ({
  useStockTakes: vi.fn(),
  refetch: vi.fn(),
  role: 'admin' as string,
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

vi.mock('@voxire/auth', () => ({
  useUserStore: (select: (state: { user: { role: string } }) => unknown) =>
    select({ user: { role: mocks.role } }),
}));

vi.mock('./useStockTakes', () => ({
  useStockTakes: mocks.useStockTakes,
  useCreateStockTake: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateStockTake: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../warehouses/useWarehouses', () => ({
  useWarehouses: () => ({ data: [{ id: '1', name: 'Main warehouse' }] }),
}));

vi.mock('../inventory/useInventory', () => ({
  useInventoryReport: () => ({ data: [], isLoading: false, isError: false }),
}));

describe('StockTakes empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useStockTakes.mockReturnValue({
      data: { stockTakes: [], hasMore: false },
      isLoading: false,
      refetch: mocks.refetch,
    });
  });

  it('shows the create action and opens the create form for a role with erp:stock_takes', () => {
    mocks.role = 'admin';
    render(<StockTakes />);

    const title = screen.getByText('stockTakes.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('stockTakes.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'stockTakes.newStockTake' }));

    expect(screen.getByText('Select warehouse…')).toBeInTheDocument();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('falls back to Refresh for a role without erp:stock_takes', () => {
    mocks.role = 'operator';
    render(<StockTakes />);

    const title = screen.getByText('stockTakes.empty.title');
    const emptyState = title.parentElement as HTMLElement;

    expect(
      within(emptyState).queryByRole('button', { name: 'stockTakes.newStockTake' })
    ).not.toBeInTheDocument();

    fireEvent.click(within(emptyState).getByRole('button', { name: 'stockTakes.refresh' }));

    expect(mocks.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when stock takes exist', () => {
    mocks.useStockTakes.mockReturnValue({
      data: {
        stockTakes: [{ id: 'st1', warehouseId: '1', status: 'in_progress', items: [] }],
        hasMore: false,
      },
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<StockTakes />);

    expect(screen.queryByText('stockTakes.empty.title')).not.toBeInTheDocument();
  });
});
