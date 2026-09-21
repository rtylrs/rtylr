import React from 'react';
import { act, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const crmApiMock = vi.hoisted(() => ({
  updateLoyaltyConfig: vi.fn(async () => ({})),
  updateLoyaltyTier: vi.fn(async () => ({})),
  createLoyaltyTier: vi.fn(async () => ({})),
}));

vi.mock('../../../services/crmApi', () => ({ crmApi: crmApiMock }));

vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<object>();
  return {
    ...actual,
    unwrap: vi.fn(async (value: unknown) => value),
  };
});

import { useUpdateLoyaltyConfig, useUpsertTier } from '../useLoyalty';
import { loyaltyConfigSchema, loyaltyTierSchema } from '../../../schemas/loyalty';

let queryClient: QueryClient;

const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

beforeEach(() => {
  vi.clearAllMocks();
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
});

describe('useUpsertTier (B45-R03)', () => {
  it('sends the discount the tier form edited, not the stale cached copy', async () => {
    // loyaltyTierSchema emits both multiplier and discountPct, so a parsed tier
    // carries its old discount in discountPct while the edit form writes
    // multiplier. The request must prefer the field the form actually edited.
    const cachedTier = loyaltyTierSchema.parse({
      id: 1,
      name: 'Gold',
      minPoints: 1000,
      discountPct: 10,
      perks: [],
      sortOrder: 2,
    });
    queryClient.setQueryData(['crm-loyalty-tiers'], [cachedTier]);

    const { result } = renderHook(() => useUpsertTier(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 1, data: { threshold: 1000, multiplier: 25 } });
    });

    expect(crmApiMock.updateLoyaltyTier).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ discountPct: 25 })
    );
  });
});

describe('useUpdateLoyaltyConfig (B60-R06)', () => {
  it('refuses to write a replacement config before the current one has loaded', async () => {
    const { result } = renderHook(() => useUpdateLoyaltyConfig(), { wrapper });

    await act(async () => {
      await expect(result.current.mutateAsync({ programEnabled: true })).rejects.toThrow(
        'loyalty config has not loaded'
      );
    });

    // Without the guard this PUT carried the hard-coded defaults and disabled
    // the program.
    expect(crmApiMock.updateLoyaltyConfig).not.toHaveBeenCalled();
  });

  it('merges a partial save over the loaded config instead of the defaults', async () => {
    const current = loyaltyConfigSchema.parse({
      isEnabled: false,
      pointsPerCurrencyUnit: 10,
      currencyPerPoint: 0.01,
      bronzeMin: '250',
      silverMin: '750',
      goldMin: '2000',
      platinumMin: '6000',
      bronzeDiscountPct: 0,
      silverDiscountPct: 5,
      goldDiscountPct: 10,
      platinumDiscountPct: 15,
      pointsExpiryDays: '365',
    });
    queryClient.setQueryData(['crm-loyalty-config'], current);

    const { result } = renderHook(() => useUpdateLoyaltyConfig(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ programEnabled: true });
    });

    expect(crmApiMock.updateLoyaltyConfig).toHaveBeenCalledWith(
      expect.objectContaining({
        isEnabled: true,
        bronzeMin: '250',
        silverMin: '750',
        pointsExpiryDays: '365',
      })
    );
  });
});
