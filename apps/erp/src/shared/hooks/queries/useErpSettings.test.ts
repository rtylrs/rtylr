import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { mockClient } = vi.hoisted(() => ({
  mockClient: {
    getErpSettings: vi.fn(),
    updateErpSettings: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: vi.fn(() => mockClient),
  getOneKitErrorMessage: vi.fn((_error, fallback) => fallback),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

import { useErpSettings, useUpdateErpSettings } from './useErpSettings';
import { useSettings } from '../../lib/settingsStore';

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

const apiSettings = {
  currency: 'EUR',
  taxRate: 20,
  lowStockThreshold: 4,
  defaultWarehouseId: '7',
  notifyOnLowStock: false,
  notifyOnExpiry: true,
};

describe('useErpSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useSettings.getState().setBusinessSettings({
      currency: 'USD',
      taxRate: 8.25,
      lowStockThreshold: 10,
      defaultWarehouseId: 1,
      notifyOnLowStock: true,
      notifyOnExpiry: true,
    });
  });

  it('loads tenant settings and hydrates the shared business state', async () => {
    mockClient.getErpSettings.mockResolvedValue({ settings: apiSettings });

    const { result } = renderHook(() => useErpSettings(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockClient.getErpSettings).toHaveBeenCalledWith({});
    expect(result.current.data).toEqual({
      currency: 'EUR',
      taxRate: 20,
      lowStockThreshold: 4,
      defaultWarehouseId: 7,
      notifyOnLowStock: false,
      notifyOnExpiry: true,
    });
    expect(useSettings.getState().taxRate).toBe(20);
  });

  it('saves all tenant business fields through the ERP contract', async () => {
    mockClient.updateErpSettings.mockResolvedValue({ settings: apiSettings });

    const { result } = renderHook(() => useUpdateErpSettings(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await result.current.mutateAsync({
        currency: 'EUR',
        taxRate: 20,
        lowStockThreshold: 4,
        defaultWarehouseId: 7,
        notifyOnLowStock: false,
        notifyOnExpiry: true,
      });
    });

    expect(mockClient.updateErpSettings).toHaveBeenCalledWith({
      currency: 'EUR',
      taxRate: 20,
      lowStockThreshold: 4,
      defaultWarehouseId: '7',
      notifyOnLowStock: false,
      notifyOnExpiry: true,
    });
    expect(useSettings.getState().defaultWarehouseId).toBe(7);
  });
});
