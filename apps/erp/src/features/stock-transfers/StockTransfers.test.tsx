import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import StockTransfers from './StockTransfers';

const mocks = vi.hoisted(() => ({
  createTransfer: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('./useStockTransfers', () => ({
  useStockTransfers: () => ({
    data: { stockTransfers: [], hasMore: false },
    isLoading: false,
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
