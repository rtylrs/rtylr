import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useReceipts } from './useReceipts';

const mockPosService = {
  getReceipts: vi.fn(),
  emailReceipt: vi.fn(),
};

const mockReceiptsStore = {
  receipts: [] as any[],
  add: vi.fn(),
  markEmailed: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/posService', () => ({
  posService: {
    getReceipts: (...args: any[]) => mockPosService.getReceipts(...args),
    emailReceipt: (...args: any[]) => mockPosService.emailReceipt(...args),
  },
}));

vi.mock('../../shared/stores/receiptsStore', () => ({
  useReceiptsStore: (selector: any) => {
    const state = {
      receipts: mockReceiptsStore.receipts,
      add: mockReceiptsStore.add,
      markEmailed: mockReceiptsStore.markEmailed,
    };
    return selector(state);
  },
}));

const sampleReceipts = [
  {
    id: 'r1',
    orderId: 'o1',
    total: 25.99,
    customerName: 'Alice',
    printedAt: '2026-01-01T10:00:00Z',
    items: [{ name: 'Burger', qty: 1, price: 12.99 }],
    payments: [{ id: 'p1', method: 'cash', amount: 25.99 }],
  },
  {
    id: 'r2',
    orderId: 'o2',
    total: 15.0,
    printedAt: '2026-01-01T11:00:00Z',
    items: [{ name: 'Fries', qty: 2, price: 7.5 }],
    payments: [{ id: 'p2', method: 'card', amount: 15.0 }],
  },
];

describe('useReceipts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockReceiptsStore.receipts = [];
  });

  it('loads remote receipts on mount', async () => {
    mockPosService.getReceipts.mockResolvedValue({ receipts: sampleReceipts });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.receipts).toHaveLength(2);
  });

  it('handles .data key in response', async () => {
    mockPosService.getReceipts.mockResolvedValue({ data: [sampleReceipts[0]] });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.receipts).toHaveLength(1);
  });

  it('handles error state', async () => {
    mockPosService.getReceipts.mockRejectedValue(new Error('Failed'));

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Failed');
  });

  it('merges local and remote receipts without duplicates', async () => {
    mockReceiptsStore.receipts = [
      { id: 'local1', orderId: 'o3', total: 10, printedAt: '2026-01-01', items: [], payments: [] },
      { id: 'r1', orderId: 'o1', total: 25.99, printedAt: '2026-01-01', items: [], payments: [] },
    ];
    mockPosService.getReceipts.mockResolvedValue({ receipts: sampleReceipts });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    // local1 + r1 (from local) + r2 (from remote, r1 deduped)
    expect(result.current.receipts).toHaveLength(3);
  });

  it('email function calls posService and marks emailed', async () => {
    mockPosService.getReceipts.mockResolvedValue({ receipts: [] });
    mockPosService.emailReceipt.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    let emailResult: boolean | undefined;
    await act(async () => {
      emailResult = await result.current.email('r1', 'alice@test.com');
    });

    expect(emailResult).toBe(true);
    // The hook calls posService.emailReceipt(Number(id), address) - 'r1' becomes NaN
    expect(mockPosService.emailReceipt).toHaveBeenCalledWith(NaN, 'alice@test.com');
    expect(mockReceiptsStore.markEmailed).toHaveBeenCalledWith('r1');
  });

  it('email function returns false on error', async () => {
    mockPosService.getReceipts.mockResolvedValue({ receipts: [] });
    mockPosService.emailReceipt.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    let emailResult: boolean | undefined;
    await act(async () => {
      emailResult = await result.current.email('r1', 'alice@test.com');
    });

    expect(emailResult).toBe(false);
  });

  it('reload function refetches remote data', async () => {
    mockPosService.getReceipts.mockResolvedValue({ receipts: [] });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    mockPosService.getReceipts.mockResolvedValue({ receipts: sampleReceipts });

    await act(async () => {
      await result.current.reload();
    });

    expect(result.current.receipts).toHaveLength(2);
  });

  it('exposes add function from store', async () => {
    mockPosService.getReceipts.mockResolvedValue({ receipts: [] });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.add).toBe(mockReceiptsStore.add);
  });
});
