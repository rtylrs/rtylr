import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useKDSOrders } from './useKDSOrders';

const mockPosService = {
  getKDSOrders: vi.fn(),
  updateKDSItemStatus: vi.fn(),
};

vi.mock('@voxire/api', () => {
  class MockRealtimeClient {
    subscribe = vi.fn(() => vi.fn());
  }
  return {
    unwrap: vi.fn((v) => v),
    createApiClient: vi.fn(() => ({})),
    safeFetch: vi.fn(() => ({})),
    RealtimeClient: MockRealtimeClient,
  };
});

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    getKDSOrders: (...args: unknown[]) => mockPosService.getKDSOrders(...args),
    updateKDSItemStatus: (...args: unknown[]) => mockPosService.updateKDSItemStatus(...args),
  },
}));

describe('useKDSOrders', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads KDS orders on mount', async () => {
    mockPosService.getKDSOrders.mockResolvedValue({
      orders: [
        {
          id: '1',
          number: '101',
          station: 'Kitchen',
          items: [{ name: 'Burger', qty: 2 }],
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
    mockPosService.getKDSOrders.mockResolvedValue({ orders: [] });

    renderHook(() => useKDSOrders('Bar'));

    await waitFor(() => expect(mockPosService.getKDSOrders).toHaveBeenCalledWith('Bar'));
  });

  it('handles error state', async () => {
    mockPosService.getKDSOrders.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Network error');
    expect(result.current.orders).toEqual([]);
  });

  it('handles .data key in response', async () => {
    mockPosService.getKDSOrders.mockResolvedValue({
      data: [{ id: '2', number: '102', station: 'Grill', items: [], status: 'cooking' }],
    });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.orders).toHaveLength(1);
    expect(result.current.orders[0].status).toBe('cooking');
  });

  it('setStatus updates order locally and calls API', async () => {
    mockPosService.getKDSOrders.mockResolvedValue({
      orders: [{ id: '1', number: '101', station: 'Kitchen', items: [], status: 'new' }],
    });
    mockPosService.updateKDSItemStatus.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.setStatus('1', 'ready');
    });

    expect(result.current.orders[0].status).toBe('ready');
    expect(mockPosService.updateKDSItemStatus).toHaveBeenCalledWith(1, 'ready');
  });

  it('setStatus rejects invalid item ids without calling the API', async () => {
    mockPosService.getKDSOrders.mockResolvedValue({
      orders: [{ id: 'local-1', number: '101', station: 'Kitchen', items: [], status: 'new' }],
    });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    await expect(async () => {
      await act(async () => {
        await result.current.setStatus('local-1', 'ready');
      });
    }).rejects.toThrow('Select a valid KDS item before updating status.');

    expect(result.current.orders[0].status).toBe('new');
    expect(mockPosService.updateKDSItemStatus).not.toHaveBeenCalled();
  });

  it('defaults missing fields correctly', async () => {
    mockPosService.getKDSOrders.mockResolvedValue({
      orders: [{ id: '3' }],
    });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.orders[0].station).toBe('Kitchen');
    expect(result.current.orders[0].status).toBe('new');
    expect(result.current.orders[0].items).toEqual([]);
  });

  it('reload function refetches data', async () => {
    mockPosService.getKDSOrders.mockResolvedValue({ orders: [] });

    const { result } = renderHook(() => useKDSOrders());

    await waitFor(() => expect(result.current.loading).toBe(false));

    mockPosService.getKDSOrders.mockResolvedValue({
      orders: [{ id: '5', number: '105', station: 'Kitchen', items: [], status: 'new' }],
    });

    await act(async () => {
      await result.current.reload();
    });

    expect(result.current.orders).toHaveLength(1);
  });
});
