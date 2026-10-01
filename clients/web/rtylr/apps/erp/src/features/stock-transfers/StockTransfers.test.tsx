import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import StockTransfers from './StockTransfers';
import { initTestI18n } from '../../shared/testing/i18n';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  createTransfer: vi.fn(),
  updateTransferStatus: vi.fn(),
  useStockTransfers: vi.fn(),
  effectiveAccess: undefined as unknown,
  refetch: vi.fn(),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: Object.assign(
    (
      select: (state: { user: { effectiveAccess: unknown }; effectiveAccess: unknown }) => unknown
    ) =>
      select({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    {
      getState: () => ({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    }
  ),
}));

vi.mock('./useStockTransfers', () => ({
  useStockTransfers: mocks.useStockTransfers,
  useCreateStockTransfer: () => ({
    mutateAsync: mocks.createTransfer,
    isPending: false,
  }),
  useUpdateStockTransferStatus: () => ({
    mutate: mocks.updateTransferStatus,
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

beforeAll(async () => {
  await initTestI18n('en');
});

describe('StockTransfers', () => {
  beforeEach(() => {
    mocks.createTransfer.mockReset();
    mocks.createTransfer.mockResolvedValue({ transferId: '1' });
    mocks.updateTransferStatus.mockReset();
    mocks.updateTransferStatus.mockImplementation((_request, callbacks) =>
      callbacks?.onSuccess?.()
    );
    mocks.useStockTransfers.mockReset();
    mocks.useStockTransfers.mockReturnValue({
      data: { stockTransfers: [], hasMore: false },
      isLoading: false,
      isError: false,
      refetch: mocks.refetch,
    });
    mocks.refetch.mockReset();
    mocks.effectiveAccess = createEffectiveAccess('erp', { isOwner: true });
  });

  it('hides the New Transfer action for an operator, who lacks erp:inventory:transfer', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { permissions: ['erp:inventory:view'] });
    render(<StockTransfers />);

    expect(screen.queryByRole('button', { name: 'New Transfer' })).not.toBeInTheDocument();
  });

  it('offers a receive action for a transfer and sends the received status', async () => {
    mocks.useStockTransfers.mockReturnValue({
      data: {
        stockTransfers: [
          {
            id: '9',
            fromWarehouseId: '1',
            toWarehouseId: '2',
            status: 'in_transit',
            items: [{ productId: '11', quantity: 2 }],
          },
        ],
        hasMore: false,
      },
      isLoading: false,
      isError: false,
      refetch: mocks.refetch,
    });
    render(<StockTransfers />);

    fireEvent.click(screen.getByRole('button', { name: 'Receive stock transfer #9' }));

    await waitFor(() =>
      expect(mocks.updateTransferStatus).toHaveBeenCalledWith(
        { id: '9', body: { status: 'received' } },
        expect.objectContaining({ onSuccess: expect.any(Function), onError: expect.any(Function) })
      )
    );
  });

  it('submits the selected product and positive quantity instead of an empty draft', async () => {
    render(<StockTransfers />);

    const [headerAction] = screen.getAllByRole('button', { name: 'New Transfer' });
    fireEvent.click(headerAction);
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
    mocks.useStockTransfers.mockReset();
    mocks.useStockTransfers.mockReturnValue({
      data: { stockTransfers: [], hasMore: false },
      isLoading: false,
      isError: false,
      refetch: mocks.refetch,
    });
    mocks.effectiveAccess = createEffectiveAccess('erp', { isOwner: true });
  });

  it('shows the create action and opens the create form with inventory transfer permission', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', {
      permissions: ['erp:inventory:transfer'],
    });
    render(<StockTransfers />);

    const title = screen.getByText('No transfers');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('Create your first stock transfer')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'New Transfer' }));

    expect(screen.getByText('New Stock Transfer')).toBeInTheDocument();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('falls back to Refresh without inventory transfer permission', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { permissions: ['erp:inventory:view'] });
    render(<StockTransfers />);

    const title = screen.getByText('No transfers');
    const emptyState = title.parentElement as HTMLElement;

    expect(
      within(emptyState).queryByRole('button', { name: 'New Transfer' })
    ).not.toBeInTheDocument();

    fireEvent.click(within(emptyState).getByRole('button', { name: 'Refresh' }));

    expect(mocks.refetch).toHaveBeenCalledTimes(1);
    expect(screen.queryByText('New Stock Transfer')).not.toBeInTheDocument();
  });
});
