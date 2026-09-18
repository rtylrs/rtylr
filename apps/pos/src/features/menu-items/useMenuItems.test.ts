import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useMenuItems } from './useMenuItems';

const menuItemClient = vi.hoisted(() => ({
  listMenuItems: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => menuItemClient,
}));

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
};

describe('useMenuItems', () => {
  beforeEach(() => vi.clearAllMocks());

  it('loads menu items from the generated menu endpoint', async () => {
    menuItemClient.listMenuItems.mockResolvedValue({
      items: [
        { id: '1', name: 'Burger', price: 12.99, category: 'Mains' },
        { id: '2', name: 'Fries', price: 4.99, category: 'Sides' },
        { id: '3', name: 'Cola', price: 2.99, category: 'Drinks' },
      ],
    });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(menuItemClient.listMenuItems).toHaveBeenCalledWith({});
    expect(result.current.items).toHaveLength(3);
    expect(result.current.items[0].name).toBe('Burger');
  });

  it('computes sorted unique categories', async () => {
    menuItemClient.listMenuItems.mockResolvedValue({
      items: [
        { id: '1', name: 'Burger', price: 12, category: 'Mains' },
        { id: '2', name: 'Fries', price: 5, category: 'Sides' },
        { id: '3', name: 'Steak', price: 25, category: 'Mains' },
        { id: '4', name: 'Cola', price: 3, category: 'Drinks' },
      ],
    });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.categories).toEqual(['Drinks', 'Mains', 'Sides']);
  });

  it('handles error state', async () => {
    menuItemClient.listMenuItems.mockRejectedValue(new Error('Failed to load'));

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Failed to load');
    expect(result.current.items).toEqual([]);
  });

  it('reload refetches data', async () => {
    menuItemClient.listMenuItems.mockResolvedValue({ items: [] });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    menuItemClient.listMenuItems.mockResolvedValue({
      items: [{ id: '10', name: 'New Item', price: 9.99, category: 'New' }],
    });

    await act(async () => {
      result.current.reload();
      await waitFor(() => expect(result.current.items).toHaveLength(1));
    });

    expect(result.current.items[0].name).toBe('New Item');
  });
});
