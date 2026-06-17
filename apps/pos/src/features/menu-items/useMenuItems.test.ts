import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useMenuItems } from './useMenuItems';

const mockPosService = {
  getMenuItems: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    getMenuItems: (...args: unknown[]) => mockPosService.getMenuItems(...args),
  },
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

  it('loads menu items on mount', async () => {
    mockPosService.getMenuItems.mockResolvedValue({
      items: [
        { id: '1', name: 'Burger', price: 12.99, category: 'Mains' },
        { id: '2', name: 'Fries', price: 4.99, category: 'Sides' },
        { id: '3', name: 'Cola', price: 2.99, category: 'Drinks' },
      ],
    });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.items).toHaveLength(3);
    expect(result.current.items[0].name).toBe('Burger');
  });

  it('computes sorted unique categories', async () => {
    mockPosService.getMenuItems.mockResolvedValue({
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
    mockPosService.getMenuItems.mockRejectedValue(new Error('Failed to load'));

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('Failed to load');
    expect(result.current.items).toEqual([]);
  });

  it('handles .data key in response', async () => {
    mockPosService.getMenuItems.mockResolvedValue({
      data: [{ id: '1', name: 'Pizza', price: 15 }],
    });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.items).toHaveLength(1);
    expect(result.current.items[0].name).toBe('Pizza');
  });

  it('defaults missing fields', async () => {
    mockPosService.getMenuItems.mockResolvedValue({
      items: [{ id: '1' }],
    });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.items[0].name).toBe('Item');
    expect(result.current.items[0].price).toBe(0);
    expect(result.current.items[0].category).toBe('Other');
    expect(result.current.items[0].available).toBe(true);
  });

  it('reload function refetches data', async () => {
    mockPosService.getMenuItems.mockResolvedValue({ items: [] });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    mockPosService.getMenuItems.mockResolvedValue({
      items: [{ id: '10', name: 'New Item', price: 9.99, category: 'New' }],
    });

    await act(async () => {
      result.current.reload();
      await waitFor(() => expect(result.current.items).toHaveLength(1));
    });

    expect(result.current.items[0].name).toBe('New Item');
  });

  it('handles non-array response gracefully', async () => {
    mockPosService.getMenuItems.mockResolvedValue({ items: 'invalid' });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.items).toEqual([]);
  });

  it('maps available field correctly', async () => {
    mockPosService.getMenuItems.mockResolvedValue({
      items: [
        { id: '1', name: 'A', price: 10, available: false },
        { id: '2', name: 'B', price: 10, available: true },
      ],
    });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.items[0].available).toBe(false);
    expect(result.current.items[1].available).toBe(true);
  });

  it('normalizes ERP-backed product aliases', async () => {
    mockPosService.getMenuItems.mockResolvedValue({
      items: [
        {
          id: 11,
          name: 'Creatine Biotech',
          sale_price: '30.00',
          category_name: 'Creatine',
          image_url: 'https://example.com/creatine.webp',
          is_active: true,
          sku: 'SUPLAB-CREATINE-BIOTECH',
          stock: '22.000',
        },
      ],
    });

    const { result } = renderHook(() => useMenuItems(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.items[0]).toMatchObject({
      id: '11',
      price: 30,
      category: 'Creatine',
      image: 'https://example.com/creatine.webp',
      available: true,
      sku: 'SUPLAB-CREATINE-BIOTECH',
      stock: 22,
    });
  });
});
