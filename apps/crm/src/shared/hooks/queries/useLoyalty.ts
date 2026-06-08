import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { unwrapList } from '../../lib/unwrapList';
import {
  loyaltyConfigSchema,
  loyaltyMemberSchema,
  loyaltyRewardSchema,
  loyaltyTierSchema,
  loyaltyTransactionSchema,
} from '../../schemas/loyalty';
import type {
  LoyaltyConfig,
  LoyaltyMember,
  LoyaltyReward,
  LoyaltyTier,
  LoyaltyTransaction,
} from '../../schemas/loyalty';

// ── Types ────────────────────────────────────────────────────────────────────

interface UpsertTierPayload {
  id?: number;
  data: Partial<LoyaltyTier>;
}

interface UpsertRewardPayload {
  id?: number;
  data: Partial<LoyaltyReward>;
}

interface AwardRedeemPayload {
  customerId: number;
  points: number;
  reason?: string;
}

// ── Queries ──────────────────────────────────────────────────────────────────

export const useLoyaltyConfig = (): UseQueryResult<LoyaltyConfig> =>
  useQuery<LoyaltyConfig>({
    queryKey: ['crm-loyalty-config'],
    queryFn: async (): Promise<LoyaltyConfig> => {
      const res = await unwrap(safe.get<unknown>('/crm/loyalty/config'));
      return loyaltyConfigSchema.parse(res);
    },
  });

export const useLoyaltyTiers = (): UseQueryResult<LoyaltyTier[]> =>
  useQuery<LoyaltyTier[]>({
    queryKey: ['crm-loyalty-tiers'],
    queryFn: async (): Promise<LoyaltyTier[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/loyalty/tiers'));
      return z.array(loyaltyTierSchema).parse(unwrapList(res, 'tiers'));
    },
  });

export const useLoyaltyMembers = (): UseQueryResult<LoyaltyMember[]> =>
  useQuery<LoyaltyMember[]>({
    queryKey: ['crm-loyalty-members'],
    queryFn: async (): Promise<LoyaltyMember[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/loyalty/members'));
      return z.array(loyaltyMemberSchema).parse(unwrapList(res, 'members'));
    },
  });

export const useLoyaltyRewards = (): UseQueryResult<LoyaltyReward[]> =>
  useQuery<LoyaltyReward[]>({
    queryKey: ['crm-loyalty-rewards'],
    queryFn: async (): Promise<LoyaltyReward[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/loyalty/rewards'));
      return z.array(loyaltyRewardSchema).parse(unwrapList(res, 'rewards'));
    },
  });

export const useLoyaltyTransactions = (): UseQueryResult<LoyaltyTransaction[]> =>
  useQuery<LoyaltyTransaction[]>({
    queryKey: ['crm-loyalty-transactions'],
    queryFn: async (): Promise<LoyaltyTransaction[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/loyalty/transactions'));
      return z.array(loyaltyTransactionSchema).parse(unwrapList(res, 'transactions'));
    },
  });

// ── Mutations ────────────────────────────────────────────────────────────────

export const useUpdateLoyaltyConfig = (): UseMutationResult<
  LoyaltyConfig,
  unknown,
  Partial<LoyaltyConfig>
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<LoyaltyConfig, unknown, Partial<LoyaltyConfig>>({
    mutationFn: async (config: Partial<LoyaltyConfig>): Promise<LoyaltyConfig> => {
      const res = await unwrap(safe.put<LoyaltyConfig>('/crm/loyalty/config', config));
      return res as LoyaltyConfig;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-config'] });
      toast.success(t('crm.loyalty.configUpdated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpsertTier = (): UseMutationResult<unknown, unknown, UpsertTierPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, UpsertTierPayload>({
    mutationFn: (payload: UpsertTierPayload): Promise<unknown> => {
      if (payload.id) {
        return unwrap(safe.put('/crm/loyalty/tiers', { id: payload.id, ...payload.data }));
      }
      return unwrap(safe.post('/crm/loyalty/tiers', payload.data));
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-tiers'] });
      toast.success(t('crm.loyalty.tierSaved'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpsertReward = (): UseMutationResult<unknown, unknown, UpsertRewardPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, UpsertRewardPayload>({
    mutationFn: (payload: UpsertRewardPayload): Promise<unknown> =>
      unwrap(
        safe.post(
          '/crm/loyalty/rewards/upsert',
          payload.id ? { id: payload.id, ...payload.data } : payload.data
        )
      ),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-rewards'] });
      toast.success(t('crm.loyalty.rewardSaved'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useAwardPoints = (): UseMutationResult<unknown, unknown, AwardRedeemPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, AwardRedeemPayload>({
    mutationFn: (payload: AwardRedeemPayload): Promise<unknown> =>
      unwrap(
        safe.post('/crm/loyalty/award', {
          customerId: payload.customerId,
          points: payload.points,
          description: payload.reason,
        })
      ),
    onSuccess: (_data: unknown, variables: AwardRedeemPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-transactions'] });
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      void qc.invalidateQueries({ queryKey: ['crm-customer', variables.customerId] });
      toast.success(t('crm.loyalty.pointsAwarded'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useRedeemPoints = (): UseMutationResult<unknown, unknown, AwardRedeemPayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, AwardRedeemPayload>({
    mutationFn: (payload: AwardRedeemPayload): Promise<unknown> =>
      unwrap(
        safe.post('/crm/loyalty/redeem', {
          customerId: payload.customerId,
          points: payload.points,
          description: payload.reason,
        })
      ),
    onSuccess: (_data: unknown, variables: AwardRedeemPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-loyalty-transactions'] });
      void qc.invalidateQueries({ queryKey: ['crm-customers'] });
      void qc.invalidateQueries({ queryKey: ['crm-customer', variables.customerId] });
      toast.success(t('crm.loyalty.pointsRedeemed'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type {
  LoyaltyConfig,
  LoyaltyMember,
  LoyaltyReward,
  LoyaltyTier,
  LoyaltyTransaction,
} from '../../schemas/loyalty';
