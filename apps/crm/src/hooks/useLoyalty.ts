import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, LoyaltyConfig, LoyaltyTier } from '../services/crmApi';

export function useLoyaltyTransactions() {
  return useQuery({
    queryKey: ['loyalty-transactions'],
    queryFn: () => crmApi.getLoyaltyTransactions(),
  });
}

export function useLoyaltyConfig() {
  return useQuery({
    queryKey: ['loyalty-config'],
    queryFn: () => crmApi.getLoyaltyConfig(),
  });
}

export function useLoyaltyMembers() {
  return useQuery({
    queryKey: ['loyalty-members'],
    queryFn: () => crmApi.getLoyaltyMembers(),
  });
}

export function useLoyaltyRewards() {
  return useQuery({
    queryKey: ['loyalty-rewards'],
    queryFn: () => crmApi.getLoyaltyRewards(),
  });
}

export function useLoyaltyTiers() {
  return useQuery({
    queryKey: ['loyalty-tiers'],
    queryFn: () => crmApi.getLoyaltyTiers(),
  });
}

export function useUpdateLoyaltyTier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<LoyaltyTier> }) =>
      crmApi.updateLoyaltyTier(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['loyalty-tiers'] }),
  });
}

export function useUpdateLoyaltyConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (config: Partial<LoyaltyConfig>) => crmApi.updateLoyaltyConfig(config),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['loyalty-config'] }),
  });
}

export function useAwardPoints() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ customerId, points, reason }: { customerId: number; points: number; reason?: string }) =>
      crmApi.awardPoints(customerId, points, reason),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['loyalty-transactions'] });
      qc.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

export function useRedeemPoints() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ customerId, points, reason }: { customerId: number; points: number; reason?: string }) =>
      crmApi.redeemPoints(customerId, points, reason),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['loyalty-transactions'] });
      qc.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}
