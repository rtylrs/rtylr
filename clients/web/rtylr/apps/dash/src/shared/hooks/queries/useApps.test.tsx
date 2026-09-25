// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { APP_URLS } from '@voxire/config';
import { useApps } from './useApps';

const mockAppClient = vi.hoisted(() => ({ listApps: vi.fn() }));

vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<object>();
  return {
    ...actual,
    unwrap: vi.fn((v) => v),
  };
});

vi.mock('../../lib/clients', () => ({ appClient: mockAppClient }));

let queryClient: QueryClient;

const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

describe('useApps', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
  });

  it('fetches and maps apps correctly', async () => {
    mockAppClient.listApps.mockResolvedValue({
      apps: [
        { id: 'pos', name: 'Point of Sale', description: 'POS app', isEnabled: true },
        { id: 'erp', name: 'ERP', description: 'ERP system', isEnabled: false },
      ],
    });

    const { result } = renderHook(() => useApps(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(result.current.data![0].key).toBe('pos');
    expect(result.current.data![0].label).toBe('Point of Sale');
    expect(result.current.data![0].status).toBe('active');
    expect(result.current.data![0].href).toBe(APP_URLS.pos);
  });

  it('maps disabled apps as coming_soon', async () => {
    mockAppClient.listApps.mockResolvedValue({
      apps: [{ id: 'erp', name: 'ERP', description: 'ERP', isEnabled: false }],
    });

    const { result } = renderHook(() => useApps(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data![0].status).toBe('coming_soon');
  });

  it('returns empty array when response is not an array', async () => {
    mockAppClient.listApps.mockResolvedValue({ apps: 'not an array' });

    const { result } = renderHook(() => useApps(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([]);
  });

  it('returns empty array when res is null', async () => {
    mockAppClient.listApps.mockResolvedValue(null);

    const { result } = renderHook(() => useApps(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([]);
  });

  it('uses fallback metadata for unknown app ids', async () => {
    mockAppClient.listApps.mockResolvedValue({
      apps: [{ id: 'unknown_app', name: 'Unknown', isEnabled: true }],
    });

    const { result } = renderHook(() => useApps(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data![0].href).toBe('#');
    expect(result.current.data![0].color).toBe('#888');
    expect(result.current.data![0].category).toBe('core');
  });

  it('handles API error', async () => {
    mockAppClient.listApps.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useApps(), { wrapper });
    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBeDefined();
  });

  it('handles flat array response (no .apps wrapper)', async () => {
    mockAppClient.listApps.mockResolvedValue([{ id: 'crm', name: 'CRM', isEnabled: true }]);

    const { result } = renderHook(() => useApps(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
    expect(result.current.data![0].key).toBe('crm');
  });

  it('calls the generated app client', async () => {
    mockAppClient.listApps.mockResolvedValue({ apps: [] });

    renderHook(() => useApps(), { wrapper });
    await waitFor(() => expect(mockAppClient.listApps).toHaveBeenCalledWith({}));
  });
});
