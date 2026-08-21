import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import StockTransfers from './StockTransfers';

const mocks = vi.hoisted(() => ({
  createTransfer: vi.fn(),
  role: 'admin' as string,
  refetch: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (select: (state: { user: { role: string } }) => unknown) =>
    select({ user: { role: mocks.role } }),
}));

vi.mock('./useStockTransfers', () => ({
  useStockTransfers: () => ({
    data: { stockTransfers: [], hasMore: false },
    isLoading: false,
    refetch: mocks.refetch,
  }),
  useCreateStockTransfer: () => ({
    mutateAsync: mocks.createTransfer,
    isPending: false,
  }),
}));

vi.mock('../warehouses/useWarehouses', () => ({
  useWarehouses: () => ({
    data: [
      { id: '1', name: 'Main warehouse' },
      { id: '2', name: 'Overflow warehouse' },
    ],
  }),
}));

vi.mock('../products/useProducts', () => ({
  useProducts: () => ({
    data: { products: [{ id: '11', name: 'Widget', sku: 'W-11' }] },
    isLoading: false,
  }),
}));

describe('StockTransfers', () => {
  beforeEach(() => {
    mocks.createTransfer.mockReset();
    mocks.createTransfer.mockResolvedValue({ transferId: '1' });
    mocks.refetch.mockReset();
    mocks.role = 'admin';
  });

  it('hides the New Transfer action for an operator, who lacks erp:inventory:transfer', () => {
    mocks.role = 'operator';
    render(<StockTransfers />);

    expect(screen.queryByRole('button', { name: 'New Transfer' })).not.toBeInTheDocument();
  });

  it('submits the selected product and positive quantity instead of an empty draft', async () => {
    render(<StockTransfers />);

    fireEvent.click(screen.getByRole('button', { name: 'New Transfer' }));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '1' } });
    fireEvent.change(selects[1], { target: { value: '2' } });
    fireEvent.change(selects[2], { target: { value: '11' } });
    fireEvent.change(screen.getByRole('spinbutton'), { target: { value: '3' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create Draft' }));

    await waitFor(() =>
      expect(mocks.createTransfer).toHaveBeenCalledWith({
        fromWarehouseId: '1',
        toWarehouseId: '2',
        notes: '',
        items: [{ productId: '11', quantity: 3 }],
      })
    );
  });
});

describe('StockTransfers empty state', () => {
  beforeEach(() => {
    mocks.refetch.mockReset();
    mocks.role = 'admin';
  });

  it('shows the create action and opens the create form for a role with erp:inventory:transfer', () => {
    mocks.role = 'admin';
    render(<StockTransfers />);

    const title = screen.getByText('stockTransfers.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('stockTransfers.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'stockTransfers.newTransfer' }));

    expect(screen.getByText('New Stock Transfer')).toBeInTheDocument();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('falls back to Refresh for a role without erp:inventory:transfer', () => {
    mocks.role = 'operator';
    render(<StockTransfers />);

    const title = screen.getByText('stockTransfers.empty.title');
    const emptyState = title.parentElement as HTMLElement;

    expect(
      within(emptyState).queryByRole('button', { name: 'stockTransfers.newTransfer' })
    ).not.toBeInTheDocument();

    fireEvent.click(within(emptyState).getByRole('button', { name: 'stockTransfers.refresh' }));

    expect(mocks.refetch).toHaveBeenCalledTimes(1);
    expect(screen.queryByText('New Stock Transfer')).not.toBeInTheDocument();
  });
});
