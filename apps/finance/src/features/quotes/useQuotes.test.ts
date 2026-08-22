// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockClient = vi.hoisted(() => ({
  listQuotes: vi.fn(),
  createQuote: vi.fn(),
  updateQuoteStatus: vi.fn(),
  convertQuote: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { useQuotes, useCreateQuote, useUpdateQuoteStatus, useConvertQuote } from './useQuotes';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

const baseQuote = {
  id: '1',
  ownerId: '1',
  customerName: 'Alpha',
  customerEmail: '',
  status: 'draft',
  lines: [],
  subtotal: 100,
  tax: 0,
  discount: 0,
  total: 100,
  validUntil: '2026-07-01',
  notes: '',
  createdAt: '2026-06-01',
  updatedAt: '',
};

describe('useQuotes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches quotes on mount with default pagination', async () => {
    mockClient.listQuotes.mockResolvedValue({
      quotes: [baseQuote],
      hasMore: false,
      page: 1,
      pageSize: 50,
    });
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.listQuotes).toHaveBeenCalledWith({ page: 1, pageSize: 50 });
    expect(result.current.data?.quotes[0]?.client).toBe('Alpha');
  });

  it('accepts custom page and pageSize params', async () => {
    mockClient.listQuotes.mockResolvedValue({ quotes: [], hasMore: false, page: 2, pageSize: 20 });
    const { result } = renderHook(() => useQuotes({ page: 2, pageSize: 20 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.listQuotes).toHaveBeenCalledWith({ page: 2, pageSize: 20 });
  });

  it('exposes hasMore from response', async () => {
    mockClient.listQuotes.mockResolvedValue({
      quotes: [{ ...baseQuote, id: '2', customerName: 'Beta' }],
      hasMore: true,
      page: 1,
      pageSize: 50,
    });
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.hasMore).toBe(true);
  });

  it('starts in loading state', () => {
    mockClient.listQuotes.mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('keeps quote id as a string', async () => {
    mockClient.listQuotes.mockResolvedValue({
      quotes: [{ ...baseQuote, id: '42', customerName: 'Gamma' }],
      hasMore: false,
      page: 1,
      pageSize: 50,
    });
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.quotes[0]?.id).toBe('42');
  });

  it('maps backend status to the capitalized UI label', async () => {
    mockClient.listQuotes.mockResolvedValue({
      quotes: [{ ...baseQuote, status: 'rejected' }],
      hasMore: false,
      page: 1,
      pageSize: 50,
    });
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.quotes[0]?.status).toBe('Declined');
  });
});

describe('useCreateQuote', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('creates a quote', async () => {
    mockClient.createQuote.mockResolvedValue({ quote: baseQuote });
    const { result } = renderHook(() => useCreateQuote(), { wrapper: createWrapper() });
    await act(() =>
      result.current.mutateAsync({
        client: 'Alpha',
        amount: 100,
        expiresAt: '2026-07-01',
      })
    );
    expect(mockClient.createQuote).toHaveBeenCalledWith(
      expect.objectContaining({ customerName: 'Alpha', total: 100 })
    );
  });
});

describe('useUpdateQuoteStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('updates quote status', async () => {
    mockClient.updateQuoteStatus.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useUpdateQuoteStatus(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: '1', status: 'accepted' }));
    expect(mockClient.updateQuoteStatus).toHaveBeenCalledWith({
      id: '1',
      body: { status: 'accepted' },
    });
  });
});

describe('useConvertQuote', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('converts a quote to an invoice', async () => {
    mockClient.convertQuote.mockResolvedValue({ invoiceId: '9', message: 'ok' });
    const { result } = renderHook(() => useConvertQuote(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync('1'));
    expect(mockClient.convertQuote).toHaveBeenCalledWith({ id: '1' });
  });
});
