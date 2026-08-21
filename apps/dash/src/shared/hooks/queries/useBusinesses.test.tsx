import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import {
  useCreateSubBusiness,
  useDeleteBusiness,
  useListMyBusinesses,
  useRenameBusiness,
  useSwitchBusiness,
} from './useBusinesses';

const mockBusinessClient = vi.hoisted(() => ({
  listMyBusinesses: vi.fn(),
  renameBusiness: vi.fn(),
  deleteBusiness: vi.fn(),
  createSubBusiness: vi.fn(),
}));
const mockUserClient = vi.hoisted(() => ({ switchBusiness: vi.fn() }));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../lib/clients', () => ({
  businessClient: mockBusinessClient,
  userClient: mockUserClient,
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('business queries', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('normalizes numeric IDs returned as JSON strings', async () => {
    mockBusinessClient.listMyBusinesses.mockResolvedValue({
      businesses: [
        {
          id: '42',
          name: 'Acme',
          logoUrl: '',
          parentId: '7',
          isActive: true,
          isOwned: true,
        },
      ],
    });

    const { result } = renderHook(() => useListMyBusinesses(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.businesses[0]).toMatchObject({ id: 42, parentId: 7 });
  });

  it('renames a business through the business API', async () => {
    mockBusinessClient.renameBusiness.mockResolvedValue({
      business: { id: '42', name: 'New name' },
    });

    const { result } = renderHook(() => useRenameBusiness(), { wrapper: createWrapper() });

    act(() => {
      result.current.mutate({ businessId: 42, name: 'New name' });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockBusinessClient.renameBusiness).toHaveBeenCalledWith({
      businessId: '42',
      name: 'New name',
    });
  });

  it('deletes a business through the business API', async () => {
    mockBusinessClient.deleteBusiness.mockResolvedValue({ message: 'Business deleted' });

    const { result } = renderHook(() => useDeleteBusiness(), { wrapper: createWrapper() });

    act(() => {
      result.current.mutate(42);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockBusinessClient.deleteBusiness).toHaveBeenCalledWith({ businessId: '42' });
  });

  it('sends the int64 parent business ID as a JSON string when creating', async () => {
    mockBusinessClient.createSubBusiness.mockResolvedValue({ id: '99' });

    const { result } = renderHook(() => useCreateSubBusiness(), { wrapper: createWrapper() });

    act(() => {
      result.current.mutate({ name: 'Second branch', parentBusinessId: 42 });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockBusinessClient.createSubBusiness).toHaveBeenCalledWith({
      name: 'Second branch',
      parentBusinessId: '42',
    });
  });

  it('sends int64 business IDs as JSON strings when switching', async () => {
    mockUserClient.switchBusiness.mockResolvedValue({ token: 'new-token' });

    const { result } = renderHook(() => useSwitchBusiness(), { wrapper: createWrapper() });

    act(() => {
      result.current.mutate(42);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockUserClient.switchBusiness).toHaveBeenCalledWith({ businessId: '42' });
  });
});
