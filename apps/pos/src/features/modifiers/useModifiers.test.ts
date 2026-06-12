import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor, act } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useModifiers } from './useModifiers';

const mockPosService = vi.hoisted(() => ({
  getModifiers: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    getModifiers: (...args: unknown[]) => mockPosService.getModifiers(...args),
  },
}));

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
};

describe('useModifiers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads modifier groups with defaults and numeric coercion', async () => {
    mockPosService.getModifiers.mockResolvedValue({
      groups: [
        {
          id: 1,
          min: '1',
          max: '3',
          modifiers: [{ id: 2, name: 'Cheese', priceDelta: '1.5' }],
        },
      ],
    });

    const { result } = renderHook(() => useModifiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.groups).toEqual([
      {
        id: '1',
        name: 'Modifiers',
        min: 1,
        max: 3,
        modifiers: [
          {
            id: '2',
            name: 'Cheese',
            priceDelta: 1.5,
            group: 'Modifiers',
            required: false,
            defaultSelected: false,
          },
        ],
      },
    ]);
  });

  it('handles .data and non-array responses safely', async () => {
    mockPosService.getModifiers.mockResolvedValueOnce({ data: [{ id: '2', name: 'Sauces' }] });

    const first = renderHook(() => useModifiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(first.result.current.loading).toBe(false));

    expect(first.result.current.groups[0].name).toBe('Sauces');

    mockPosService.getModifiers.mockResolvedValueOnce({ groups: 'invalid' });

    const second = renderHook(() => useModifiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(second.result.current.loading).toBe(false));

    expect(second.result.current.groups).toEqual([]);
  });

  it('exposes clean error messages', async () => {
    mockPosService.getModifiers.mockRejectedValue(new Error('Failed to load modifiers'));

    const { result } = renderHook(() => useModifiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Failed to load modifiers');
    expect(result.current.groups).toEqual([]);
  });

  it('reload refetches modifier groups', async () => {
    mockPosService.getModifiers.mockResolvedValue({ groups: [] });

    const { result } = renderHook(() => useModifiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    mockPosService.getModifiers.mockResolvedValue({ groups: [{ id: '3', name: 'Toppings' }] });

    await act(async () => {
      await result.current.reload();
    });

    await waitFor(() => expect(result.current.groups).toHaveLength(1));
    expect(result.current.groups[0].name).toBe('Toppings');
  });
});
