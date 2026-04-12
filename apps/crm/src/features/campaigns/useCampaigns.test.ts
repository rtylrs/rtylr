import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useCampaigns,
  useCampaignById,
  useCampaignStats,
  useCreateCampaign,
  useUpdateCampaign,
  useUpdateCampaignStatus,
} from './useCampaigns';

const mockCrmApi = {
  getCampaigns: vi.fn(),
  getCampaignById: vi.fn(),
  getCampaignStats: vi.fn(),
  createCampaign: vi.fn(),
  updateCampaign: vi.fn(),
  updateCampaignStatus: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/crmApi', () => ({
  crmApi: {
    getCampaigns: (...args: any[]) => mockCrmApi.getCampaigns(...args),
    getCampaignById: (...args: any[]) => mockCrmApi.getCampaignById(...args),
    getCampaignStats: (...args: any[]) => mockCrmApi.getCampaignStats(...args),
    createCampaign: (...args: any[]) => mockCrmApi.createCampaign(...args),
    updateCampaign: (...args: any[]) => mockCrmApi.updateCampaign(...args),
    updateCampaignStatus: (...args: any[]) => mockCrmApi.updateCampaignStatus(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleCampaigns = [
  {
    id: 1,
    name: 'Summer Sale',
    description: 'Summer deals',
    channel: 'email',
    status: 'active',
    audience: 5000,
    sent: 4500,
    openRate: '45%',
    clickRate: '12%',
    revenue: '$5000',
    updatedAt: '2026-04-01',
  },
  {
    id: 2,
    name: 'Winter Promo',
    description: 'Winter deals',
    channel: 'sms',
    status: 'draft',
    audience: 3000,
    sent: 0,
    openRate: '0%',
    clickRate: '0%',
    revenue: '$0',
    updatedAt: '2026-04-02',
  },
];

describe('useCampaigns', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches all campaigns', async () => {
    mockCrmApi.getCampaigns.mockResolvedValue(sampleCampaigns);

    const { result } = renderHook(() => useCampaigns(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(result.current.data![0].name).toBe('Summer Sale');
  });

  it('handles API error', async () => {
    mockCrmApi.getCampaigns.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useCampaigns(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useCampaignById', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single campaign', async () => {
    mockCrmApi.getCampaignById.mockResolvedValue(sampleCampaigns[0]);

    const { result } = renderHook(() => useCampaignById(1), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.name).toBe('Summer Sale');
    expect(mockCrmApi.getCampaignById).toHaveBeenCalledWith(1);
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useCampaignById(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCampaignStats', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches campaign stats', async () => {
    const stats = {
      id: 1,
      sent: 1000,
      delivered: 950,
      opened: 450,
      clicked: 200,
      unsubscribed: 5,
      bounced: 10,
      revenue: 5000,
    };
    mockCrmApi.getCampaignStats.mockResolvedValue(stats);

    const { result } = renderHook(() => useCampaignStats(1), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.sent).toBe(1000);
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useCampaignStats(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCreateCampaign', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createCampaign', async () => {
    mockCrmApi.createCampaign.mockResolvedValue({ id: 3, name: 'New Campaign' });

    const { result } = renderHook(() => useCreateCampaign(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ name: 'New Campaign', channel: 'email' });
    });

    expect(mockCrmApi.createCampaign).toHaveBeenCalledWith({
      name: 'New Campaign',
      channel: 'email',
    });
  });
});

describe('useUpdateCampaign', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateCampaign with id and data', async () => {
    mockCrmApi.updateCampaign.mockResolvedValue({ id: 1, name: 'Updated' });

    const { result } = renderHook(() => useUpdateCampaign(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ id: 1, data: { name: 'Updated' } });
    });

    expect(mockCrmApi.updateCampaign).toHaveBeenCalledWith(1, { name: 'Updated' });
  });
});

describe('useUpdateCampaignStatus', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateCampaignStatus with id and status', async () => {
    mockCrmApi.updateCampaignStatus.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useUpdateCampaignStatus(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ id: 1, status: 'paused' });
    });

    expect(mockCrmApi.updateCampaignStatus).toHaveBeenCalledWith(1, 'paused');
  });
});
