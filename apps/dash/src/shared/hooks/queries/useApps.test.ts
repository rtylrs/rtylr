import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { APP_URLS } from '@voxire/config';
import { useApps } from './useApps';

const mockGet = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../lib/api', () => ({
  safe: { get: (...args: unknown[]) => mockGet(...args) },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useApps', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches and maps apps correctly', async () => {
    mockGet.mockResolvedValue({
      apps: [
        { id: 'pos', name: 'Point of Sale', description: 'POS app', isEnabled: true },
        { id: 'erp', name: 'ERP', description: 'ERP system', isEnabled: false },
      ],
    });

    const { result } = renderHook(() => useApps(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(result.current.data![0].key).toBe('pos');
    expect(result.current.data![0].label).toBe('Point of Sale');
    expect(result.current.data![0].status).toBe('active');
    expect(result.current.data![0].href).toBe(APP_URLS.pos);
  });

  it('maps disabled apps as coming_soon', async () => {
    mockGet.mockResolvedValue({
      apps: [{ id: 'erp', name: 'ERP', description: 'ERP', isEnabled: false }],
    });

    const { result } = renderHook(() => useApps(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data![0].status).toBe('coming_soon');
  });

  it('returns empty array when response is not an array', async () => {
    mockGet.mockResolvedValue({ apps: 'not an array' });

    const { result } = renderHook(() => useApps(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([]);
  });

  it('returns empty array when res is null', async () => {
    mockGet.mockResolvedValue(null);

    const { result } = renderHook(() => useApps(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([]);
  });

  it('uses fallback metadata for unknown app ids', async () => {
    mockGet.mockResolvedValue({
      apps: [{ id: 'unknown_app', name: 'Unknown', isEnabled: true }],
    });

    const { result } = renderHook(() => useApps(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data![0].href).toBe('#');
    expect(result.current.data![0].color).toBe('#888');
    expect(result.current.data![0].category).toBe('core');
  });

  it('handles API error', async () => {
    mockGet.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useApps(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBeDefined();
  });

  it('handles flat array response (no .apps wrapper)', async () => {
    mockGet.mockResolvedValue([{ id: 'crm', name: 'CRM', isEnabled: true }]);

    const { result } = renderHook(() => useApps(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
    expect(result.current.data![0].key).toBe('crm');
  });

  it('calls safe.get with /api/hub/app/v1/apps', async () => {
    mockGet.mockResolvedValue({ apps: [] });

    renderHook(() => useApps(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockGet).toHaveBeenCalledWith('/api/hub/app/v1/apps'));
  });
});
