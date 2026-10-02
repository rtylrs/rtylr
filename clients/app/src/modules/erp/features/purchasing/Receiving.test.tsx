import React from 'react';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Receiving from './Receiving';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  usePurchaseOrders: vi.fn(),
  usePurchaseOrder: vi.fn(),
  updateStatus: vi.fn(),
  navigate: vi.fn(),
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
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

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('./usePurchaseOrders', () => ({
  usePurchaseOrders: mocks.usePurchaseOrders,
  usePurchaseOrder: mocks.usePurchaseOrder,
  useUpdatePurchaseOrderStatus: () => ({ mutate: mocks.updateStatus, isPending: false }),
}));

vi.mock('../suppliers/useSuppliers', () => ({
  useSuppliers: () => ({ data: { suppliers: [] } }),
}));

vi.mock('sonner', () => ({
  toast: { success: mocks.toastSuccess, error: mocks.toastError },
}));

const openOrder = {
  id: 'po1',
  supplierId: 's1',
  status: 'sent',
  items: [],
  expectedDelivery: '2026-01-01',
};

function showOpenOrder(): void {
  mocks.usePurchaseOrders.mockReturnValue({
    data: { purchaseOrders: [openOrder] },
    isLoading: false,
  });
}

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Receiving empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.usePurchaseOrders.mockReturnValue({
      data: { purchaseOrders: [] },
      isLoading: false,
    });
    mocks.usePurchaseOrder.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it('shows the illustration and empty copy, and navigates to purchasing on CTA click', () => {
    render(<Receiving />);

    const title = screen.getByText('No open orders to receive');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('All orders have been received')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Go to Purchase Orders' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/erp/purchasing' });
  });

  it('does not show the empty state when there are open orders', () => {
    showOpenOrder();

    render(<Receiving />);

    expect(screen.queryByText('No open orders to receive')).not.toBeInTheDocument();
  });
});

describe('Receiving line items', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    showOpenOrder();
    mocks.usePurchaseOrder.mockReturnValue({ data: null, isLoading: true, isError: false });
  });

  it('blocks submit while the order detail is still loading', () => {
    render(<Receiving />);
    fireEvent.click(screen.getByRole('button', { name: /PO-po1/ }));

    const markReceived = screen.getByRole('button', { name: 'Mark Received' });
    expect(markReceived).toBeDisabled();

    fireEvent.click(markReceived);
    expect(mocks.updateStatus).not.toHaveBeenCalled();
  });

  it('disables submit when the detail has no line items', () => {
    mocks.usePurchaseOrder.mockReturnValue({
      data: { id: 'po1', items: [] },
      isLoading: false,
      isError: false,
    });

    render(<Receiving />);
    fireEvent.click(screen.getByRole('button', { name: /PO-po1/ }));

    expect(screen.getByText('No line items available to receive')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Mark Received' })).toBeDisabled();
    expect(mocks.updateStatus).not.toHaveBeenCalled();
  });

  it('submits the loaded line items and closes the panel only after success', () => {
    mocks.usePurchaseOrder.mockReturnValue({
      data: {
        id: 'po1',
        items: [
          { id: 'i1', productId: 'p1', quantityOrdered: 5 },
          { id: 'i2', productId: 'p2', quantityOrdered: 4, quantityReceived: 4 },
        ],
      },
      isLoading: false,
      isError: false,
    });

    render(<Receiving />);
    fireEvent.click(screen.getByRole('button', { name: /PO-po1/ }));

    expect(screen.getByText('Product #p1')).toBeInTheDocument();
    expect(screen.getByText('Product #p2')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Mark Received' }));

    expect(mocks.updateStatus).toHaveBeenCalledTimes(1);
    const [request, handlers] = mocks.updateStatus.mock.calls[0];
    expect(request).toEqual({
      id: 'po1',
      body: {
        status: 'received',
        items: [
          { id: 'i1', quantityOrdered: 5, quantityReceived: 5 },
          { id: 'i2', quantityOrdered: 4, quantityReceived: 4 },
        ],
      },
    });
    // The panel must not claim success before the API confirms it.
    expect(mocks.toastSuccess).not.toHaveBeenCalled();
    expect(screen.getByRole('button', { name: 'Mark Received' })).toBeInTheDocument();

    act(() => handlers.onSuccess());
    expect(mocks.toastSuccess).toHaveBeenCalledWith('Receipt completed, all items received');
    expect(screen.queryByRole('button', { name: 'Mark Received' })).not.toBeInTheDocument();
  });

  it('keeps the receipt open and reports failure when the update fails', () => {
    mocks.usePurchaseOrder.mockReturnValue({
      data: { id: 'po1', items: [{ id: 'i1', productId: 'p1', quantityOrdered: 5 }] },
      isLoading: false,
      isError: false,
    });

    render(<Receiving />);
    fireEvent.click(screen.getByRole('button', { name: /PO-po1/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Mark Received' }));

    const [, handlers] = mocks.updateStatus.mock.calls[0];
    act(() => handlers.onError(new Error('network')));

    expect(mocks.toastSuccess).not.toHaveBeenCalled();
    expect(mocks.toastError).toHaveBeenCalledWith('Failed to update receipt status');
    expect(screen.getByRole('button', { name: 'Mark Received' })).toBeInTheDocument();
  });
});
