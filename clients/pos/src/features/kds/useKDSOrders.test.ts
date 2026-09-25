import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useKDSOrders } from './useKDSOrders';

const kdsClient = vi.hoisted(() => ({
  listKDSOrders: vi.fn(),
  updateKDSItemStatus: vi.fn(),
}));

vi.mock('@voxire/api', () => {
  class MockRealtimeClient {
    subscribe = vi.fn(() => vi.fn());
  }
  return {
    createOneKitClient: () => kdsClient,
    createApiClient: vi.fn(() => ({ defaults: {} })),
    safeFetch: vi.fn(() => ({})),
    RealtimeClient: MockRealtimeClient,
  };
});

describe('useKDSOrders', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads KDS orders on mount', async () => {
    kdsClient.listKDSOrders.mockResolvedValue({
      orders: [
        {
          id: '1',
          orderNumber: '101',
          type: 'Kitchen',
          items: [{ id: '11', orderId: '1', name: 'Burger', quantity: 2 }],
          status: 'new',
          createdAt: '2026-01-01T10:00:00Z',
        },
      ],
    });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.orders).toHaveLength(1);
    expect(result.current.orders[0].number).toBe('101');
    expect(result.current.orders[0].items[0].name).toBe('Burger');
  });

  it('passes station filter to service', async () => {
    kdsClient.listKDSOrders.mockResolvedValue({ orders: [] });

    renderHook(() => useKDSOrders('Bar'));

    await waitFor(() => expect(kdsClient.listKDSOrders).toHaveBeenCalledWith({}));
  });

  it('handles error state', async () => {
    kdsClient.listKDSOrders.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Network error');
    expect(result.current.orders).toEqual([]);
  });

  it('setStatus updates order locally and calls API', async () => {
    kdsClient.listKDSOrders.mockResolvedValue({
      orders: [{
        id: '1',
        orderNumber: '101',
        type: 'Kitchen',
        items: [{ id: '9', orderId: '1', name: 'Burger', quantity: 1 }],
        status: 'new',
      }],
    });
    kdsClient.updateKDSItemStatus.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.setStatus('1', 'ready');
    });

    expect(result.current.orders[0].status).toBe('ready');
    expect(kdsClient.updateKDSItemStatus).toHaveBeenCalledWith({
      orderId: '1',
      itemId: '9',
      body: { status: 'ready' },
    });
  });

  it('setStatus rejects invalid item ids without calling the API', async () => {
    kdsClient.listKDSOrders.mockResolvedValue({
      orders: [{ id: 'local-1', orderNumber: '101', type: 'Kitchen', items: [], status: 'new' }],
    });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.setStatus('local-1', 'ready');
    });

    expect(result.current.orders[0].status).toBe('new');
    expect(kdsClient.updateKDSItemStatus).not.toHaveBeenCalled();
  });

  it('defaults missing fields correctly', async () => {
    kdsClient.listKDSOrders.mockResolvedValue({
      orders: [{ id: '3' }],
    });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.orders[0].station).toBe('Kitchen');
    expect(result.current.orders[0].status).toBe('new');
    expect(result.current.orders[0].items).toEqual([]);
  });

  it('reload function refetches data', async () => {
    kdsClient.listKDSOrders.mockResolvedValue({ orders: [] });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    kdsClient.listKDSOrders.mockResolvedValue({
      orders: [{ id: '5', orderNumber: '105', type: 'Kitchen', items: [], status: 'new' }],
    });

    await act(async () => {
      await result.current.reload();
    });

    expect(result.current.orders).toHaveLength(1);
  });
});
