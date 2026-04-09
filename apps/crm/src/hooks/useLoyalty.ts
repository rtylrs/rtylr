import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, LoyaltyConfig } from '../services/crmApi';

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
    mutationFn: ({ customerId, points }: { customerId: number; points: number }) =>
      crmApi.awardPoints(customerId, points),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['loyalty-transactions'] });
      qc.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

export function useRedeemPoints() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ customerId, points }: { customerId: number; points: number }) =>
      crmApi.redeemPoints(customerId, points),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['loyalty-transactions'] });
      qc.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}
