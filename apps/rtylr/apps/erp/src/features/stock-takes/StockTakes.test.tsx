import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import StockTakes from './StockTakes';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useStockTakes: vi.fn(),
  refetch: vi.fn(),
  role: 'admin' as string,
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

beforeAll(async () => {
  await initTestI18n('en');
});

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

    const title = screen.getByText('No stock takes');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('Start a count to find discrepancies')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'New Stock Take' }));

    expect(screen.getByText('Select warehouse…')).toBeInTheDocument();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('falls back to Refresh for a role without erp:stock_takes', () => {
    mocks.role = 'operator';
    render(<StockTakes />);

    const title = screen.getByText('No stock takes');
    const emptyState = title.parentElement as HTMLElement;

    expect(
      within(emptyState).queryByRole('button', { name: 'New Stock Take' })
    ).not.toBeInTheDocument();

    fireEvent.click(within(emptyState).getByRole('button', { name: 'Refresh' }));

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

    expect(screen.queryByText('No stock takes')).not.toBeInTheDocument();
  });
});
