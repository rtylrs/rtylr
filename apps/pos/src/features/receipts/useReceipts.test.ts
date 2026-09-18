import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useReceipts } from './useReceipts';

const receiptClient = vi.hoisted(() => ({
  listReceipts: vi.fn(),
  emailReceipt: vi.fn(),
}));

const mockReceiptsStore = {
  receipts: [] as any[],
  add: vi.fn(),
  markEmailed: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => receiptClient,
}));

vi.mock('../../shared/stores/receiptsStore', () => ({
  useReceiptsStore: (selector: (s: typeof mockReceiptsStore) => unknown) => {
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
    receiptClient.listReceipts.mockResolvedValue({ receipts: sampleReceipts });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.receipts).toHaveLength(2);
  });

  it('handles error state', async () => {
    receiptClient.listReceipts.mockRejectedValue(new Error('Failed'));

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Failed');
  });

  it('merges local and remote receipts without duplicates', async () => {
    mockReceiptsStore.receipts = [
      { id: 'local1', orderId: 'o3', total: 10, printedAt: '2026-01-01', items: [], payments: [] },
      { id: 'r1', orderId: 'o1', total: 25.99, printedAt: '2026-01-01', items: [], payments: [] },
    ];
    receiptClient.listReceipts.mockResolvedValue({ receipts: sampleReceipts });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    // local1 + r1 (from local) + r2 (from remote, r1 deduped)
    expect(result.current.receipts).toHaveLength(3);
  });

  it('email function calls generated receipt endpoint and marks emailed for remote receipt ids', async () => {
    receiptClient.listReceipts.mockResolvedValue({ receipts: [] });
    receiptClient.emailReceipt.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    let emailResult: boolean | undefined;
    await act(async () => {
      emailResult = await result.current.email('42', 'alice@test.com');
    });

    expect(emailResult).toBe(true);
    expect(receiptClient.emailReceipt).toHaveBeenCalledWith({ id: '42', email: 'alice@test.com' });
    expect(mockReceiptsStore.markEmailed).toHaveBeenCalledWith('42');
  });

  it('does not call backend email API for local-only receipt ids', async () => {
    receiptClient.listReceipts.mockResolvedValue({ receipts: [] });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    let emailResult: boolean | undefined;
    await act(async () => {
      emailResult = await result.current.email('receipt-local-1', 'alice@test.com');
    });

    expect(emailResult).toBe(false);
    expect(receiptClient.emailReceipt).not.toHaveBeenCalled();
    expect(mockReceiptsStore.markEmailed).not.toHaveBeenCalled();
  });

  it('email function returns false on error', async () => {
    receiptClient.listReceipts.mockResolvedValue({ receipts: [] });
    receiptClient.emailReceipt.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    let emailResult: boolean | undefined;
    await act(async () => {
      emailResult = await result.current.email('r1', 'alice@test.com');
    });

    expect(emailResult).toBe(false);
  });

  it('reload function refetches remote data', async () => {
    receiptClient.listReceipts.mockResolvedValue({ receipts: [] });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    receiptClient.listReceipts.mockResolvedValue({ receipts: sampleReceipts });

    await act(async () => {
      await result.current.reload();
    });

    expect(result.current.receipts).toHaveLength(2);
  });

  it('exposes add function from store', async () => {
    receiptClient.listReceipts.mockResolvedValue({ receipts: [] });

    const { result } = renderHook(() => useReceipts());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.add).toBe(mockReceiptsStore.add);
  });
});
