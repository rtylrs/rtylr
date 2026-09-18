// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockClient = vi.hoisted(() => ({
  listRecurringInvoices: vi.fn(),
  createRecurringInvoice: vi.fn(),
  pauseRecurringInvoice: vi.fn(),
  resumeRecurringInvoice: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  unwrap: (val: unknown) => Promise.resolve(val),
  oneKitResult: (op: () => Promise<unknown>) => op(),
  // clients.ts uses the explicit REST fallback for the few calls no contract
  // declares, so the mock covers that transport as well as the generated one.
  createRestRouteClient: () => ({
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

import {
  useRecurringInvoices,
  useCreateRecurringInvoice,
  useSetRecurringStatus,
} from './useRecurringInvoices';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useRecurringInvoices', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches recurring invoices on mount', async () => {
    mockClient.listRecurringInvoices.mockResolvedValue({
      recurringInvoices: [
        {
          id: '1',
          ownerId: '1',
          templateInvoiceId: '42',
          frequency: 'monthly',
          nextInvoiceDate: '2026-07-01',
          endDate: null,
          isActive: true,
          createdAt: '',
        },
      ],
    });
    const { result } = renderHook(() => useRecurringInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.listRecurringInvoices).toHaveBeenCalledWith({});
    expect(result.current.data).toHaveLength(1);
  });

  it('starts in loading state', () => {
    mockClient.listRecurringInvoices.mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useRecurringInvoices(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('returns empty array when list is empty', async () => {
    mockClient.listRecurringInvoices.mockResolvedValue({ recurringInvoices: [] });
    const { result } = renderHook(() => useRecurringInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('parses the backend payload shape', async () => {
    mockClient.listRecurringInvoices.mockResolvedValue({
      recurringInvoices: [
        {
          id: '7',
          ownerId: '1',
          templateInvoiceId: '9',
          frequency: 'weekly',
          nextInvoiceDate: null,
          endDate: null,
          isActive: false,
          createdAt: '',
        },
      ],
    });
    const { result } = renderHook(() => useRecurringInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.id).toBe('7');
    expect(result.current.data?.[0]?.isActive).toBe(false);
  });
});

describe('useCreateRecurringInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('creates a recurring invoice', async () => {
    mockClient.createRecurringInvoice.mockResolvedValue({ id: '5', message: 'ok' });
    const { result } = renderHook(() => useCreateRecurringInvoice(), { wrapper: createWrapper() });
    await act(() =>
      result.current.mutateAsync({
        templateInvoiceId: '42',
        frequency: 'monthly',
        nextInvoiceDate: '2026-07-01',
        isActive: true,
      })
    );
    expect(mockClient.createRecurringInvoice).toHaveBeenCalledWith(
      expect.objectContaining({ templateInvoiceId: '42', frequency: 'monthly' })
    );
  });
});

describe('useSetRecurringStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('pauses a recurring invoice', async () => {
    mockClient.pauseRecurringInvoice.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useSetRecurringStatus(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: '1', status: 'paused' }));
    expect(mockClient.pauseRecurringInvoice).toHaveBeenCalledWith({ id: '1' });
  });

  it('resumes a recurring invoice', async () => {
    mockClient.resumeRecurringInvoice.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useSetRecurringStatus(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: '1', status: 'active' }));
    expect(mockClient.resumeRecurringInvoice).toHaveBeenCalledWith({ id: '1' });
  });
});
