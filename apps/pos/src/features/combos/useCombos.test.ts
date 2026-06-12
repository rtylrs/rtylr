import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor, act } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCombos } from './useCombos';

const mockPosService = vi.hoisted(() => ({
  getCombos: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    getCombos: (...args: unknown[]) => mockPosService.getCombos(...args),
  },
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

  it('loads combos with schema defaults and numeric coercion', async () => {
    mockPosService.getCombos.mockResolvedValue({
      combos: [{ id: 1, price: '20', items: [{ itemId: 2, qty: '2' }] }],
    });

    const { result } = renderHook(() => useCombos(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.combos).toEqual([
      {
        id: '1',
        name: 'Combo',
        price: 20,
        description: undefined,
        items: [{ itemId: '2', qty: 2 }],
      },
    ]);
  });

  it('handles .data and non-array responses safely', async () => {
    mockPosService.getCombos.mockResolvedValueOnce({ data: [{ id: '2', name: 'Lunch' }] });

    const first = renderHook(() => useCombos(), { wrapper: createWrapper() });
    await waitFor(() => expect(first.result.current.loading).toBe(false));

    expect(first.result.current.combos[0].name).toBe('Lunch');

    mockPosService.getCombos.mockResolvedValueOnce({ combos: 'invalid' });

    const second = renderHook(() => useCombos(), { wrapper: createWrapper() });
    await waitFor(() => expect(second.result.current.loading).toBe(false));

    expect(second.result.current.combos).toEqual([]);
  });

  it('exposes clean error messages', async () => {
    mockPosService.getCombos.mockRejectedValue(new Error('Failed to load combos'));

    const { result } = renderHook(() => useCombos(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Failed to load combos');
    expect(result.current.combos).toEqual([]);
  });

  it('reload refetches combos', async () => {
    mockPosService.getCombos.mockResolvedValue({ combos: [] });

    const { result } = renderHook(() => useCombos(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    mockPosService.getCombos.mockResolvedValue({ combos: [{ id: '3', name: 'Dinner' }] });

    await act(async () => {
      await result.current.reload();
    });

    await waitFor(() => expect(result.current.combos).toHaveLength(1));
    expect(result.current.combos[0].name).toBe('Dinner');
  });
});
