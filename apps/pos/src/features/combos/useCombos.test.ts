import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor, act } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCombos } from './useCombos';

const comboClient = vi.hoisted(() => ({
  listCombos: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => comboClient,
}));

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
};

describe('useCombos', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads combos from generated combo rows', async () => {
    comboClient.listCombos.mockResolvedValue({
      combos: [{ id: '1', name: 'Lunch', price: 20, items: JSON.stringify([{ itemId: 2, qty: 2 }]) }],
    });

    const { result } = renderHook(() => useCombos(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(comboClient.listCombos).toHaveBeenCalledWith({});
    expect(result.current.combos).toEqual([
      {
        id: '1',
        name: 'Lunch',
        price: 20,
        description: undefined,
        items: [{ itemId: '2', qty: 2 }],
      },
    ]);
  });

  it('exposes clean error messages', async () => {
    comboClient.listCombos.mockRejectedValue(new Error('Failed to load combos'));

    const { result } = renderHook(() => useCombos(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Failed to load combos');
    expect(result.current.combos).toEqual([]);
  });

  it('reload refetches combos', async () => {
    comboClient.listCombos.mockResolvedValue({ combos: [] });

    const { result } = renderHook(() => useCombos(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    comboClient.listCombos.mockResolvedValue({
      combos: [{ id: '3', name: 'Dinner', price: 15, items: '' }],
    });

    await act(async () => {
      await result.current.reload();
    });

    await waitFor(() => expect(result.current.combos).toHaveLength(1));
    expect(result.current.combos[0].name).toBe('Dinner');
  });
});
