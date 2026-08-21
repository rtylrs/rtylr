import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import StockTransfers from './StockTransfers';

const mocks = vi.hoisted(() => ({
  useStockTransfers: vi.fn(),
  createTransfer: vi.fn(),
  updateStatus: vi.fn(),
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
  useStockTransfers: mocks.useStockTransfers,
  useCreateStockTransfer: () => ({
    mutateAsync: mocks.createTransfer,
    isPending: false,
  }),
  useUpdateStockTransferStatus: () => ({
    mutateAsync: mocks.updateStatus,
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
    mocks.updateStatus.mockReset();
    mocks.updateStatus.mockResolvedValue({});
    mocks.refetch.mockReset();
    mocks.role = 'admin';
    mocks.useStockTransfers.mockReturnValue({
      data: { stockTransfers: [], hasMore: false },
      isLoading: false,
      refetch: mocks.refetch,
    });
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
    mocks.useStockTransfers.mockReturnValue({
      data: { stockTransfers: [], hasMore: false },
      isLoading: false,
      refetch: mocks.refetch,
    });
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

describe('StockTransfers execution actions', () => {
  beforeEach(() => {
    mocks.updateStatus.mockReset();
    mocks.updateStatus.mockResolvedValue({});
    mocks.role = 'admin';
  });

  it('sends a drafted transfer only after confirming in the modal', async () => {
    mocks.useStockTransfers.mockReturnValue({
      data: {
        stockTransfers: [{ id: 't1', status: 'draft', items: [] }],
        hasMore: false,
      },
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<StockTransfers />);

    fireEvent.click(screen.getByRole('button', { name: 'Send transfer #t1' }));
    expect(mocks.updateStatus).not.toHaveBeenCalled();

    fireEvent.click(screen.getByTestId('confirm-modal-confirm'));

    await waitFor(() =>
      expect(mocks.updateStatus).toHaveBeenCalledWith({ id: 't1', body: { status: 'in_transit' } })
    );
  });

  it('marks an in-transit transfer received and hides row actions from operators', async () => {
    mocks.useStockTransfers.mockReturnValue({
      data: {
        stockTransfers: [{ id: 't2', status: 'in_transit', items: [] }],
        hasMore: false,
      },
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<StockTransfers />);
    expect(screen.getByRole('button', { name: 'Mark transfer #t2 received' })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Mark transfer #t2 received' }));
    fireEvent.click(screen.getByTestId('confirm-modal-confirm'));

    await waitFor(() =>
      expect(mocks.updateStatus).toHaveBeenCalledWith({ id: 't2', body: { status: 'received' } })
    );

    cleanup();
    mocks.role = 'operator';
    render(<StockTransfers />);
    expect(
      screen.queryByRole('button', { name: 'Mark transfer #t2 received' })
    ).not.toBeInTheDocument();
  });
});
