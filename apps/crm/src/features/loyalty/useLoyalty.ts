import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi, LoyaltyConfig, LoyaltyTier } from '../../shared/services/crmApi';
import {
  loyaltyConfigSchema,
  loyaltyTransactionSchema,
  loyaltyMemberSchema,
  loyaltyRewardSchema,
  loyaltyTierSchema,
} from '../../shared/schemas/loyalty';

// Many loyalty endpoints return `{ <key>: [...] }` envelopes from the backend.
// Accept either a raw array or an enveloped response so the schema parse
// never sees undefined, which would throw.
function unwrapList<T>(res: unknown, ...keys: string[]): T[] {
  if (Array.isArray(res)) return res as T[];
  if (res && typeof res === 'object') {
    for (const key of keys) {
      const value = (res as Record<string, unknown>)[key];
      if (Array.isArray(value)) return value as T[];
    }
  }
  return [];
}

export function useLoyaltyTransactions() {
  return useQuery({
    queryKey: ['loyalty-transactions'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getLoyaltyTransactions());
      return z.array(loyaltyTransactionSchema).parse(unwrapList(res, 'transactions'));
    },
  });
}

export function useLoyaltyConfig() {
  return useQuery({
    queryKey: ['loyalty-config'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getLoyaltyConfig());
      return loyaltyConfigSchema.parse(res);
    },
  });
}

export function useLoyaltyMembers() {
  return useQuery({
    queryKey: ['loyalty-members'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getLoyaltyMembers());
      return z.array(loyaltyMemberSchema).parse(unwrapList(res, 'members'));
    },
  });
}

export function useLoyaltyRewards() {
  return useQuery({
    queryKey: ['loyalty-rewards'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getLoyaltyRewards());
      return z.array(loyaltyRewardSchema).parse(unwrapList(res, 'rewards'));
    },
  });
}

export function useLoyaltyTiers() {
  return useQuery({
    queryKey: ['loyalty-tiers'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getLoyaltyTiers());
      return z.array(loyaltyTierSchema).parse(unwrapList(res, 'tiers'));
    },
  });
}

export function useUpdateLoyaltyTier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<LoyaltyTier> }) =>
      unwrap(crmApi.updateLoyaltyTier(id, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['loyalty-tiers'] }),
  });
}

export function useUpdateLoyaltyConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (config: Partial<LoyaltyConfig>) => unwrap(crmApi.updateLoyaltyConfig(config)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['loyalty-config'] }),
  });
}

export function useAwardPoints() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      customerId,
      points,
      reason,
    }: {
      customerId: number;
      points: number;
      reason?: string;
    }) => unwrap(crmApi.awardPoints(customerId, points, reason)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['loyalty-transactions'] });
      qc.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

export function useRedeemPoints() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      customerId,
      points,
      reason,
    }: {
      customerId: number;
      points: number;
      reason?: string;
    }) => unwrap(crmApi.redeemPoints(customerId, points, reason)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['loyalty-transactions'] });
      qc.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}
