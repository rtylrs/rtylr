import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor, act } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useModifiers } from './useModifiers';

const modifierClient = vi.hoisted(() => ({
  listModifiers: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => modifierClient,
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

  it('groups generated modifier rows for the UI', async () => {
    modifierClient.listModifiers.mockResolvedValue({
      modifiers: [
        {
          id: '2',
          name: 'Cheese',
          groupName: 'Toppings',
          priceDelta: 1.5,
          isRequired: true,
        },
      ],
    });

    const { result } = renderHook(() => useModifiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(modifierClient.listModifiers).toHaveBeenCalledWith({});
    expect(result.current.groups).toEqual([
      {
        id: 'toppings',
        name: 'Toppings',
        min: 1,
        max: 1,
        modifiers: [
          {
            id: '2',
            name: 'Cheese',
            priceDelta: 1.5,
            group: 'Toppings',
            required: true,
            defaultSelected: false,
          },
        ],
      },
    ]);
  });

  it('exposes clean error messages', async () => {
    modifierClient.listModifiers.mockRejectedValue(new Error('Failed to load modifiers'));

    const { result } = renderHook(() => useModifiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Failed to load modifiers');
    expect(result.current.groups).toEqual([]);
  });

  it('reload refetches modifier groups', async () => {
    modifierClient.listModifiers.mockResolvedValue({ modifiers: [] });

    const { result } = renderHook(() => useModifiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    modifierClient.listModifiers.mockResolvedValue({
      modifiers: [{
        id: '3',
        name: 'Sauce',
        groupName: 'Toppings',
        priceDelta: 0,
        isRequired: false,
      }],
    });

    await act(async () => {
      await result.current.reload();
    });

    await waitFor(() => expect(result.current.groups).toHaveLength(1));
    expect(result.current.groups[0].name).toBe('Toppings');
  });
});
