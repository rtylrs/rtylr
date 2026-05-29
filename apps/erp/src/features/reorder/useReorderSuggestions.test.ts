import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useReorderSuggestions } from './useReorderSuggestions';

const mockErpApi = {
  getReorderSuggestions: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    getReorderSuggestions: (...args: unknown[]) => mockErpApi.getReorderSuggestions(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleSuggestions = [
  {
    productId: 1,
    name: 'Widget A',
    sku: 'WA-001',
    currentStock: 5,
    avgDailySales: 2,
    suggestedQuantity: 50,
    leadTimeDays: 7,
    confidence: 'high' as const,
  },
  {
    productId: 2,
    name: 'Widget B',
    sku: 'WB-001',
    currentStock: 0,
    avgDailySales: 1,
    suggestedQuantity: 30,
    leadTimeDays: 5,
    confidence: 'medium' as const,
  },
];

describe('useReorderSuggestions', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches reorder suggestions', async () => {
    mockErpApi.getReorderSuggestions.mockResolvedValue(sampleSuggestions);
    const { result } = renderHook(() => useReorderSuggestions(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when API returns null', async () => {
    mockErpApi.getReorderSuggestions.mockResolvedValue(null);
    const { result } = renderHook(() => useReorderSuggestions(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockErpApi.getReorderSuggestions.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useReorderSuggestions(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('uses "reorder-suggestions" query key', async () => {
    mockErpApi.getReorderSuggestions.mockResolvedValue([]);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children);
    renderHook(() => useReorderSuggestions(), { wrapper });
    await waitFor(() => expect(client.getQueryData(['reorder-suggestions'])).toEqual([]));
  });
});
