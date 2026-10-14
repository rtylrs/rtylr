import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';

const lookupBusinessBySlugMock = vi.hoisted(() => vi.fn());
const getPublicBusinessMock = vi.hoisted(() => vi.fn());
const listPublicMenuItemsMock = vi.hoisted(() => vi.fn());

vi.mock('../lib/clients', () => ({
  businessClient: {
    lookupBusinessBySlug: lookupBusinessBySlugMock,
    getPublicBusiness: getPublicBusinessMock,
  },
  menuItemClient: {
    listPublicMenuItems: listPublicMenuItemsMock,
  },
}));

vi.mock('@voxire/config', () => ({
  cdnImageUrl: (path: string) => `https://cdn.example.com/${path}`,
}));

import { usePublicBusinessBySlug, usePublicMenu } from './usePublicMenu';

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('usePublicBusinessBySlug hook', () => {
  beforeEach(() => {
    lookupBusinessBySlugMock.mockClear();
  });

  it('fetches and maps business bundle by slug', async () => {
    lookupBusinessBySlugMock.mockResolvedValueOnce({
      business: {
        id: '42',
        name: 'Acme Cafe',
        slug: 'acme',
        logoUrl: 'acme-logo.png',
        currency: 'EUR',
        timezone: 'Europe/Paris',
      },
      locations: [
        {
          id: '101',
          name: 'Downtown',
          slug: 'downtown',
          address: '123 Main St',
          timezone: 'Europe/Paris',
          isActive: true,
        },
      ],
    });

    const { result } = renderHook(() => usePublicBusinessBySlug('acme'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.data).toBeDefined());

    expect(lookupBusinessBySlugMock).toHaveBeenCalledWith({ slug: 'acme' });
    expect(result.current.data?.business).toEqual({
      id: 42,
      name: 'Acme Cafe',
      slug: 'acme',
      logoUrl: 'https://cdn.example.com/acme-logo.png',
      currency: 'EUR',
      timezone: 'Europe/Paris',
    });
    expect(result.current.data?.locations).toEqual([
      {
        id: 101,
        name: 'Downtown',
        slug: 'downtown',
        address: '123 Main St',
        timezone: 'Europe/Paris',
        isActive: true,
      },
    ]);
  });

  it('does not execute query when slug is null or empty', () => {
    const { result } = renderHook(() => usePublicBusinessBySlug(''), {
      wrapper: createWrapper(),
    });

    expect(lookupBusinessBySlugMock).not.toHaveBeenCalled();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toBeUndefined();
  });
});

describe('usePublicMenu hook', () => {
  beforeEach(() => {
    getPublicBusinessMock.mockClear();
    listPublicMenuItemsMock.mockClear();
  });

  it('fetches business and items in parallel', async () => {
    getPublicBusinessMock.mockResolvedValueOnce({
      business: {
        id: '42',
        name: 'Acme Cafe',
        logoUrl: 'logo.jpg',
        currency: 'USD',
        timezone: 'UTC',
      },
    });

    listPublicMenuItemsMock.mockResolvedValueOnce({
      items: [
        {
          id: '1',
          ownerId: '42',
          categoryId: '5',
          name: 'Iced Coffee',
          sku: 'IC-01',
          salePrice: 4.5,
          imageUrl: 'coffee.jpg',
          isActive: true,
          createdAt: '2026-01-01',
          updatedAt: '2026-01-02',
        },
      ],
    });

    const { result } = renderHook(() => usePublicMenu(42, 101), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.data).toBeDefined());

    expect(getPublicBusinessMock).toHaveBeenCalledWith({ businessId: '42' });
    expect(listPublicMenuItemsMock).toHaveBeenCalledWith({
      businessId: '42',
      locationId: '101',
    });

    expect(result.current.data?.business).toEqual({
      id: 42,
      name: 'Acme Cafe',
      logoUrl: 'https://cdn.example.com/logo.jpg',
      currency: 'USD',
      timezone: 'UTC',
    });

    expect(result.current.data?.items).toEqual([
      {
        id: 1,
        ownerId: 42,
        categoryId: 5,
        name: 'Iced Coffee',
        sku: 'IC-01',
        salePrice: 4.5,
        imageUrl: 'https://cdn.example.com/coffee.jpg',
        isActive: true,
        createdAt: '2026-01-01',
        updatedAt: '2026-01-02',
      },
    ]);
  });

  it('does not execute query when businessId is null', () => {
    const { result } = renderHook(() => usePublicMenu(null), {
      wrapper: createWrapper(),
    });

    expect(getPublicBusinessMock).not.toHaveBeenCalled();
    expect(listPublicMenuItemsMock).not.toHaveBeenCalled();
    expect(result.current.data).toBeUndefined();
  });
});
