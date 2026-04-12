import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useLoyaltyTransactions,
  useLoyaltyConfig,
  useLoyaltyMembers,
  useLoyaltyRewards,
  useLoyaltyTiers,
  useUpdateLoyaltyTier,
  useUpdateLoyaltyConfig,
  useAwardPoints,
  useRedeemPoints,
} from './useLoyalty';

const mockCrmApi = {
  getLoyaltyTransactions: vi.fn(),
  getLoyaltyConfig: vi.fn(),
  getLoyaltyMembers: vi.fn(),
  getLoyaltyRewards: vi.fn(),
  getLoyaltyTiers: vi.fn(),
  updateLoyaltyTier: vi.fn(),
  updateLoyaltyConfig: vi.fn(),
  awardPoints: vi.fn(),
  redeemPoints: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/crmApi', () => ({
  crmApi: {
    getLoyaltyTransactions: (...args: any[]) => mockCrmApi.getLoyaltyTransactions(...args),
    getLoyaltyConfig: (...args: any[]) => mockCrmApi.getLoyaltyConfig(...args),
    getLoyaltyMembers: (...args: any[]) => mockCrmApi.getLoyaltyMembers(...args),
    getLoyaltyRewards: (...args: any[]) => mockCrmApi.getLoyaltyRewards(...args),
    getLoyaltyTiers: (...args: any[]) => mockCrmApi.getLoyaltyTiers(...args),
    updateLoyaltyTier: (...args: any[]) => mockCrmApi.updateLoyaltyTier(...args),
    updateLoyaltyConfig: (...args: any[]) => mockCrmApi.updateLoyaltyConfig(...args),
    awardPoints: (...args: any[]) => mockCrmApi.awardPoints(...args),
    redeemPoints: (...args: any[]) => mockCrmApi.redeemPoints(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useLoyaltyTransactions', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches loyalty transactions', async () => {
    const txns = [
      { customer: 'Alice', action: 'Earned 100 points', time: '2026-04-01', type: 'earn' },
    ];
    mockCrmApi.getLoyaltyTransactions.mockResolvedValue(txns);

    const { result } = renderHook(() => useLoyaltyTransactions(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
  });

  it('handles error', async () => {
    mockCrmApi.getLoyaltyTransactions.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useLoyaltyTransactions(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useLoyaltyConfig', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches loyalty config', async () => {
    const config = {
      programEnabled: true,
      pointsPerDollar: 10,
      pointsExpiry: 365,
      redemptionRate: 0.01,
    };
    mockCrmApi.getLoyaltyConfig.mockResolvedValue(config);

    const { result } = renderHook(() => useLoyaltyConfig(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.pointsPerDollar).toBe(10);
  });
});

describe('useLoyaltyMembers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches loyalty members', async () => {
    mockCrmApi.getLoyaltyMembers.mockResolvedValue([
      {
        id: 1,
        name: 'Alice',
        initials: 'AL',
        points: 500,
        tier: 'Gold',
        visits: 20,
        joinedAt: '2025-01-01',
        lastActive: '2026-04-01',
      },
    ]);

    const { result } = renderHook(() => useLoyaltyMembers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
  });
});

describe('useLoyaltyRewards', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches loyalty rewards', async () => {
    mockCrmApi.getLoyaltyRewards.mockResolvedValue([
      {
        id: 1,
        name: 'Free Coffee',
        description: 'A free coffee drink',
        pointsCost: 200,
        category: 'Beverages',
        stock: 50,
        active: true,
      },
    ]);

    const { result } = renderHook(() => useLoyaltyRewards(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
  });
});

describe('useLoyaltyTiers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches loyalty tiers', async () => {
    mockCrmApi.getLoyaltyTiers.mockResolvedValue([
      { id: 1, name: 'Bronze', threshold: 0, multiplier: 1, color: '#CD7F32' },
      { id: 2, name: 'Silver', threshold: 500, multiplier: 1.5, color: '#C0C0C0' },
    ]);

    const { result } = renderHook(() => useLoyaltyTiers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
  });
});

describe('useUpdateLoyaltyTier', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateLoyaltyTier with id and data', async () => {
    mockCrmApi.updateLoyaltyTier.mockResolvedValue({ id: 1, name: 'Gold' });

    const { result } = renderHook(() => useUpdateLoyaltyTier(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ id: 1, data: { name: 'Gold' } });
    });

    expect(mockCrmApi.updateLoyaltyTier).toHaveBeenCalledWith(1, { name: 'Gold' });
  });
});

describe('useUpdateLoyaltyConfig', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateLoyaltyConfig with config data', async () => {
    mockCrmApi.updateLoyaltyConfig.mockResolvedValue({ pointsPerDollar: 15 });

    const { result } = renderHook(() => useUpdateLoyaltyConfig(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ pointsPerDollar: 15 });
    });

    expect(mockCrmApi.updateLoyaltyConfig).toHaveBeenCalledWith({ pointsPerDollar: 15 });
  });
});

describe('useAwardPoints', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls awardPoints with customerId, points, and reason', async () => {
    mockCrmApi.awardPoints.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useAwardPoints(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ customerId: 1, points: 100, reason: 'Purchase' });
    });

    expect(mockCrmApi.awardPoints).toHaveBeenCalledWith(1, 100, 'Purchase');
  });
});

describe('useRedeemPoints', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls redeemPoints with customerId, points, and reason', async () => {
    mockCrmApi.redeemPoints.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useRedeemPoints(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ customerId: 1, points: 50, reason: 'Reward' });
    });

    expect(mockCrmApi.redeemPoints).toHaveBeenCalledWith(1, 50, 'Reward');
  });
});
